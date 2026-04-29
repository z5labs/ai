// Copyright (c) 2026 z5labs
//
// Licensed under the MIT License (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT

package toml

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
)

// Pos describes a position within a TOML source document. Lines and Columns
// are 1-indexed.
type Pos struct {
	Line   int
	Column int
}

// String returns a "line:column" representation of the position, useful for
// error messages.
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// TokenType identifies the lexical category of a [Token].
type TokenType int

const (
	// TokenComment is a TOML comment beginning with '#' and ending at the
	// next newline.
	TokenComment TokenType = iota
	// TokenIdentifier is an unquoted bare key.
	TokenIdentifier
	// TokenSymbol is a punctuation symbol (=, ., [, ], {, }, ,).
	TokenSymbol
	// TokenString is a quoted string literal.
	TokenString
	// TokenNumber is a numeric literal (integer or float).
	TokenNumber
)

// String returns a human-readable name for the token type.
func (t TokenType) String() string {
	switch t {
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
		return fmt.Sprintf("TokenType(%d)", int(t))
	}
}

// Token is a single lexical unit produced by the tokenizer.
type Token struct {
	Pos   Pos
	Type  TokenType
	Value string
}

// String returns a human-readable representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("%s %s %q", t.Pos, t.Type, t.Value)
}

// UnexpectedCharacterError is returned when the tokenizer encounters a rune
// that is not valid in the current context.
type UnexpectedCharacterError struct {
	Pos  Pos
	Rune rune
}

func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("unexpected character %q at %s", e.Rune, e.Pos)
}

// tokenizer wraps a buffered reader and tracks the current position within
// the source.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos
	// last is the most recently read rune, returned by next. It is used by
	// backup to support a single-rune lookahead.
	last     rune
	hasLast  bool
	backedUp bool
}

// newTokenizer constructs a tokenizer that reads from r.
func newTokenizer(r io.Reader) *tokenizer {
	return &tokenizer{
		r:   bufio.NewReader(r),
		pos: Pos{Line: 1, Column: 1},
	}
}

// next reads the next rune from the input, advances the position, and
// returns it. It returns io.EOF (wrapped in error) when the input is
// exhausted.
func (t *tokenizer) next() (rune, error) {
	if t.backedUp {
		t.backedUp = false
		// Re-advance position to what it was before backup.
		t.prevPos = t.pos
		t.pos = advancePos(t.pos, t.last)
		return t.last, nil
	}
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prevPos = t.pos
	t.pos = advancePos(t.pos, r)
	t.last = r
	t.hasLast = true
	return r, nil
}

// backup undoes the most recent successful call to next. It may only be
// called once between calls to next.
func (t *tokenizer) backup() {
	if !t.hasLast || t.backedUp {
		return
	}
	t.backedUp = true
	t.pos = t.prevPos
}

// advancePos returns the position that follows pos after consuming r.
func advancePos(pos Pos, r rune) Pos {
	if r == '\n' {
		return Pos{Line: pos.Line + 1, Column: 1}
	}
	return Pos{Line: pos.Line, Column: pos.Column + 1}
}

// tokenizerAction is one step of the tokenizer state machine. Returning nil
// terminates the loop.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldErr emits an error to the consumer and returns nil to stop the
// state machine. It is the canonical way to propagate a tokenizer error.
func yieldErr(yield func(Token, error) bool, err error) tokenizerAction {
	yield(Token{}, err)
	return nil
}

// yieldToken emits tok to the consumer. It returns next as the follow-up
// action, or nil if the consumer requested early termination.
func yieldToken(yield func(Token, error) bool, tok Token, next tokenizerAction) tokenizerAction {
	if !yield(tok, nil) {
		return nil
	}
	return next
}

// skipWhitespace advances past any spaces, tabs, carriage returns, or
// newlines. It does not yield any tokens.
func skipWhitespace(t *tokenizer) error {
	for {
		r, err := t.next()
		if err != nil {
			return err
		}
		switch r {
		case ' ', '\t', '\r', '\n':
			continue
		default:
			t.backup()
			return nil
		}
	}
}

// Tokenize returns an iterator over the tokens in r. Iteration stops at the
// first error or when the consumer breaks out of the loop.
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		t := newTokenizer(r)
		action := tokenizeStart
		for action != nil {
			action = action(t, yield)
		}
	}
}

// tokenizeStart is the entry point of the tokenizer state machine. The
// scaffold simply reads one rune to verify the wiring and then exits.
func tokenizeStart(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	if err := skipWhitespace(t); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return yieldErr(yield, err)
	}
	r, err := t.next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return yieldErr(yield, err)
	}
	// TODO: dispatch on r to comment/string/number/identifier/symbol
	// actions. For now the scaffold simply consumes one rune and exits.
	_ = r
	return nil
}
