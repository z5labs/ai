package toml

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"unicode"
)

// Pos is the source position of a token, measured in 1-based line and column.
type Pos struct {
	Line   int
	Column int
}

// String renders Pos as "line:column".
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// TokenType enumerates the kinds of tokens the tokenizer can emit.
type TokenType int

const (
	// TokenComment is a comment token.
	TokenComment TokenType = iota
	// TokenIdentifier is a bare identifier (key name, etc.).
	TokenIdentifier
	// TokenSymbol is a punctuation/operator symbol (e.g. '=', '[').
	TokenSymbol
	// TokenString is a quoted string literal.
	TokenString
	// TokenNumber is a numeric literal.
	TokenNumber
)

// String returns a human-readable name for the TokenType. Named values pay
// for themselves the first time a test failure prints a token type.
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

// Token is a single lexical unit produced by the tokenizer.
type Token struct {
	Pos   Pos
	Type  TokenType
	Value string
}

// String renders the Token in a debug-friendly form.
func (t Token) String() string {
	return fmt.Sprintf("%s %s %q", t.Pos, t.Type, t.Value)
}

// UnexpectedCharacterError is returned by the tokenizer when it encounters a
// rune no action wanted to consume.
type UnexpectedCharacterError struct {
	Pos  Pos
	Char rune
}

// Error implements the error interface.
func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("toml: unexpected character %q at %s", e.Char, e.Pos)
}

// tokenizer wraps an *bufio.Reader and tracks position. It is the state
// passed to every tokenizerAction.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos
	hasPrev bool
}

// newTokenizer constructs a tokenizer ready to emit tokens starting at 1:1.
func newTokenizer(r io.Reader) *tokenizer {
	return &tokenizer{
		r:   bufio.NewReader(r),
		pos: Pos{Line: 1, Column: 1},
	}
}

// next advances one rune and updates Pos. The returned error is io.EOF at
// end of input, or any other error from the underlying reader.
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

// backup rewinds the last rune read by next, restoring Pos. backup may only
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

// tokenizerAction is one step of the tokenizer state machine. An action
// reads runes, optionally yields tokens, and returns the next action to
// run. Returning nil ends the iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldThen emits tok and, if the consumer is still pulling, returns next.
// This is the most common ending of any action.
func yieldThen(tok Token, next tokenizerAction) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		if !yield(tok, nil) {
			return nil
		}
		return next
	}
}

// yieldErrorAndStop emits err and terminates the iteration. Every error
// path uses this so the convention is consistent.
func yieldErrorAndStop(err error) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		yield(Token{}, err)
		return nil
	}
}

// skipWhitespace consumes whitespace runes (excluding newlines, which many
// formats treat as significant) and chains back to dispatch.
func skipWhitespace(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	for {
		r, err := t.next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return yieldErrorAndStop(err)
		}
		if !unicode.IsSpace(r) || r == '\n' {
			t.backup()
			return tokenizeDispatch
		}
	}
}

// tokenizeDispatch is the entry-point action: it reads one rune and decides
// which sub-action to run. The implementer wires up the dispatch switch as
// the spec is filled in.
func tokenizeDispatch(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return yieldErrorAndStop(err)
	}
	// Stub: the real dispatch will branch on r here.
	_ = r
	return nil
}

// Tokenize returns an iter.Seq2 that lazily emits tokens from r.
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		t := newTokenizer(r)
		for action := tokenizerAction(tokenizeDispatch); action != nil; {
			action = action(t, yield)
		}
	}
}
