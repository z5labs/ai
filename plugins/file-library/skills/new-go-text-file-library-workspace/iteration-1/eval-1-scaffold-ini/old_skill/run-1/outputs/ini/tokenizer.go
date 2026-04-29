package ini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"unicode"
)

// Pos records a 1-based line and column position within the input stream.
type Pos struct {
	Line   int
	Column int
}

// String renders Pos as "line:column".
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// TokenType discriminates Token values.
type TokenType int

const (
	// TokenInvalid is the zero value; it should never be yielded by the
	// tokenizer and serves only as a sentinel for uninitialized tokens.
	TokenInvalid TokenType = iota
	// TokenComment is a comment token (its lexical conventions are
	// format-specific; the implementer fills in the action that yields it).
	TokenComment
	// TokenIdentifier is a bare-word identifier (e.g. a key name).
	TokenIdentifier
	// TokenSymbol is a single-rune punctuation token.
	TokenSymbol
	// TokenString is a quoted string literal.
	TokenString
	// TokenNumber is a numeric literal.
	TokenNumber
)

// String returns a human-readable name for a TokenType.
func (tt TokenType) String() string {
	switch tt {
	case TokenInvalid:
		return "Invalid"
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

// String renders a Token in a debug-friendly form.
func (t Token) String() string {
	return fmt.Sprintf("%s %s %q", t.Pos, t.Type, t.Value)
}

// UnexpectedCharacterError is returned when the tokenizer encounters a rune
// that no action wanted at the current position.
type UnexpectedCharacterError struct {
	Pos  Pos
	Char rune
}

func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("unexpected character %q at %s", e.Char, e.Pos)
}

// tokenizer is the lexical state shared across actions. It wraps a
// *bufio.Reader for one-rune lookahead and tracks the current position so
// every yielded token records where it came from.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	lastPos Pos // position before the most recent next() call (used by backup)
	hasLast bool
}

// newTokenizer constructs a tokenizer over r with position 1:1.
func newTokenizer(r io.Reader) *tokenizer {
	return &tokenizer{
		r:   bufio.NewReader(r),
		pos: Pos{Line: 1, Column: 1},
	}
}

// next reads one rune and advances pos. It returns io.EOF at end of input.
func (t *tokenizer) next() (rune, error) {
	t.lastPos = t.pos
	t.hasLast = true
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

// backup rewinds the most recent rune read by next() and restores the
// position. Calling backup more than once between next() calls is not
// supported.
func (t *tokenizer) backup() error {
	if !t.hasLast {
		return errors.New("tokenizer: backup called without a prior next")
	}
	if err := t.r.UnreadRune(); err != nil {
		return err
	}
	t.pos = t.lastPos
	t.hasLast = false
	return nil
}

// tokenizerAction is one step in the tokenizer state machine. An action
// reads runes, optionally yields a token, and returns the next action (or
// nil to stop).
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// Tokenize returns an iter.Seq2 that yields tokens read from r. Errors are
// surfaced inline via the second value; once an error is yielded iteration
// stops.
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		t := newTokenizer(r)
		for action := tokenizeStart; action != nil; {
			action = action(t, yield)
		}
	}
}

// yieldToken yields tok and returns the dispatch action so the most common
// "emit and keep tokenizing" path is a one-liner.
func yieldToken(tok Token, yield func(Token, error) bool) tokenizerAction {
	if !yield(tok, nil) {
		return nil
	}
	return tokenizeStart
}

// yieldError yields err and returns nil so every error path is consistent.
func yieldError(err error, yield func(Token, error) bool) tokenizerAction {
	yield(Token{}, err)
	return nil
}

// skipWhitespace consumes whitespace runes (excluding newline by default;
// implementers can adjust per-format) and chains back to dispatch.
func skipWhitespace(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	for {
		r, err := t.next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return yieldError(err, yield)
		}
		if !unicode.IsSpace(r) {
			if err := t.backup(); err != nil {
				return yieldError(err, yield)
			}
			return tokenizeStart
		}
	}
}

// tokenizeStart is the entry-point dispatch action. The implementer fans
// out from here based on the first rune.
func tokenizeStart(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return yieldError(err, yield)
	}
	// Stub: the implementer dispatches on r here. For now we back up and
	// stop so the scaffold compiles and exits cleanly.
	if err := t.backup(); err != nil {
		return yieldError(err, yield)
	}
	_ = r
	return nil
}
