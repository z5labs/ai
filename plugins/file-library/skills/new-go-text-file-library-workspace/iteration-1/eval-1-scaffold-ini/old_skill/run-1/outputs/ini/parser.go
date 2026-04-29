package ini

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node for a parsed INI file. The implementer
// fills this in as concrete sections, keys, and values are added.
type File struct {
	// Nodes is a placeholder slice for the top-level AST nodes. Replace or
	// extend this as the real AST shape solidifies.
	Nodes []Type
}

// Type is the marker interface implemented by every AST node so the parser
// and printer can talk about "any node" without a sprawling type switch.
type Type interface {
	iniNode()
}

// UnexpectedEndOfTokensError is returned when the parser pulls a token but
// the stream has been exhausted in a position that requires more input.
type UnexpectedEndOfTokensError struct {
	// Want is the set of token types that would have been valid here. Empty
	// when the parser only knows "more input is required".
	Want []TokenType
}

func (e *UnexpectedEndOfTokensError) Error() string {
	if len(e.Want) == 0 {
		return "unexpected end of tokens"
	}
	return fmt.Sprintf("unexpected end of tokens; want one of %v", e.Want)
}

// UnexpectedTokenError is returned by parser.expect when the next token does
// not match any of the expected types.
type UnexpectedTokenError struct {
	Got  Token
	Want []TokenType
}

func (e *UnexpectedTokenError) Error() string {
	return fmt.Sprintf("unexpected token %s; want one of %v", e.Got, e.Want)
}

// parser is the parsing state shared across actions. It owns the
// pull-based token source produced by iter.Pull2(Tokenize(r)).
type parser struct {
	next func() (Token, error, bool)
}

// expect pulls the next token and verifies its type matches one of the
// supplied types. It returns the token on success, or a typed error
// (UnexpectedEndOfTokensError or UnexpectedTokenError) on failure.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.next()
	if !ok {
		return Token{}, &UnexpectedEndOfTokensError{Want: types}
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

// parserAction is the generic state-machine step used by every level of the
// parser. T is the AST node currently being built; nested parsers run their
// own loop with a different T.
//
// Returning (nil, nil) completes the loop successfully. Returning
// (nil, err) terminates with the supplied error. Every error path returns a
// nil next action so the loop stays monotone.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// Parse reads tokens from r and returns the parsed *File. It is the public
// entry point for the parser pipeline.
func Parse(r io.Reader) (*File, error) {
	next, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{next: next}
	f := &File{}

	var err error
	for action := parseStart; action != nil && err == nil; {
		action, err = action(p, f)
	}
	if err != nil {
		return nil, err
	}
	return f, nil
}

// parseStart is the top-level entry action. The implementer dispatches on
// the first token here. The scaffold returns (nil, nil) so the parser
// terminates cleanly on empty input.
func parseStart(p *parser, f *File) (parserAction[*File], error) {
	_ = p
	_ = f
	return nil, nil
}
