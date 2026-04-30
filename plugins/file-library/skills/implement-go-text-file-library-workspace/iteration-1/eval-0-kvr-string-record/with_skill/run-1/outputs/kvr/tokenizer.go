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

// UnterminatedStringError is returned when a string literal is not closed
// before end-of-file or before a literal newline rune.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a backslash inside a string literal is
// followed by a rune that is not one of the recognised escapes (\\, \", \n,
// \t).
type InvalidEscapeError struct {
	Pos  Pos
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape sequence \\%c at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position.
//
// pos always points at the next rune to be read: before next() consumes a
// rune, t.pos is that rune's position; after next() returns, t.pos has been
// advanced to the rune that follows. Specialised actions therefore capture
// `start := t.pos` BEFORE calling next() to remember where a token began.
type tokenizer struct {
	r   *bufio.Reader
	pos Pos
	// lastWasNewline records whether the previous next() returned '\n' so
	// backup() can restore the column on the prior line. Only one rune of
	// lookback is supported (matching bufio.Reader.UnreadRune).
	lastWasNewline bool
	prevColumn     int
}

// next advances the cursor by one rune. The returned rune was at the position
// recorded in t.pos at call time; on return, t.pos has been advanced.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.lastWasNewline = false
	if r == '\n' {
		t.prevColumn = t.pos.Column
		t.pos.Line++
		t.pos.Column = 1
		t.lastWasNewline = true
	} else {
		t.prevColumn = t.pos.Column
		t.pos.Column++
	}
	return r, nil
}

// backup rewinds the last rune read by next.
func (t *tokenizer) backup() {
	if t.r.UnreadRune() != nil {
		return
	}
	if t.lastWasNewline {
		t.pos.Line--
		t.pos.Column = t.prevColumn
		t.lastWasNewline = false
	} else {
		t.pos.Column = t.prevColumn
	}
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldErr emits err and ends the iterator.
func yieldErr(err error) tokenizerAction {
	return func(_ *tokenizer, yield func(Token, error) bool) tokenizerAction {
		yield(Token{}, err)
		return nil
	}
}

// tokenize is the top-level dispatch action. It peeks one rune, dispatches by
// kind, and returns a specialised action (which itself returns `tokenize`
// after yielding so dispatch resumes).
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	start := t.pos
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return yieldErr(err)
	}
	switch {
	case unicode.IsSpace(r):
		return tokenize
	case isIdentifierStart(r):
		return tokenizeIdentifier(start, r)
	case r == '=':
		if !yield(Token{Pos: start, Type: TokenSymbol, Value: "="}, nil) {
			return nil
		}
		return tokenize
	case r == '{', r == '}', r == ';':
		if !yield(Token{Pos: start, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	case r == '"':
		return tokenizeString(start)
	}
	return yieldErr(&UnexpectedCharacterError{Pos: start, Char: r})
}

func isIdentifierStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentifierContinue(r rune) bool {
	return isIdentifierStart(r) || (r >= '0' && r <= '9')
}

// tokenizeIdentifier returns a closure that consumes the rest of an
// identifier and yields a TokenIdentifier with `start` as its position.
func tokenizeIdentifier(start Pos, first rune) tokenizerAction {
	var sb strings.Builder
	sb.WriteRune(first)
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenIdentifier, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if err != nil {
				return yieldErr(err)
			}
			if !isIdentifierContinue(r) {
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

// tokenizeString returns a closure that reads characters up to the closing
// quote and yields a TokenString with `start` set to the opening-quote
// position. Decoded escape sequences are supported per the spec.
func tokenizeString(start Pos) tokenizerAction {
	var sb strings.Builder
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		for {
			r, err := t.next()
			if err == io.EOF {
				return yieldErr(&UnterminatedStringError{Pos: start})
			}
			if err != nil {
				return yieldErr(err)
			}
			if r == '\n' {
				return yieldErr(&UnterminatedStringError{Pos: start})
			}
			if r == '"' {
				if !yield(Token{Pos: start, Type: TokenString, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if r == '\\' {
				escPos := Pos{Line: t.pos.Line, Column: t.pos.Column - 1}
				esc, eerr := t.next()
				if eerr == io.EOF {
					return yieldErr(&UnterminatedStringError{Pos: start})
				}
				if eerr != nil {
					return yieldErr(eerr)
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
					return yieldErr(&InvalidEscapeError{Pos: escPos, Char: esc})
				}
				continue
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
