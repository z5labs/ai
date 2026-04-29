package ini

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"unicode"
)

// Pos is a 1-indexed position in the input source.
type Pos struct {
	Line   int
	Column int
}

// String renders Pos as "line:column".
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// TokenType is the kind of a Token.
type TokenType int

const (
	// TokenComment is a comment token (e.g. "; comment" or "# comment").
	TokenComment TokenType = iota
	// TokenIdentifier is a bare identifier (key name, section name).
	TokenIdentifier
	// TokenSymbol is a single punctuation rune (e.g. '=', '[', ']').
	TokenSymbol
	// TokenString is a quoted string literal.
	TokenString
	// TokenNumber is a numeric literal.
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

// Token is a single lexical token produced by the tokenizer.
type Token struct {
	Pos   Pos
	Type  TokenType
	Value string
}

// String renders a Token in the form "Type(value)@line:column".
func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%s", t.Type, t.Value, t.Pos)
}

// UnexpectedCharacterError is returned when the tokenizer encounters a rune
// that no tokenizer action wants to consume.
type UnexpectedCharacterError struct {
	Pos  Pos
	Char rune
}

// Error implements the error interface.
func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("unexpected character %q at %s", e.Char, e.Pos)
}

// tokenizer wraps a *bufio.Reader and tracks the current source position.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevCol int  // column before the most recent next(), used by backup()
	hasPrev bool // true when a rune has been read since the last backup()
	prevRn  rune // the rune most recently returned by next()
}

func newTokenizer(r io.Reader) *tokenizer {
	return &tokenizer{
		r:   bufio.NewReader(r),
		pos: Pos{Line: 1, Column: 1},
	}
}

// next advances by one rune and updates pos. It returns io.EOF at end of input.
func (t *tokenizer) next() (rune, error) {
	rn, _, err := t.r.ReadRune()
	if err != nil {
		t.hasPrev = false
		return 0, err
	}
	// Save state for backup().
	t.prevCol = t.pos.Column
	t.prevRn = rn
	t.hasPrev = true
	if rn == '\n' {
		t.pos.Line++
		t.pos.Column = 1
	} else {
		t.pos.Column++
	}
	return rn, nil
}

// backup rewinds the most recent rune read by next(). It is safe to call only
// once per next().
func (t *tokenizer) backup() {
	if !t.hasPrev {
		return
	}
	if err := t.r.UnreadRune(); err != nil {
		return
	}
	if t.prevRn == '\n' {
		t.pos.Line--
	}
	t.pos.Column = t.prevCol
	t.hasPrev = false
}

// tokenizerAction is one step of the tokenizer state machine. It may yield
// tokens via the supplied yield func, then returns the next action to run.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldThen calls yield(tok, nil) and, if the consumer is still listening,
// returns next; otherwise returns nil to end iteration.
func yieldThen(tok Token, next tokenizerAction) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		if !yield(tok, nil) {
			return nil
		}
		return next
	}
}

// yieldErrorAndStop yields the error to the consumer and ends iteration.
func yieldErrorAndStop(err error) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		yield(Token{}, err)
		return nil
	}
}

// skipWhitespace consumes runs of unicode whitespace and chains back to next.
func skipWhitespace(next tokenizerAction) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		for {
			rn, err := t.next()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return yieldErrorAndStop(err)
			}
			if !unicode.IsSpace(rn) {
				t.backup()
				return next
			}
		}
	}
}

// Tokenize returns an iter.Seq2 that lazily produces tokens read from r.
// Errors are surfaced inline at the position they occur; iteration stops on
// the first error or when the underlying reader returns io.EOF.
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		tk := newTokenizer(r)
		for action := tokenizeMain; action != nil; {
			action = action(tk, yield)
		}
	}
}

// tokenizeMain is the entry-point dispatch action. The implementer wires up
// the dispatch switch (comments, identifiers, symbols, strings, numbers) by
// returning the appropriate sub-action for the rune just read.
func tokenizeMain(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	_, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return yieldErrorAndStop(err)
	}
	// Stub: implementer dispatches here based on the leading rune.
	return nil
}
