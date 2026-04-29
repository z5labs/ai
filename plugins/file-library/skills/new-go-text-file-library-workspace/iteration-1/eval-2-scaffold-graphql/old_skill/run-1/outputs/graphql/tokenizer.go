package graphql

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"unicode"
)

// Pos is a position within a source document.
type Pos struct {
	Line   int
	Column int
}

// String returns the position formatted as "line:column".
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// TokenType identifies the kind of a Token.
type TokenType int

const (
	// TokenComment is a source-level comment token.
	TokenComment TokenType = iota
	// TokenIdentifier is a name token (e.g., type name, field name).
	TokenIdentifier
	// TokenSymbol is a punctuation or operator token.
	TokenSymbol
	// TokenString is a quoted string literal token.
	TokenString
	// TokenNumber is a numeric literal token.
	TokenNumber
)

// String returns a human-readable name for the TokenType.
func (tt TokenType) String() string {
	switch tt {
	case TokenComment:
		return "comment"
	case TokenIdentifier:
		return "identifier"
	case TokenSymbol:
		return "symbol"
	case TokenString:
		return "string"
	case TokenNumber:
		return "number"
	default:
		return fmt.Sprintf("TokenType(%d)", int(tt))
	}
}

// Token is a lexical token produced by the tokenizer.
type Token struct {
	Pos   Pos
	Type  TokenType
	Value string
}

// String returns a debug representation of the Token.
func (t Token) String() string {
	return fmt.Sprintf("%s %s %q", t.Pos, t.Type, t.Value)
}

// UnexpectedCharacterError indicates that the tokenizer encountered a rune it
// could not match against any known token rule.
type UnexpectedCharacterError struct {
	Pos  Pos
	Rune rune
}

// Error implements error.
func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("graphql: unexpected character %q at %s", e.Rune, e.Pos)
}

// tokenizer is the lexer state. It wraps a buffered reader and tracks the
// current source position.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos
	hasPrev bool
}

// newTokenizer returns a tokenizer that reads from r. Source positions begin
// at line 1, column 1.
func newTokenizer(r io.Reader) *tokenizer {
	return &tokenizer{
		r:   bufio.NewReader(r),
		pos: Pos{Line: 1, Column: 1},
	}
}

// next reads the next rune and advances the position. It returns io.EOF when
// the input is exhausted.
func (t *tokenizer) next() (rune, error) {
	ch, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prevPos = t.pos
	t.hasPrev = true
	if ch == '\n' {
		t.pos.Line++
		t.pos.Column = 1
	} else {
		t.pos.Column++
	}
	return ch, nil
}

// backup unreads the most recently read rune and restores the previous
// position. Calling backup twice in a row without an intervening next is not
// supported.
func (t *tokenizer) backup() {
	if !t.hasPrev {
		return
	}
	if err := t.r.UnreadRune(); err != nil {
		return
	}
	t.pos = t.prevPos
	t.hasPrev = false
}

// tokenizerAction is one step of the tokenizer state machine. Returning nil
// signals that tokenization is complete.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldErr emits an error to the consumer and returns nil to terminate the
// state machine.
func yieldErr(yield func(Token, error) bool, err error) tokenizerAction {
	yield(Token{}, err)
	return nil
}

// yieldToken emits tok to the consumer. If the consumer signals stop, it
// returns nil to terminate the state machine; otherwise it returns next.
func yieldToken(yield func(Token, error) bool, tok Token, next tokenizerAction) tokenizerAction {
	if !yield(tok, nil) {
		return nil
	}
	return next
}

// skipWhitespace consumes whitespace runes. It returns the next non-whitespace
// rune, or an error (typically io.EOF) when the input is exhausted.
func (t *tokenizer) skipWhitespace() (rune, error) {
	for {
		ch, err := t.next()
		if err != nil {
			return 0, err
		}
		if !unicode.IsSpace(ch) {
			return ch, nil
		}
	}
}

// Tokenize returns an iterator over tokens read from r. The iterator yields a
// final (zero, error) pair if a non-EOF error occurs; io.EOF terminates the
// iterator without an error.
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		tk := newTokenizer(r)
		for action := tokenize; action != nil; {
			action = action(tk, yield)
		}
	}
}

// tokenize is the entry-point action of the tokenizer state machine. The
// scaffold reads one rune and stops; real grammar will dispatch on the rune
// to specific sub-actions.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	_, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return yieldErr(yield, err)
	}
	return nil
}
