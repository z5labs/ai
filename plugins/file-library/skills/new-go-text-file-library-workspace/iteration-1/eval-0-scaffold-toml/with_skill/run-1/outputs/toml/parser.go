package toml

import (
	"fmt"
	"io"
	"iter"
	"strings"
)

// File is the top-level AST node returned by Parse. Replace the placeholder
// field with the real top-level structure once the spec is filled in.
type File struct {
	// Nodes holds top-level AST entries. Placeholder for the real shape.
	Nodes []Type
}

// Type is the marker interface every concrete AST node implements.
type Type interface {
	isType()
}

// UnexpectedEndOfTokensError is returned when the parser needs another
// token but the tokenizer has already reported io.EOF.
type UnexpectedEndOfTokensError struct{}

// Error implements the error interface.
func (e *UnexpectedEndOfTokensError) Error() string {
	return "toml: unexpected end of tokens"
}

// UnexpectedTokenError is returned when expect encounters a token whose
// type is not in the wanted set.
type UnexpectedTokenError struct {
	Got  Token
	Want []TokenType
}

// Error implements the error interface.
func (e *UnexpectedTokenError) Error() string {
	wantNames := make([]string, len(e.Want))
	for i, w := range e.Want {
		wantNames[i] = w.String()
	}
	return fmt.Sprintf(
		"toml: unexpected token %s at %s, want one of [%s]",
		e.Got.Type, e.Got.Pos, strings.Join(wantNames, ", "),
	)
}

// parser wraps a pull-based token source. Use expect for any grammar
// production that requires a specific TokenType.
type parser struct {
	next func() (Token, error, bool)
}

// expect pulls the next token and verifies its type matches one of types.
// Returns UnexpectedTokenError on mismatch and UnexpectedEndOfTokensError
// at EOF. Never inline the type check in actions; always call expect.
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
	return tok, &UnexpectedTokenError{Got: tok, Want: append([]TokenType(nil), types...)}
}

// parserAction is one step of the parser state machine. The type parameter
// T is the AST node currently being built; nested parses use the same
// loop with a different T. Returning (nil, nil) completes successfully;
// returning (nil, err) terminates with err.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level entry action. Returns (nil, nil) so the
// empty-input test passes against the stub.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	// Stub: the real implementation dispatches based on the next token.
	return nil, nil
}

// Parse reads tokens from r and constructs the AST.
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
