// Copyright (c) 2026 z5labs
//
// Licensed under the MIT License (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT

package toml

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node produced by [Parse]. It represents a
// complete TOML document.
type File struct {
	// Items is a placeholder for the top-level entries of the document
	// (key/value pairs, table headers, etc.). The concrete shape will be
	// fleshed out as the parser is implemented.
	Items []Type
}

// Type is the marker interface implemented by every node in the AST.
type Type interface {
	tomlType()
}

// UnexpectedEndOfTokensError is returned when the parser exhausts the
// token stream while still expecting more input.
type UnexpectedEndOfTokensError struct {
	// Expected describes what the parser was looking for when the stream
	// ended. It may be empty.
	Expected string
}

func (e *UnexpectedEndOfTokensError) Error() string {
	if e.Expected == "" {
		return "unexpected end of tokens"
	}
	return fmt.Sprintf("unexpected end of tokens, expected %s", e.Expected)
}

// UnexpectedTokenError is returned when the parser receives a token that
// does not match what it was expecting.
type UnexpectedTokenError struct {
	Got      Token
	Expected string
}

func (e *UnexpectedTokenError) Error() string {
	if e.Expected == "" {
		return fmt.Sprintf("unexpected token %s", e.Got)
	}
	return fmt.Sprintf("unexpected token %s, expected %s", e.Got, e.Expected)
}

// parser wraps a pull-style token source.
type parser struct {
	next func() (Token, error, bool)
}

// expect pulls the next token and verifies it has the given type. The
// returned token is the one that was consumed. If the stream is exhausted
// or the token type does not match, an error is returned.
func (p *parser) expect(tt TokenType) (Token, error) {
	tok, err, ok := p.next()
	if err != nil {
		return Token{}, err
	}
	if !ok {
		return Token{}, &UnexpectedEndOfTokensError{Expected: tt.String()}
	}
	if tok.Type != tt {
		return Token{}, &UnexpectedTokenError{Got: tok, Expected: tt.String()}
	}
	return tok, nil
}

// parserAction is one step of the parser state machine. T is the AST node
// being built (typically [*File] at the top level). Returning a nil action
// terminates the loop.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// Parse reads a TOML document from r and returns its AST.
func Parse(r io.Reader) (*File, error) {
	next, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{next: next}
	f := &File{}

	action := parseStart
	for action != nil {
		nextAction, err := action(p, f)
		if err != nil {
			return nil, err
		}
		action = nextAction
	}
	return f, nil
}

// parseStart is the entry point of the parser state machine. The scaffold
// returns immediately; real productions will dispatch on the next token.
func parseStart(p *parser, f *File) (parserAction[*File], error) {
	_ = p
	_ = f
	return nil, nil
}
