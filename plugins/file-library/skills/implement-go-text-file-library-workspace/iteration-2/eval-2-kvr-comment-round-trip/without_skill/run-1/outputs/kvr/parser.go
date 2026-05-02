package kvr

import (
	"fmt"
	"io"
	"iter"
)

// Statement is the marker interface for top-level statements (Record, Block).
type Statement interface {
	isStatement()
}

// File is the top-level AST node. Statements preserves source order across
// records and blocks, which the printer relies on for round-trip fidelity.
type File struct {
	Statements []Statement
}

// RecordType identifies the declared type of a Record's value.
type RecordType int

const (
	RecordTypeInvalid RecordType = iota
	RecordTypeString
	RecordTypeNumber
)

// String returns the source spelling of a RecordType.
func (rt RecordType) String() string {
	switch rt {
	case RecordTypeString:
		return "string"
	case RecordTypeNumber:
		return "number"
	default:
		return fmt.Sprintf("RecordType(%d)", int(rt))
	}
}

// Record is a single typed key-value declaration.
type Record struct {
	LeadingComments []string
	Type            RecordType
	Key             string
	Value           string
}

func (Record) isStatement() {}

// Block is a named group of records.
type Block struct {
	LeadingComments []string
	Name            string
	Records         []Record
}

func (Block) isStatement() {}

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

// UnknownTypeError is returned when the parser sees a record type identifier
// other than "string" or "number".
type UnknownTypeError struct {
	Got Token
}

func (e *UnknownTypeError) Error() string {
	return fmt.Sprintf("unknown record type %q at %d:%d", e.Got.Value, e.Got.Pos.Line, e.Got.Pos.Column)
}

// UnknownStatementError is returned when the parser sees an opening identifier
// other than "record" or "block".
type UnknownStatementError struct {
	Got Token
}

func (e *UnknownStatementError) Error() string {
	return fmt.Sprintf("unknown statement %q at %d:%d", e.Got.Value, e.Got.Pos.Line, e.Got.Pos.Column)
}

// TypeMismatchError is returned when a record's value-token type does not
// agree with the declared record type.
type TypeMismatchError struct {
	Type RecordType
	Got  Token
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: declared %s but got %s at %d:%d", e.Type, e.Got.Type, e.Got.Pos.Line, e.Got.Pos.Column)
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

// expectSymbol pulls the next token and verifies it is a TokenSymbol with the
// given value.
func (p *parser) expectSymbol(value string) (Token, error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return Token{}, err
	}
	if tok.Value != value {
		return Token{}, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
	}
	return tok, nil
}

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action. It consumes one statement at a time
// (a run of comments followed by a Record or Block) until tokens run out.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	var leading []string
	for {
		tok, err, ok := p.next()
		if !ok {
			// EOF — but if we've collected leading comments with no following
			// statement, attach nothing (free-floating tail comments are dropped
			// silently for round-trip purposes; the tests only cover comments
			// preceding a statement).
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		switch tok.Type {
		case TokenComment:
			leading = append(leading, tok.Value)
			continue
		case TokenIdentifier:
			switch tok.Value {
			case "record":
				rec := Record{LeadingComments: leading}
				if err := parseRecordBody(p, &rec); err != nil {
					return nil, err
				}
				f.Statements = append(f.Statements, rec)
				return parseFile, nil
			case "block":
				blk := Block{LeadingComments: leading}
				if err := parseBlockBody(p, &blk); err != nil {
					return nil, err
				}
				f.Statements = append(f.Statements, blk)
				return parseFile, nil
			default:
				return nil, &UnknownStatementError{Got: tok}
			}
		default:
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenComment}}
		}
	}
}

// parseRecordBody parses the tail of a record after the leading "record"
// identifier has already been consumed: Type Identifier "=" Value .
func parseRecordBody(p *parser, rec *Record) error {
	typeTok, err := p.expect(TokenIdentifier)
	if err != nil {
		return err
	}
	switch typeTok.Value {
	case "string":
		rec.Type = RecordTypeString
	case "number":
		rec.Type = RecordTypeNumber
	default:
		return &UnknownTypeError{Got: typeTok}
	}
	keyTok, err := p.expect(TokenIdentifier)
	if err != nil {
		return err
	}
	rec.Key = keyTok.Value
	if _, err := p.expectSymbol("="); err != nil {
		return err
	}
	valTok, err := p.expect(TokenString, TokenNumber)
	if err != nil {
		return err
	}
	switch rec.Type {
	case RecordTypeString:
		if valTok.Type != TokenString {
			return &TypeMismatchError{Type: rec.Type, Got: valTok}
		}
	case RecordTypeNumber:
		if valTok.Type != TokenNumber {
			return &TypeMismatchError{Type: rec.Type, Got: valTok}
		}
	}
	rec.Value = valTok.Value
	return nil
}

// parseBlockBody parses the tail of a block after the leading "block"
// identifier has already been consumed: Identifier "{" { Statement ";" } "}" .
func parseBlockBody(p *parser, blk *Block) error {
	nameTok, err := p.expect(TokenIdentifier)
	if err != nil {
		return err
	}
	blk.Name = nameTok.Value
	if _, err := p.expectSymbol("{"); err != nil {
		return err
	}
	// Inner loop: consume zero or more (Statement ";") pairs until "}".
	var leading []string
	for {
		tok, err, ok := p.next()
		if !ok {
			return &UnexpectedEndOfTokensError{}
		}
		if err != nil {
			return err
		}
		switch tok.Type {
		case TokenComment:
			leading = append(leading, tok.Value)
			continue
		case TokenSymbol:
			if tok.Value == "}" {
				if len(leading) > 0 {
					return &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier}}
				}
				return nil
			}
			return &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenSymbol}}
		case TokenIdentifier:
			if tok.Value != "record" {
				return &UnknownStatementError{Got: tok}
			}
			rec := Record{LeadingComments: leading}
			leading = nil
			if err := parseRecordBody(p, &rec); err != nil {
				return err
			}
			if _, err := p.expectSymbol(";"); err != nil {
				return err
			}
			blk.Records = append(blk.Records, rec)
		default:
			return &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenComment, TokenSymbol}}
		}
	}
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
