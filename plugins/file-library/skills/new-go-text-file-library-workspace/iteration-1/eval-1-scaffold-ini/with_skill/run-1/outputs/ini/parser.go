package ini

import (
	"fmt"
	"io"
	"iter"
	"strings"
)

// File is the top-level AST node produced by Parse. The implementer fills
// out its fields (sections, key/value pairs, comments) from SPEC.md.
type File struct {
	// Nodes are the top-level AST elements of the file in source order.
	// The placeholder element type lets the package compile while the
	// implementer settles on the concrete shape (e.g. *Section, *KeyValue).
	Nodes []Type
}

// Type is the marker interface satisfied by every concrete AST node.
// Concrete types live alongside the parser actions that produce them and
// must implement isType to participate.
type Type interface {
	isType()
}

// UnexpectedEndOfTokensError is returned when the parser pulls past the end
// of the token stream while it still expected more input.
type UnexpectedEndOfTokensError struct{}

// Error implements the error interface.
func (e *UnexpectedEndOfTokensError) Error() string {
	return "unexpected end of token stream"
}

// UnexpectedTokenError is returned by parser.expect when the next token's
// type is not in the accepted set.
type UnexpectedTokenError struct {
	Got  Token
	Want []TokenType
}

// Error implements the error interface.
func (e *UnexpectedTokenError) Error() string {
	names := make([]string, 0, len(e.Want))
	for _, tt := range e.Want {
		names = append(names, tt.String())
	}
	return fmt.Sprintf("unexpected token %s; want one of [%s]", e.Got, strings.Join(names, ", "))
}

// parser pulls tokens from the upstream iter.Seq2 via iter.Pull2.
type parser struct {
	next func() (Token, error, bool)
}

// expect pulls the next token and verifies its type is one of the allowed
// types. On mismatch it returns *UnexpectedTokenError; on premature
// exhaustion it returns *UnexpectedEndOfTokensError.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.next()
	if !ok {
		return Token{}, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return Token{}, err
	}
	for _, tt := range types {
		if tok.Type == tt {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedTokenError{Got: tok, Want: types}
}

// parserAction is one step of the parser state machine. T is the AST node
// being built (e.g. *File at the top level, *Section inside a section).
// Returning (nil, nil) completes successfully; (nil, err) terminates with
// error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// Parse reads INI-formatted text from r and returns the parsed *File.
func Parse(r io.Reader) (*File, error) {
	next, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{next: next}
	f := &File{}

	var err error
	for action := parseFile; action != nil && err == nil; {
		action, err = action(p, f)
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// parseFile is the top-level entry action. The implementer wires up dispatch
// here (sections, key/value pairs, comments) and uses the inner action loop
// pattern for any nested structure (see CLAUDE.md).
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	return nil, nil
}
