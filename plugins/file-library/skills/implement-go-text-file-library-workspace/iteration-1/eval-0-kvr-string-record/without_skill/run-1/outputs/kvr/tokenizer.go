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

// UnterminatedStringError is returned when a quoted string literal reaches a
// real newline rune or end-of-file before its closing quote. The Pos field
// carries the position of the opening quote.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when the tokenizer sees a backslash inside a
// quoted string followed by a rune that is not a recognised escape.
type InvalidEscapeError struct {
	Pos  Pos
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape %q at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos // position before the most recent next(); used by backup.
	hasPrev bool
}

// next advances the cursor by one rune and updates pos.
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

// backup rewinds the last rune read by next, restoring the prior position.
func (t *tokenizer) backup() {
	_ = t.r.UnreadRune()
	if t.hasPrev {
		t.pos = t.prevPos
		t.hasPrev = false
	}
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// isIdentStart reports whether r can start an identifier.
func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isIdentCont reports whether r can continue an identifier.
func isIdentCont(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

// tokenize is the top-level dispatch action. It consumes one rune to decide
// which specialised action runs next; specialised actions are responsible for
// emitting their token and returning back here.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		yield(Token{}, err)
		return nil
	}

	// Skip insignificant whitespace.
	if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
		return tokenize
	}

	// Identifier (letter or underscore start).
	if isIdentStart(r) {
		// Capture position of the first rune; t.pos is now one past it.
		start := Pos{Line: t.pos.Line, Column: t.pos.Column - 1}
		t.backup()
		return tokenizeIdentifier(start)
	}

	// String literal — opening quote is consumed; start is its position.
	if r == '"' {
		start := Pos{Line: t.pos.Line, Column: t.pos.Column - 1}
		return tokenizeString(start)
	}

	// Single-character symbol.
	if r == '=' {
		pos := Pos{Line: t.pos.Line, Column: t.pos.Column - 1}
		if !yield(Token{Pos: pos, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	}

	pos := Pos{Line: t.pos.Line, Column: t.pos.Column - 1}
	yield(Token{}, &UnexpectedCharacterError{Pos: pos, Char: r})
	return nil
}

// tokenizeIdentifier returns an action that reads the rest of the identifier
// (the first rune was un-read by the dispatcher) and yields it with the
// captured start position.
func tokenizeIdentifier(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		for {
			r, err := t.next()
			if err == io.EOF {
				break
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isIdentCont(r) {
				t.backup()
				break
			}
			b.WriteRune(r)
		}
		if !yield(Token{Pos: start, Type: TokenIdentifier, Value: b.String()}, nil) {
			return nil
		}
		return tokenize
	}
}

// tokenizeString reads characters up to the closing double quote, decoding
// recognised backslash escapes. The opening quote has already been consumed;
// start carries its position.
func tokenizeString(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
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
				if !yield(Token{Pos: start, Type: TokenString, Value: b.String()}, nil) {
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
					b.WriteRune('\\')
				case '"':
					b.WriteRune('"')
				case 'n':
					b.WriteRune('\n')
				case 't':
					b.WriteRune('\t')
				default:
					escPos := Pos{Line: t.pos.Line, Column: t.pos.Column - 1}
					yield(Token{}, &InvalidEscapeError{Pos: escPos, Char: esc})
					return nil
				}
				continue
			}
			b.WriteRune(r)
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
