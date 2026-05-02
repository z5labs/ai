package kvr

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"strings"
	"unicode"
)

// Pos identifies a 1-based line/column position in the source.
type Pos struct {
	Line   int
	Column int
}

// TokenType identifies the kind of a Token.
type TokenType int

const (
	TokenInvalid TokenType = iota
	TokenIdentifier
	TokenSymbol
	TokenString
	TokenNumber
	TokenComment
)

// String returns a human-readable name for a TokenType. Implementer extends
// this when new token types are added so test failures print readable names.
func (t TokenType) String() string {
	switch t {
	case TokenIdentifier:
		return "IDENT"
	case TokenSymbol:
		return "SYMBOL"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenComment:
		return "COMMENT"
	default:
		return fmt.Sprintf("TokenType(%d)", int(t))
	}
}

// Token is one lexical element produced by the tokenizer.
type Token struct {
	Pos   Pos
	Type  TokenType
	Value string
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%d:%d", t.Type, t.Value, t.Pos.Line, t.Pos.Column)
}

// UnexpectedCharacterError is returned when the tokenizer encounters a rune
// that no action wanted.
type UnexpectedCharacterError struct {
	Pos  Pos
	Char rune
}

func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("unexpected character %q at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// UnterminatedStringError is returned when a string literal is closed by
// end-of-input or a literal newline rather than a closing quote.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a backslash inside a string is followed
// by a rune that isn't a recognised escape character.
type InvalidEscapeError struct {
	Pos  Pos
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape \\%c at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position. prevPos snapshots the
// position before the most recent next() so backup() can restore it
// (including across newline boundaries) — never reconstruct via column
// arithmetic, since that underflows when the previous next() reset Column
// to 1 after consuming '\n'.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos
	hasPrev bool
}

// next advances the cursor by one rune and updates pos. It snapshots pos
// into prevPos before mutating so backup() can restore it.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prevPos = t.pos
	t.hasPrev = true
	if r == '\n' {
		t.pos.Line++
		t.pos.Column = 1
	} else {
		t.pos.Column++
	}
	return r, nil
}

// backup rewinds the last rune read by next, restoring pos. backup may only
// be called once per next.
func (t *tokenizer) backup() {
	if err := t.r.UnreadRune(); err != nil {
		return
	}
	if t.hasPrev {
		t.pos = t.prevPos
		t.hasPrev = false
	}
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// tokenize is the top-level dispatch action. The implementer extends the switch
// to recognise new token types and returns the appropriate specialised action.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		yield(Token{}, err)
		return nil
	}

	// Skip whitespace (excluding '\n' which is just whitespace too — KVR has
	// no statement terminators outside blocks).
	if unicode.IsSpace(r) {
		return tokenize
	}

	// Comments
	if r == '#' {
		startPos := t.prevPos
		return tokenizeComment(startPos)
	}

	// Symbols
	switch r {
	case '=', '{', '}', ';':
		pos := t.prevPos
		if !yield(Token{Pos: pos, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	}

	// String literal
	if r == '"' {
		startPos := t.prevPos
		return tokenizeString(startPos)
	}

	// Identifier
	if isIdentStart(r) {
		startPos := t.prevPos
		var sb strings.Builder
		sb.WriteRune(r)
		return tokenizeIdentifier(startPos, &sb)
	}

	// Number
	if r >= '0' && r <= '9' {
		startPos := t.prevPos
		var sb strings.Builder
		sb.WriteRune(r)
		return tokenizeNumber(startPos, &sb)
	}

	yield(Token{}, &UnexpectedCharacterError{Pos: t.prevPos, Char: r})
	return nil
}

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

func tokenizeIdentifier(start Pos, sb *strings.Builder) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		r, err := t.next()
		if err == io.EOF {
			if !yield(Token{Pos: start, Type: TokenIdentifier, Value: sb.String()}, nil) {
				return nil
			}
			return nil
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if isIdentPart(r) {
			sb.WriteRune(r)
			return tokenizeIdentifier(start, sb)
		}
		t.backup()
		if !yield(Token{Pos: start, Type: TokenIdentifier, Value: sb.String()}, nil) {
			return nil
		}
		return tokenize
	}
}

func tokenizeNumber(start Pos, sb *strings.Builder) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		r, err := t.next()
		if err == io.EOF {
			if !yield(Token{Pos: start, Type: TokenNumber, Value: sb.String()}, nil) {
				return nil
			}
			return nil
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
			return tokenizeNumber(start, sb)
		}
		t.backup()
		if !yield(Token{Pos: start, Type: TokenNumber, Value: sb.String()}, nil) {
			return nil
		}
		return tokenize
	}
}

func tokenizeString(start Pos) tokenizerAction {
	var sb strings.Builder
	var inner tokenizerAction
	inner = func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		r, err := t.next()
		if err == io.EOF {
			yield(Token{}, &UnterminatedStringError{Pos: start})
			return nil
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if r == '\n' {
			yield(Token{}, &UnterminatedStringError{Pos: start})
			return nil
		}
		if r == '"' {
			if !yield(Token{Pos: start, Type: TokenString, Value: sb.String()}, nil) {
				return nil
			}
			return tokenize
		}
		if r == '\\' {
			esc, err := t.next()
			if err == io.EOF {
				yield(Token{}, &UnterminatedStringError{Pos: start})
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			switch esc {
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			default:
				yield(Token{}, &InvalidEscapeError{Pos: t.prevPos, Char: esc})
				return nil
			}
			return inner
		}
		sb.WriteRune(r)
		return inner
	}
	return inner
}

func tokenizeComment(start Pos) tokenizerAction {
	var sb strings.Builder
	leadingDone := false
	var inner tokenizerAction
	inner = func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		r, err := t.next()
		if err == io.EOF {
			if !yield(Token{Pos: start, Type: TokenComment, Value: sb.String()}, nil) {
				return nil
			}
			return nil
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if r == '\n' {
			if !yield(Token{Pos: start, Type: TokenComment, Value: sb.String()}, nil) {
				return nil
			}
			return tokenize
		}
		if !leadingDone {
			if r == ' ' || r == '\t' {
				return inner
			}
			leadingDone = true
		}
		sb.WriteRune(r)
		return inner
	}
	return inner
}

// Tokenize streams tokens from r as an iter.Seq2[Token, error].
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		t := &tokenizer{r: bufio.NewReader(r), pos: Pos{Line: 1, Column: 1}}
		for action := tokenize; action != nil; {
			action = action(t, yield)
		}
	}
}
