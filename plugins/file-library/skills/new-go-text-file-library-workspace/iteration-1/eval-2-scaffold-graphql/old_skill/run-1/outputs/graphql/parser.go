package graphql

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node, representing a parsed GraphQL schema
// document. The Definitions field is a placeholder that the implementer will
// replace with concrete definition node types.
type File struct {
	Definitions []Type
}

// Type is implemented by every concrete AST node type that can appear in a
// File. The unexported marker method keeps the closed-set property of the
// interface inside this package.
type Type interface {
	typeNode()
}

// UnexpectedEndOfTokensError indicates that the parser exhausted the token
// stream while still expecting more input.
type UnexpectedEndOfTokensError struct {
	Expected string
}

// Error implements error.
func (e *UnexpectedEndOfTokensError) Error() string {
	if e.Expected == "" {
		return "graphql: unexpected end of tokens"
	}
	return fmt.Sprintf("graphql: unexpected end of tokens, expected %s", e.Expected)
}

// UnexpectedTokenError indicates that the parser saw a token that did not
// match the grammar at the current position.
type UnexpectedTokenError struct {
	Got      Token
	Expected string
}

// Error implements error.
func (e *UnexpectedTokenError) Error() string {
	if e.Expected == "" {
		return fmt.Sprintf("graphql: unexpected token %s", e.Got)
	}
	return fmt.Sprintf("graphql: unexpected token %s, expected %s", e.Got, e.Expected)
}

// parser is the parser state. It pulls tokens from the tokenizer through the
// next function, which mirrors the signature returned by iter.Pull2.
type parser struct {
	next func() (Token, error, bool)
}

// expect pulls the next token and returns it if its type matches want. If the
// token stream is exhausted it returns *UnexpectedEndOfTokensError; if the
// token type is wrong it returns *UnexpectedTokenError; if the underlying
// tokenizer surfaces an error, that error is returned unchanged.
func (p *parser) expect(want TokenType, description string) (Token, error) {
	tok, err, ok := p.next()
	if !ok {
		return Token{}, &UnexpectedEndOfTokensError{Expected: description}
	}
	if err != nil {
		return Token{}, err
	}
	if tok.Type != want {
		return Token{}, &UnexpectedTokenError{Got: tok, Expected: description}
	}
	return tok, nil
}

// parserAction is one step of the parser state machine. The generic parameter
// T lets each action thread the AST node currently being built. Returning
// (nil, nil) signals successful completion; (nil, err) terminates with an
// error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// Parse reads a GraphQL schema document from r and returns the resulting
// File. The scaffold returns an empty *File; the implementer extends this
// state machine to populate Definitions.
func Parse(r io.Reader) (*File, error) {
	next, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{next: next}
	f := &File{}

	var err error
	for action := parseFile; action != nil; {
		action, err = action(p, f)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

// parseFile is the entry-point action of the parser state machine. The
// scaffold completes immediately; real grammar will dispatch to definition
// parsers based on lookahead.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	return nil, nil
}
