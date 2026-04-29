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
// before a newline or end-of-file.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a backslash escape inside a string is
// not one of the recognised forms.
type InvalidEscapeError struct {
	Pos  Pos
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape %q at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position.
type tokenizer struct {
	r   *bufio.Reader
	pos Pos
	// prev tracks the position before the most recent next() call so backup()
	// can restore line/column accurately even when crossing a newline.
	prev Pos
}

// next advances the cursor by one rune and updates pos.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prev = t.pos
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
	t.pos = t.prev
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// tokenize is the top-level dispatch action. It reads one rune and dispatches
// to the matching specialised action.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		yield(Token{}, err)
		return nil
	}

	switch {
	case r == '\n' || unicode.IsSpace(r):
		return tokenize
	case r == '#':
		return tokenizeComment(t.prev)
	case r == '"':
		return tokenizeString(t.prev)
	case isIdentStart(r):
		return tokenizeIdentifier(t.prev, string(r))
	case isDigit(r):
		return tokenizeNumber(t.prev, string(r))
	case isSymbol(r):
		if !yield(Token{Pos: t.prev, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	default:
		yield(Token{}, &UnexpectedCharacterError{Pos: t.prev, Char: r})
		return nil
	}
}

// tokenizeComment consumes from just after the '#' to end-of-line (or EOF).
// Leading horizontal whitespace inside the comment is stripped from Value.
func tokenizeComment(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: start, Type: TokenComment, Value: strings.TrimLeft(sb.String(), " \t")}, nil) {
					return nil
				}
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if r == '\n' {
				if !yield(Token{Pos: start, Type: TokenComment, Value: strings.TrimLeft(sb.String(), " \t")}, nil) {
					return nil
				}
				return tokenize
			}
			sb.WriteRune(r)
		}
	}
}

// tokenizeString consumes a double-quoted string. start is the position of the
// opening quote. Recognised escapes: \\ \" \n \t.
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
			switch r {
			case '"':
				if !yield(Token{Pos: start, Type: TokenString, Value: sb.String()}, nil) {
					return nil
				}
				return tokenize
			case '\n':
				yield(Token{}, &UnterminatedStringError{Pos: start})
				return nil
			case '\\':
				escPos := t.prev
				er, err := t.next()
				if err == io.EOF {
					yield(Token{}, &UnterminatedStringError{Pos: start})
					return nil
				}
				if err != nil {
					yield(Token{}, err)
					return nil
				}
				switch er {
				case '\\':
					sb.WriteRune('\\')
				case '"':
					sb.WriteRune('"')
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				default:
					yield(Token{}, &InvalidEscapeError{Pos: escPos, Char: er})
					return nil
				}
			default:
				sb.WriteRune(r)
			}
		}
	}
}

// tokenizeIdentifier accumulates an identifier given the first rune already
// read.
func tokenizeIdentifier(start Pos, first string) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		sb.WriteString(first)
		for {
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
			if !isIdentCont(r) {
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

// tokenizeNumber accumulates one or more digits.
func tokenizeNumber(start Pos, first string) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		sb.WriteString(first)
		for {
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

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentCont(r rune) bool {
	return isIdentStart(r) || isDigit(r)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isSymbol(r rune) bool {
	switch r {
	case '=', '{', '}', ';':
		return true
	}
	return false
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
