package kvr

import (
	"fmt"
	"io"
	"iter"
)

// RecordType identifies the value type of a Record.
type RecordType int

const (
	RecordTypeInvalid RecordType = iota
	RecordTypeString
)

// String returns a human-readable name for a RecordType.
func (rt RecordType) String() string {
	switch rt {
	case RecordTypeString:
		return "string"
	default:
		return fmt.Sprintf("RecordType(%d)", int(rt))
	}
}

// Record is a single typed key-value declaration of the form
//
//	record TYPE KEY = VALUE
//
// at the top level of a KVR file.
type Record struct {
	Type  RecordType
	Key   string
	Value string
}

func (Record) isType() {}

// File is the top-level AST node. Implementer extends this with Records,
// Blocks, and any other top-level constructs the format introduces.
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

// UnknownTypeError is returned when a record's type name is not one of the
// supported types ("string"). The position points at the type-name token.
type UnknownTypeError struct {
	Pos  Pos
	Name string
}

func (e *UnknownTypeError) Error() string {
	return fmt.Sprintf("unknown type %q at %d:%d", e.Name, e.Pos.Line, e.Pos.Column)
}

// UnknownStatementError is returned when a top-level identifier is something
// other than "record" (or other supported statement openers).
type UnknownStatementError struct {
	Pos  Pos
	Name string
}

func (e *UnknownStatementError) Error() string {
	return fmt.Sprintf("unknown statement %q at %d:%d", e.Name, e.Pos.Line, e.Pos.Column)
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

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action. The implementer extends this to dispatch
// on token types into record-parsing, block-parsing, and other actions.
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
	default:
		return nil, &UnknownStatementError{Pos: tok.Pos, Name: tok.Value}
	}
}

// parseRecord parses a Record using an inner action loop. The opening
// "record" identifier has already been consumed by parseFile.
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

// parseRecordType reads the type-name identifier (currently only "string").
func parseRecordType(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	switch tok.Value {
	case "string":
		r.Type = RecordTypeString
	default:
		return nil, &UnknownTypeError{Pos: tok.Pos, Name: tok.Value}
	}
	return parseRecordKey, nil
}

// parseRecordKey reads the record's key identifier.
func parseRecordKey(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	r.Key = tok.Value
	return parseRecordEquals, nil
}

// parseRecordEquals consumes the "=" symbol between key and value.
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

// parseRecordValue reads the record's value, which must match the declared
// type. Currently only string-typed values are supported.
func parseRecordValue(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err, ok := p.next()
	if !ok {
		return nil, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return nil, err
	}
	switch r.Type {
	case RecordTypeString:
		if tok.Type != TokenString {
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenString}}
		}
		r.Value = tok.Value
	default:
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenString}}
	}
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
