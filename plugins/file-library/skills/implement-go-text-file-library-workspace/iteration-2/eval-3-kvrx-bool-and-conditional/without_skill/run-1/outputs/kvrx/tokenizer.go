package kvrx

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
	TokenNewline
)

// String returns a human-readable name for a TokenType.
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
	case TokenNewline:
		return "NEWLINE"
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

// UnterminatedStringError is returned when a `"` opens a string but no
// matching closing `"` is found before end-of-input or end-of-line.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos
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

// backup rewinds the last rune read by next, restoring pos.
func (t *tokenizer) backup() {
	if err := t.r.UnreadRune(); err != nil {
		return
	}
	if t.hasPrev {
		t.pos = t.prevPos
		t.hasPrev = false
	}
}

// peek returns the next rune without advancing. Returns 0, io.EOF at EOF.
func (t *tokenizer) peek() (rune, error) {
	r, err := t.next()
	if err != nil {
		return 0, err
	}
	t.backup()
	return r, nil
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

// tokenize is the top-level dispatch action.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	// Snapshot the position BEFORE reading so the start-of-token Pos is the
	// position of the first rune of the token (next() advances column past it).
	startPos := t.pos
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		yield(Token{}, err)
		return nil
	}

	switch {
	case r == '\n':
		if !yield(Token{Pos: startPos, Type: TokenNewline, Value: "\n"}, nil) {
			return nil
		}
		return tokenize
	case r == ' ' || r == '\t' || r == '\r':
		// skip whitespace
		return tokenize
	case r == '#':
		return scanLineComment(startPos)
	case r == '"':
		return scanString(startPos)
	case isIdentStart(r):
		return scanIdentifier(startPos, r)
	case isDigit(r):
		return scanNumber(startPos, r)
	case r == '=':
		// could be `=` or `==`
		next, err := t.peek()
		if err == nil && next == '=' {
			_, _ = t.next() // consume second '='
			if !yield(Token{Pos: startPos, Type: TokenSymbol, Value: "=="}, nil) {
				return nil
			}
			return tokenize
		}
		if !yield(Token{Pos: startPos, Type: TokenSymbol, Value: "="}, nil) {
			return nil
		}
		return tokenize
	case r == '{' || r == '}' || r == '(' || r == ')' || r == ';' || r == '&':
		if !yield(Token{Pos: startPos, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	default:
		yield(Token{}, &UnexpectedCharacterError{Pos: startPos, Char: r})
		return nil
	}
}

// scanLineComment consumes the rest of a `#`-introduced line comment until
// a newline or EOF. The leading `#` and any single leading space are dropped
// from Value.
func scanLineComment(start Pos) tokenizerAction {
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
			if r == '\n' {
				t.backup()
				break
			}
			b.WriteRune(r)
		}
		val := strings.TrimPrefix(b.String(), " ")
		if !yield(Token{Pos: start, Type: TokenComment, Value: val}, nil) {
			return nil
		}
		return tokenize
	}
}

// scanString consumes a `"..."` string. Escapes are not processed for this
// slice (the bool/conditional fixture does not exercise them).
func scanString(start Pos) tokenizerAction {
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
			if r == '"' {
				if !yield(Token{Pos: start, Type: TokenString, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if r == '\n' {
				yield(Token{}, &UnterminatedStringError{Pos: start})
				return nil
			}
			b.WriteRune(r)
		}
	}
}

// scanIdentifier consumes the rest of an identifier whose first rune was r.
func scanIdentifier(start Pos, first rune) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		b.WriteRune(first)
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

// scanNumber consumes a decimal integer. Only the form needed for the
// bool/conditional slice is implemented (the value field of `record number`
// is not exercised in the targeted tests, but accepting it keeps the
// tokenizer well-formed for ports like 8080 in the spec examples).
func scanNumber(start Pos, first rune) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		b.WriteRune(first)
		for {
			r, err := t.next()
			if err == io.EOF {
				break
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isDigit(r) {
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

// Tokenize streams tokens from r as an iter.Seq2[Token, error].
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		t := &tokenizer{r: bufio.NewReader(r), pos: Pos{Line: 1, Column: 1}}
		for action := tokenize; action != nil; {
			action = action(t, yield)
		}
	}
}
