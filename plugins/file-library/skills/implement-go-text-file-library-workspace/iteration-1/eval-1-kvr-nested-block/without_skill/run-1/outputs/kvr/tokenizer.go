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

// UnterminatedStringError is returned when a quoted string is closed by a
// literal newline or end-of-file before its closing quote is seen.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned for an unrecognised backslash escape inside
// a quoted string.
type InvalidEscapeError struct {
	Pos  Pos
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape \\%c at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position.
type tokenizer struct {
	r   *bufio.Reader
	pos Pos
}

// next advances the cursor by one rune and updates pos.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	if r == '\n' {
		t.pos.Line++
		t.pos.Column = 1
	} else {
		t.pos.Column++
	}
	return r, nil
}

// backup rewinds the last rune read by next.
func (t *tokenizer) backup() {
	_ = t.r.UnreadRune()
	t.pos.Column--
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

func isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func isIdentCont(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// startPos returns the Pos of the most recently consumed rune (one column to
// the left of the cursor). Newlines complicate this; the tokenizer instead
// tracks the start of the current token explicitly when needed.
func (t *tokenizer) currentPos() Pos {
	// Column points to the next-rune-to-be-read column; the rune just consumed
	// sits one column earlier on the same line.
	return Pos{Line: t.pos.Line, Column: t.pos.Column - 1}
}

// tokenize is the top-level dispatch action.
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

	pos := t.currentPos()

	switch {
	case r == '#':
		return tokenizeComment(pos)
	case r == '"':
		return tokenizeString(pos)
	case r == '=' || r == '{' || r == '}' || r == ';':
		if !yield(Token{Pos: pos, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	case isIdentStart(r):
		var b strings.Builder
		b.WriteRune(r)
		return tokenizeIdent(pos, &b)
	case isDigit(r):
		var b strings.Builder
		b.WriteRune(r)
		return tokenizeNumber(pos, &b)
	default:
		yield(Token{}, &UnexpectedCharacterError{Pos: pos, Char: r})
		return nil
	}
}

func tokenizeIdent(start Pos, b *strings.Builder) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenIdentifier, Value: b.String()}, nil) {
					return nil
				}
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isIdentCont(r) {
				t.backup()
				if !yield(Token{Pos: start, Type: TokenIdentifier, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			b.WriteRune(r)
		}
	}
}

func tokenizeNumber(start Pos, b *strings.Builder) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenNumber, Value: b.String()}, nil) {
					return nil
				}
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isDigit(r) {
				t.backup()
				if !yield(Token{Pos: start, Type: TokenNumber, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			b.WriteRune(r)
		}
	}
}

func tokenizeComment(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		// Skip leading horizontal whitespace.
		leading := true
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenComment, Value: b.String()}, nil) {
					return nil
				}
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if r == '\n' {
				if !yield(Token{Pos: start, Type: TokenComment, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if leading && (r == ' ' || r == '\t') {
				continue
			}
			leading = false
			b.WriteRune(r)
		}
	}
}

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
					yield(Token{}, &InvalidEscapeError{Pos: t.currentPos(), Char: esc})
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
