package kvr

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node. It holds the ordered sequence of records
// declared at the top level of a KVR source file.
type File struct {
	Records []Record
}

// Record is the AST node for a `record TYPE KEY = VALUE` statement. Type is
// the type name as written in source (currently "string"); Key is the
// identifier following the type; Value is the decoded literal value.
type Record struct {
	Type  string
	Key   string
	Value string
}

func (Record) isType() {}

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

// UnknownKeywordError is returned when the parser saw an identifier at a
// position where a recognised top-level keyword (`record`) was expected.
type UnknownKeywordError struct {
	Pos Pos
	Got string
}

func (e *UnknownKeywordError) Error() string {
	return fmt.Sprintf("unknown keyword %q at %d:%d", e.Got, e.Pos.Line, e.Pos.Column)
}

// UnknownTypeError is returned when a record declaration uses a type name
// that is not one of the recognised types (currently `string`).
type UnknownTypeError struct {
	Pos Pos
	Got string
}

func (e *UnknownTypeError) Error() string {
	return fmt.Sprintf("unknown record type %q at %d:%d", e.Got, e.Pos.Line, e.Pos.Column)
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

// parseFile is the top-level action. It pulls one token, dispatches on the
// keyword identifier, and returns the appropriate specialised action.
// EOF (nothing more to read) returns (nil, nil) cleanly.
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
	return nil, &UnknownKeywordError{Pos: tok.Pos, Got: tok.Value}
}

// parseRecord runs the inner action loop for a single record statement. Each
// state of the parse — type name, key, equals symbol, value — has its own
// parserAction[*Record]. The completed record is appended to f.Records and
// dispatch returns to parseFile.
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

// parseRecordType expects an identifier naming a recognised type and stores
// it on rec.Type.
func parseRecordType(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	switch tok.Value {
	case "string":
		rec.Type = tok.Value
		return parseRecordKey, nil
	}
	return nil, &UnknownTypeError{Pos: tok.Pos, Got: tok.Value}
}

// parseRecordKey expects the record's key identifier and stores it on rec.Key.
func parseRecordKey(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	rec.Key = tok.Value
	return parseRecordEquals, nil
}

// parseRecordEquals expects the `=` symbol token.
func parseRecordEquals(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return nil, err
	}
	if tok.Value != "=" {
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
	}
	return parseRecordValue, nil
}

// parseRecordValue expects a value token of the type matching rec.Type and
// stores its decoded text on rec.Value. This is the terminal action of the
// record's inner loop — it returns (nil, nil) on success.
func parseRecordValue(p *parser, rec *Record) (parserAction[*Record], error) {
	switch rec.Type {
	case "string":
		tok, err := p.expect(TokenString)
		if err != nil {
			return nil, err
		}
		rec.Value = tok.Value
		return nil, nil
	}
	// Unreachable: parseRecordType already gates on known type names.
	return nil, &UnknownTypeError{Got: rec.Type}
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
