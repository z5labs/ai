package graphql

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node — a parsed GraphQL schema document.
type File struct {
	// Definitions holds the top-level definitions found in the document.
	// The implementer fills this slice with concrete Type values.
	Definitions []Type
}

// Type is the interface satisfied by every concrete AST node in the
// package. The unexported marker method prevents foreign types from
// accidentally satisfying the interface.
type Type interface {
	isType()
}

// UnexpectedEndOfTokensError is returned when the parser needs another
// token but the tokenizer stream has been exhausted.
type UnexpectedEndOfTokensError struct{}

// Error implements the error interface.
func (e *UnexpectedEndOfTokensError) Error() string {
	return "unexpected end of tokens"
}

// UnexpectedTokenError is returned when expect() reads a token whose type
// does not match any of the wanted types.
type UnexpectedTokenError struct {
	Got  Token
	Want []TokenType
}

// Error implements the error interface.
func (e *UnexpectedTokenError) Error() string {
	if len(e.Want) == 1 {
		return fmt.Sprintf("unexpected token %s at %s: want %s", e.Got.Type, e.Got.Pos, e.Want[0])
	}
	return fmt.Sprintf("unexpected token %s at %s: want one of %v", e.Got.Type, e.Got.Pos, e.Want)
}

// parser is the internal state for parsing. It wraps the pull-based view
// of the tokenizer's iter.Seq2 stream.
type parser struct {
	next func() (Token, error, bool)
}

// expect pulls the next token from the stream and verifies its type is one
// of the given types. It returns the token on success, or a typed error.
//
// expect is the only place type-checking is allowed in the parser; never
// inline a type comparison.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.next()
	if !ok {
		return Token{}, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return Token{}, err
	}
	for _, want := range types {
		if tok.Type == want {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedTokenError{Got: tok, Want: types}
}

// parserAction is one step of the parser state machine, generic over the
// AST node currently being built. Returning (nil, nil) completes
// successfully; (nil, err) terminates with an error.
//
// Generic actions let nested parsers reuse the same loop without an
// interface dance. For complex/nested types, use the inner action loop
// pattern (see CLAUDE.md) — never an inline for+switch.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level entry action. The scaffold returns
// (nil, nil) immediately so the empty-input test passes; the implementer
// fills in dispatch on the next token.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	return nil, nil
}

// Parse consumes r as a GraphQL document and returns the resulting AST.
// It is the only public parser surface: tests must drive Parse with real
// source strings, never construct AST nodes by hand.
func Parse(r io.Reader) (*File, error) {
	pull, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{next: pull}
	f := &File{}

	var err error
	for action := parserAction[*File](parseFile); action != nil && err == nil; {
		action, err = action(p, f)
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}
