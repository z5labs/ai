package kvr

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node. It carries the file's records in source
// order.
type File struct {
	Records []Record
}

// Type is the marker interface for AST node types in the KVR AST. Concrete
// node types implement isType() to satisfy this interface.
type Type interface {
	isType()
}

// Record is a single typed key-value declaration of the form
// `record <Type> <Key> = <Value>`. For the string record family, Type is
// "string" and Value is the decoded string content.
type Record struct {
	Type  string
	Key   string
	Value string
}

func (Record) isType() {}

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

// UnexpectedKeywordError is returned when an identifier with a reserved-word
// role (e.g. "record") was expected but a different identifier appeared.
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %q, expected one of %v", e.Got.Value, e.Want)
}

// TypeMismatchError is returned when a record's value-token type does not
// match its declared type (e.g. `record string K = 42`).
type TypeMismatchError struct {
	Type string
	Got  Token
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: record declared %s but value is %s at %d:%d",
		e.Type, e.Got.Type, e.Got.Pos.Line, e.Got.Pos.Column)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
type parser struct {
	next func() (Token, error, bool)
	stop func()
}

// peek returns the next token without consuming the underlying iterator. The
// public API is `expect`; peek is internal to dispatch helpers that need to
// branch on the next token type.
func (p *parser) pull() (Token, error, bool) {
	return p.next()
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

// expectKeyword is a small helper layered on top of expect: pull an
// identifier and verify its Value matches one of the given keywords. The
// returned token is the matched identifier.
func (p *parser) expectKeyword(keywords ...string) (Token, error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return Token{}, err
	}
	for _, kw := range keywords {
		if tok.Value == kw {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedKeywordError{Got: tok, Want: keywords}
}

// expectSymbol pulls a symbol token and verifies its Value matches.
func (p *parser) expectSymbol(symbols ...string) (Token, error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return Token{}, err
	}
	for _, sym := range symbols {
		if tok.Value == sym {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
}

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action. It peeks the next token to decide which
// statement-kind action to dispatch, or returns (nil, nil) at end-of-input.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	tok, err, ok := p.pull()
	if !ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case TokenIdentifier:
		switch tok.Value {
		case "record":
			return parseRecord, nil
		default:
			return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record"}}
		}
	default:
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier}}
	}
}

// parseRecord reads the body of a `record` statement (the `record` keyword
// has already been consumed by parseFile). It uses the inner action loop
// pattern: each state of the record parse — type, key, equals, value — is
// its own parserAction[*Record].
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

func parseRecordType(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expectKeyword("string")
	if err != nil {
		return nil, err
	}
	rec.Type = tok.Value
	return parseRecordKey, nil
}

func parseRecordKey(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	rec.Key = tok.Value
	return parseRecordEquals, nil
}

func parseRecordEquals(p *parser, rec *Record) (parserAction[*Record], error) {
	if _, err := p.expectSymbol("="); err != nil {
		return nil, err
	}
	return parseRecordValue, nil
}

func parseRecordValue(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenString)
	if err != nil {
		return nil, err
	}
	if rec.Type != "string" {
		return nil, &TypeMismatchError{Type: rec.Type, Got: tok}
	}
	rec.Value = tok.Value
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
