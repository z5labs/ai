package kvr

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node. It carries top-level records and named
// blocks (each of which carries its own records).
type File struct {
	Records []Record
	Blocks  []Block
}

// Record is a single typed key-value declaration. Its concrete shape is the
// same whether it appears at top level or inside a block.
type Record struct {
	Pos   Pos
	Type  string
	Key   string
	Value string
}

func (Record) isType() {}

// Block is a named group of records: `block NAME { record ... ; record ... ; }`.
type Block struct {
	Pos     Pos
	Name    string
	Records []Record
}

func (Block) isType() {}

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

// TypeMismatchError is returned when a record's declared type and value-token
// type disagree (e.g. `record string K = 42`).
type TypeMismatchError struct {
	Pos  Pos
	Type string
	Got  TokenType
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: declared %s but got %s at %d:%d",
		e.Type, e.Got, e.Pos.Line, e.Pos.Column)
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

// expectSymbol pulls the next token, requires TokenSymbol, and additionally
// checks the symbol value matches one of values. Returns UnexpectedTokenError
// if the type or value does not match.
func (p *parser) expectSymbol(values ...string) (Token, error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return Token{}, err
	}
	for _, v := range values {
		if tok.Value == v {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
}

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level dispatch action. It peeks the next identifier
// token and dispatches on its value: `record` → parseRecord, `block` →
// parseBlock. End-of-tokens completes parsing.
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
		return parseTopLevelRecord(tok.Pos), nil
	case "block":
		return parseBlock(tok.Pos), nil
	}
	return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier}}
}

// parseTopLevelRecord parses a record at the top level (no trailing `;`) and
// appends it to the File's Records slice. The opening `record` keyword has
// already been consumed; startPos is its position.
func parseTopLevelRecord(startPos Pos) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		rec, err := parseRecordBody(p, startPos)
		if err != nil {
			return nil, err
		}
		f.Records = append(f.Records, rec)
		return parseFile, nil
	}
}

// parseRecordBody parses everything after the `record` keyword: type,
// identifier, `=`, value. It is shared between top-level records and
// block-internal records — only the trailing terminator differs and the
// caller handles that.
func parseRecordBody(p *parser, startPos Pos) (Record, error) {
	typeTok, err := p.expect(TokenIdentifier)
	if err != nil {
		return Record{}, err
	}
	if typeTok.Value != "string" && typeTok.Value != "number" {
		return Record{}, &UnexpectedTokenError{Got: typeTok, Want: []TokenType{TokenIdentifier}}
	}
	keyTok, err := p.expect(TokenIdentifier)
	if err != nil {
		return Record{}, err
	}
	if _, err := p.expectSymbol("="); err != nil {
		return Record{}, err
	}
	valTok, err := p.expect(TokenString, TokenNumber)
	if err != nil {
		return Record{}, err
	}
	switch typeTok.Value {
	case "string":
		if valTok.Type != TokenString {
			return Record{}, &TypeMismatchError{Pos: valTok.Pos, Type: typeTok.Value, Got: valTok.Type}
		}
	case "number":
		if valTok.Type != TokenNumber {
			return Record{}, &TypeMismatchError{Pos: valTok.Pos, Type: typeTok.Value, Got: valTok.Type}
		}
	}
	return Record{
		Pos:   startPos,
		Type:  typeTok.Value,
		Key:   keyTok.Value,
		Value: valTok.Value,
	}, nil
}

// parseBlock parses a `block NAME { ... }` construct at the top level. It uses
// the inner action loop pattern: each state of the block parse — the open
// brace, each member, the separator, and the close brace — gets its own
// parserAction[*Block]. The opening `block` keyword has already been consumed;
// startPos is its position.
func parseBlock(startPos Pos) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		nameTok, err := p.expect(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		blk := &Block{Pos: startPos, Name: nameTok.Value}
		for action := parseBlockOpen; action != nil; {
			action, err = action(p, blk)
			if err != nil {
				return nil, err
			}
		}
		f.Blocks = append(f.Blocks, *blk)
		return parseFile, nil
	}
}

// parseBlockOpen consumes the opening `{` and dispatches to either
// parseBlockClose (empty block) or parseBlockMember.
func parseBlockOpen(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	return parseBlockMemberOrClose, nil
}

// parseBlockMemberOrClose peeks the next token: `}` ends the block, `record`
// starts a new member, anything else is an error.
func parseBlockMemberOrClose(p *parser, blk *Block) (parserAction[*Block], error) {
	tok, err, ok := p.next()
	if !ok {
		return nil, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return nil, err
	}
	if tok.Type == TokenSymbol && tok.Value == "}" {
		return nil, nil
	}
	if tok.Type == TokenIdentifier && tok.Value == "record" {
		return parseBlockMember(tok.Pos), nil
	}
	return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol, TokenIdentifier}}
}

// parseBlockMember parses one record body inside a block. The opening
// `record` keyword has already been consumed; startPos is its position.
// After the record, expects the `;` separator and chains to
// parseBlockMemberOrClose.
func parseBlockMember(startPos Pos) parserAction[*Block] {
	return func(p *parser, blk *Block) (parserAction[*Block], error) {
		rec, err := parseRecordBody(p, startPos)
		if err != nil {
			return nil, err
		}
		blk.Records = append(blk.Records, rec)
		return parseBlockSeparator, nil
	}
}

// parseBlockSeparator consumes the mandatory `;` between (and after) inner
// statements, then chains to parseBlockMemberOrClose.
func parseBlockSeparator(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol(";"); err != nil {
		return nil, err
	}
	return parseBlockMemberOrClose, nil
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
