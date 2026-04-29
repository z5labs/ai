package graphql

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"unicode"
)

// Pos identifies a position in the input by line and column. Both are 1-based.
type Pos struct {
	Line   int
	Column int
}

// String renders the position as "line:column".
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// TokenType enumerates the kinds of tokens a GraphQL tokenizer can emit.
//
// Named values are intentional: when a parser test asserts an unexpected token
// kind, the message reads as a name rather than a number.
type TokenType int

const (
	// TokenComment is a comment token (e.g. `# comment`).
	TokenComment TokenType = iota
	// TokenIdentifier is a name/identifier token.
	TokenIdentifier
	// TokenSymbol is a punctuation/symbol token (e.g. `{`, `}`, `:`).
	TokenSymbol
	// TokenString is a string-literal token.
	TokenString
	// TokenNumber is a numeric-literal token.
	TokenNumber
)

// String returns the human-readable name of the token type.
func (tt TokenType) String() string {
	switch tt {
	case TokenComment:
		return "Comment"
	case TokenIdentifier:
		return "Identifier"
	case TokenSymbol:
		return "Symbol"
	case TokenString:
		return "String"
	case TokenNumber:
		return "Number"
	default:
		return fmt.Sprintf("TokenType(%d)", int(tt))
	}
}

// Token is a single token emitted by the tokenizer.
type Token struct {
	Pos   Pos
	Type  TokenType
	Value string
}

// String renders a token in a compact "Type(value)@pos" form.
func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%s", t.Type, t.Value, t.Pos)
}

// UnexpectedCharacterError reports a rune that no tokenizer action accepted.
type UnexpectedCharacterError struct {
	Pos  Pos
	Char rune
}

// Error implements the error interface.
func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("unexpected character %q at %s", e.Char, e.Pos)
}

// tokenizer is the internal state for tokenization. It wraps a bufio.Reader
// for one-rune lookahead and tracks the current position.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos
	hasPrev bool
}

// newTokenizer constructs a tokenizer with position starting at line 1,
// column 0 — the first call to next() advances to column 1.
func newTokenizer(r io.Reader) *tokenizer {
	return &tokenizer{
		r:   bufio.NewReader(r),
		pos: Pos{Line: 1, Column: 0},
	}
}

// next returns the next rune from the input, updating pos. It returns
// io.EOF (and the zero rune) when the input is exhausted.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prevPos = t.pos
	t.hasPrev = true
	if r == '\n' {
		t.pos.Line++
		t.pos.Column = 0
	} else {
		t.pos.Column++
	}
	return r, nil
}

// backup rewinds the last rune read by next(). It is safe to call exactly
// once after a successful next(); calling it twice in a row is a programmer
// error in this package and a no-op here.
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

// tokenizerAction is one step of the tokenizer state machine. An action
// reads runes, optionally yields tokens, and returns the next action to
// run. Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldThen emits tok and, if the consumer is still pulling, returns next.
// If yield returns false (consumer stopped), iteration ends.
func yieldThen(tok Token, next tokenizerAction) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		if !yield(tok, nil) {
			return nil
		}
		return next
	}
}

// yieldErrorAndStop emits err and ends iteration. Every error path uses this
// helper so the convention is consistent across the tokenizer.
func yieldErrorAndStop(err error) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		yield(Token{}, err)
		return nil
	}
}

// skipWhitespace consumes whitespace runes and chains back to dispatch.
func skipWhitespace(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	for {
		r, err := t.next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return yieldErrorAndStop(err)
		}
		if !unicode.IsSpace(r) {
			t.backup()
			return tokenizeDispatch
		}
	}
}

// tokenizeDispatch is the entry-point action. Real implementations dispatch
// on the next rune; the scaffold reads one rune, terminates on EOF, and
// otherwise terminates as well — the implementer wires up the real switch.
func tokenizeDispatch(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	_, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return yieldErrorAndStop(err)
	}
	// Implementer: replace this stub with the real dispatch switch.
	return nil
}

// Tokenize returns an iter.Seq2 that lazily yields tokens from r.
// Errors are surfaced inline at the position they occur; iteration stops
// after the first error.
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		t := newTokenizer(r)
		for action := tokenizerAction(tokenizeDispatch); action != nil; {
			action = action(t, yield)
		}
	}
}
