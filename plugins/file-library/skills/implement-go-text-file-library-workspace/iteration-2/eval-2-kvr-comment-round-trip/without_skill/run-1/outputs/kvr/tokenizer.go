package kvr

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"strings"
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

// UnterminatedStringError is returned when a string literal has no closing
// quote before end-of-line or end-of-input.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
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

func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHorizontalSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

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
	// Skip whitespace (including newlines).
	if isHorizontalSpace(r) || r == '\n' {
		return tokenize
	}
	// Save the start position of this token. After next() pos has already
	// advanced; the start of the token is prevPos.
	start := t.prevPos
	switch {
	case r == '#':
		return tokenizeComment(start)
	case r == '"':
		return tokenizeString(start)
	case isIdentStart(r):
		return tokenizeIdent(start, r)
	case isDigit(r):
		return tokenizeNumber(start, r)
	case r == '=' || r == '{' || r == '}' || r == ';':
		if !yield(Token{Pos: start, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	}
	yield(Token{}, &UnexpectedCharacterError{Pos: start, Char: r})
	return nil
}

func tokenizeComment(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		// Skip leading horizontal whitespace after '#'.
		leading := true
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenComment, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
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
			if leading && isHorizontalSpace(r) {
				continue
			}
			leading = false
			sb.WriteRune(r)
		}
	}
}

func tokenizeString(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		for {
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
					yield(Token{}, &UnexpectedCharacterError{Pos: t.prevPos, Char: esc})
					return nil
				}
				continue
			}
			sb.WriteRune(r)
		}
	}
}

func tokenizeIdent(start Pos, first rune) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		sb.WriteRune(first)
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenIdentifier, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isIdentPart(r) {
				t.backup()
				if !yield(Token{Pos: start, Type: TokenIdentifier, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
			}
			sb.WriteRune(r)
		}
	}
}

func tokenizeNumber(start Pos, first rune) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		sb.WriteRune(first)
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenNumber, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isDigit(r) {
				t.backup()
				if !yield(Token{Pos: start, Type: TokenNumber, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
			}
			sb.WriteRune(r)
		}
	}
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
