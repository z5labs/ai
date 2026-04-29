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
// before EOF or before a literal newline. Pos is the opening-quote position.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a backslash escape inside a string
// literal is not one of the recognised forms. Pos is the position of the
// backslash; Char is the rune that followed it.
type InvalidEscapeError struct {
	Pos  Pos
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape sequence \\%c at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position. pos is the position of
// the *next* rune to be read — i.e. before calling next(), t.pos is where
// the upcoming rune sits.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos // pos snapshot before the most recent next(); restored by backup()
}

// next advances the cursor by one rune and updates pos to the position
// following the returned rune.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prevPos = t.pos
	if r == '\n' {
		t.pos.Line++
		t.pos.Column = 1
	} else {
		t.pos.Column++
	}
	return r, nil
}

// backup rewinds the last rune read by next, restoring pos to its value
// before that read.
func (t *tokenizer) backup() {
	_ = t.r.UnreadRune()
	t.pos = t.prevPos
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldErr is the standard error-and-stop ending used by every error path.
func yieldErr(err error) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		yield(Token{}, err)
		return nil
	}
}

// tokenize is the top-level dispatch action. It captures the position of the
// next rune *before* reading it, so specialised actions get the correct start.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	startPos := t.pos
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return yieldErr(err)
	}
	switch {
	case r == '\n' || unicode.IsSpace(r):
		return tokenize
	case r == '#':
		return tokenizeComment(startPos)
	case r == '"':
		return tokenizeString(startPos)
	case r == '=' || r == '{' || r == '}' || r == ';':
		return tokenizeSymbol(startPos, r)
	case unicode.IsDigit(r):
		t.backup()
		return tokenizeNumber(startPos)
	case isIdentStart(r):
		t.backup()
		return tokenizeIdentifier(startPos)
	}
	return yieldErr(&UnexpectedCharacterError{Pos: startPos, Char: r})
}

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

func tokenizeSymbol(start Pos, r rune) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		if !yield(Token{Pos: start, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	}
}

func tokenizeIdentifier(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		for {
			r, err := t.next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return yieldErr(err)
			}
			if !isIdentPart(r) {
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

func tokenizeNumber(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		for {
			r, err := t.next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return yieldErr(err)
			}
			if r < '0' || r > '9' {
				t.backup()
				break
			}
			b.WriteRune(r)
		}
		if !yield(Token{Pos: start, Type: TokenNumber, Value: b.String()}, nil) {
			return nil
		}
		return tokenize
	}
}

func tokenizeString(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		for {
			escapePos := t.pos
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
				if !yield(Token{Pos: start, Type: TokenString, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if r == '\\' {
				esc, err := t.next()
				if err == io.EOF {
					return yieldErr(&UnterminatedStringError{Pos: start})
				}
				if err != nil {
					return yieldErr(err)
				}
				switch esc {
				case '\\':
					b.WriteByte('\\')
				case '"':
					b.WriteByte('"')
				case 'n':
					b.WriteByte('\n')
				case 't':
					b.WriteByte('\t')
				default:
					return yieldErr(&InvalidEscapeError{Pos: escapePos, Char: esc})
				}
				continue
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
				break
			}
			if err != nil {
				return yieldErr(err)
			}
			if r == '\n' {
				break
			}
			if leading && (r == ' ' || r == '\t') {
				continue
			}
			leading = false
			b.WriteRune(r)
		}
		if !yield(Token{Pos: start, Type: TokenComment, Value: b.String()}, nil) {
			return nil
		}
		return tokenize
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
