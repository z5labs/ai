package kvr

import (
	"fmt"
	"io"
	"iter"
)

// Record is the AST node for `record TYPE KEY = VALUE`.
type Record struct {
	Type  string
	Key   string
	Value string
}

func (Record) isType() {}

// File is the top-level AST node. It holds the sequence of records parsed
// from the source.
type File struct {
	Records []Record
}

// Type is the marker interface for AST node types in the KVR AST.
// Concrete node types implement isType() to satisfy this interface.
type Type interface {
	isType()
}

// UnexpectedEndOfTokensError is returned when the parser ran out of tokens
// while it still expected more.
type UnexpectedEndOfTokensError struct{}

func (e *UnexpectedEndOfTokensError) Error() string {
	return "unexpected end of tokens"
}

// UnexpectedTokenError is returned when the parser saw a token whose type was
// not in the set of acceptable types for the current grammar position.
type UnexpectedTokenError struct {
	Got  Token
	Want []TokenType
}

func (e *UnexpectedTokenError) Error() string {
	return fmt.Sprintf("unexpected token %s, expected one of %v", e.Got, e.Want)
}

// UnexpectedKeywordError is returned when an identifier was where a specific
// keyword (`record`, `string`, ...) was required but the value didn't match.
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %s, expected one of %v", e.Got, e.Want)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
type parser struct {
	next func() (Token, error, bool)
	stop func()
}

// expect pulls the next token and verifies its type matches one of the given
// types. Use it everywhere the grammar requires a specific token; never
// inline the type check.
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

// expectKeyword pulls the next token, requires TokenIdentifier, and requires
// its value to match one of the given keywords.
func (p *parser) expectKeyword(words ...string) (Token, error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return Token{}, err
	}
	for _, w := range words {
		if tok.Value == w {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedKeywordError{Got: tok, Want: words}
}

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action. It peeks the next token; if there are
// none, it terminates. Otherwise it dispatches into the appropriate
// statement-level action based on the leading keyword.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	tok, err, ok := p.next()
	if !ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if tok.Type != TokenIdentifier {
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier}}
	}
	switch tok.Value {
	case "record":
		return parseRecord, nil
	}
	return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record"}}
}

// parseRecord drives the record parse via the inner action loop pattern.
// Each grammar position (type, key, equals, value) is its own action.
func parseRecord(p *parser, f *File) (parserAction[*File], error) {
	rec := &Record{}
	var err error
	for action := parseRecordType; action != nil && err == nil; {
		action, err = action(p, rec)
	}
	if err != nil {
		return nil, err
	}
	f.Records = append(f.Records, *rec)
	return parseFile, nil
}

// parseRecordType expects the type keyword (`string`) and stores it.
func parseRecordType(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expectKeyword("string")
	if err != nil {
		return nil, err
	}
	r.Type = tok.Value
	return parseRecordKey, nil
}

// parseRecordKey expects the key identifier and stores it.
func parseRecordKey(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	r.Key = tok.Value
	return parseRecordEquals, nil
}

// parseRecordEquals expects the `=` symbol.
func parseRecordEquals(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return nil, err
	}
	if tok.Value != "=" {
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
	}
	return parseRecordValue, nil
}

// parseRecordValue expects a string-literal value and stores its decoded text.
func parseRecordValue(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenString)
	if err != nil {
		return nil, err
	}
	r.Value = tok.Value
	return nil, nil
}

// Parse reads a KVR file from r.
func Parse(r io.Reader) (*File, error) {
	next, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{
		next: func() (Token, error, bool) {
			tok, err, ok := next()
			return tok, err, ok
		},
		stop: stop,
	}

	f := &File{}
	for action, err := parseFile, error(nil); action != nil; {
		action, err = action(p, f)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}
