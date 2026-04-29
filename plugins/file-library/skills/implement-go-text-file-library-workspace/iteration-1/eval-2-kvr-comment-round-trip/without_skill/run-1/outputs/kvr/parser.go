package kvr

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node.
type File struct {
	Statements []Statement
}

// Statement is the marker interface implemented by Record and Block.
type Statement interface {
	isStatement()
}

func (Record) isStatement() {}
func (Block) isStatement()  {}

// Record is a single typed key-value declaration.
type Record struct {
	LeadingComments []string
	Type            string // "string" | "number"
	Key             string
	Value           string // decoded value text; type interprets it
}

// Block is a named group of records.
type Block struct {
	LeadingComments []string
	Name            string
	Records         []Record
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

// UnexpectedKeywordError is returned when the parser sees an identifier where
// a specific keyword (like "record" or "block") was required.
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %s, expected one of %v", e.Got, e.Want)
}

// TypeMismatchError is returned when a record's declared type does not match
// the kind of literal that follows the `=`.
type TypeMismatchError struct {
	Type string
	Got  Token
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: declared %q but value was %s", e.Type, e.Got)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
type parser struct {
	next func() (Token, error, bool)
	stop func()
	// peeked token (if any) that a previous action over-read and pushed back.
	peeked    Token
	hasPeeked bool
	peekedEnd bool
}

// pull returns the next token. If something was peeked, it is returned first.
// The third return is false when the stream has ended.
func (p *parser) pull() (Token, error, bool) {
	if p.hasPeeked {
		tok := p.peeked
		p.hasPeeked = false
		if p.peekedEnd {
			p.peekedEnd = false
			return Token{}, nil, false
		}
		return tok, nil, true
	}
	return p.next()
}

// pushBack stashes a token to be returned by the next pull call.
func (p *parser) pushBack(tok Token) {
	p.peeked = tok
	p.hasPeeked = true
	p.peekedEnd = false
}

// expect pulls the next token and verifies its type matches one of the given
// types. Use it everywhere the grammar requires a specific token; never
// inline the type check.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.pull()
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

// expectIdentifierValue pulls the next token, requires it to be an identifier
// whose value is in the given set, and returns it.
func (p *parser) expectIdentifierValue(values ...string) (Token, error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return Token{}, err
	}
	for _, v := range values {
		if tok.Value == v {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedKeywordError{Got: tok, Want: values}
}

// expectSymbolValue pulls the next token, requires it to be a symbol with the
// given character, and returns it.
func (p *parser) expectSymbolValue(value string) (Token, error) {
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

// parseFile is the top-level action. It dispatches on the next token: comments
// accumulate into a leading-comments buffer that gets attached to the next
// record or block. End-of-stream completes the parse cleanly.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	return parseFileStatement(nil)(p, f)
}

// parseFileStatement returns an action that consumes any pending leading
// comments plus a record/block opener and dispatches to the statement parser.
func parseFileStatement(leading []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		tok, err, ok := p.pull()
		if !ok {
			// EOF after some leading comments turns them into bare comment
			// statements? The spec says comment is a valid statement on its
			// own, but only block/record carry LeadingComments. Trailing
			// dangling comments at EOF are lost across round-trip — same
			// rule as blank lines per Semantics. We simply finish.
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		switch {
		case tok.Type == TokenComment:
			return parseFileStatement(append(leading, tok.Value)), nil
		case tok.Type == TokenIdentifier && tok.Value == "record":
			rec, err := parseRecordBody(p, leading)
			if err != nil {
				return nil, err
			}
			f.Statements = append(f.Statements, rec)
			return parseFileStatement(nil), nil
		case tok.Type == TokenIdentifier && tok.Value == "block":
			blk, err := parseBlockBody(p, leading)
			if err != nil {
				return nil, err
			}
			f.Statements = append(f.Statements, blk)
			return parseFileStatement(nil), nil
		default:
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenComment, TokenIdentifier}}
		}
	}
}

// parseRecordBody parses everything after the literal `record` keyword has
// already been consumed: Type Identifier "=" Value. The leading comments are
// attached to the resulting Record.
func parseRecordBody(p *parser, leading []string) (Record, error) {
	rec := &Record{LeadingComments: leading}
	var err error
	for action := parseRecordType; action != nil && err == nil; {
		action, err = action(p, rec)
	}
	if err != nil {
		return Record{}, err
	}
	return *rec, nil
}

func parseRecordType(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expectIdentifierValue("string", "number")
	if err != nil {
		return nil, err
	}
	r.Type = tok.Value
	return parseRecordKey, nil
}

func parseRecordKey(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	r.Key = tok.Value
	return parseRecordEquals, nil
}

func parseRecordEquals(p *parser, r *Record) (parserAction[*Record], error) {
	if _, err := p.expectSymbolValue("="); err != nil {
		return nil, err
	}
	return parseRecordValue, nil
}

func parseRecordValue(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenString, TokenNumber)
	if err != nil {
		return nil, err
	}
	switch r.Type {
	case "string":
		if tok.Type != TokenString {
			return nil, &TypeMismatchError{Type: r.Type, Got: tok}
		}
	case "number":
		if tok.Type != TokenNumber {
			return nil, &TypeMismatchError{Type: r.Type, Got: tok}
		}
	}
	r.Value = tok.Value
	return nil, nil
}

// parseBlockBody parses everything after the literal `block` keyword has been
// consumed: Identifier "{" { Statement ";" } "}". Leading comments attach to
// the block.
func parseBlockBody(p *parser, leading []string) (Block, error) {
	blk := &Block{LeadingComments: leading}
	var err error
	for action := parseBlockName; action != nil && err == nil; {
		action, err = action(p, blk)
	}
	if err != nil {
		return Block{}, err
	}
	return *blk, nil
}

func parseBlockName(p *parser, b *Block) (parserAction[*Block], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	b.Name = tok.Value
	return parseBlockOpen, nil
}

func parseBlockOpen(p *parser, b *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbolValue("{"); err != nil {
		return nil, err
	}
	return parseBlockStatement(nil), nil
}

// parseBlockStatement consumes any leading comments and either dispatches to a
// nested record parse (followed by the required `;`) or to the block close.
func parseBlockStatement(leading []string) parserAction[*Block] {
	return func(p *parser, b *Block) (parserAction[*Block], error) {
		tok, err, ok := p.pull()
		if !ok {
			return nil, &UnexpectedEndOfTokensError{}
		}
		if err != nil {
			return nil, err
		}
		switch {
		case tok.Type == TokenComment:
			return parseBlockStatement(append(leading, tok.Value)), nil
		case tok.Type == TokenSymbol && tok.Value == "}":
			if len(leading) > 0 {
				// Dangling comments before close-brace — drop them; they do
				// not survive round-trip per the attachment rule.
				_ = leading
			}
			return nil, nil
		case tok.Type == TokenIdentifier && tok.Value == "record":
			rec, err := parseRecordBody(p, leading)
			if err != nil {
				return nil, err
			}
			if _, err := p.expectSymbolValue(";"); err != nil {
				return nil, err
			}
			b.Records = append(b.Records, rec)
			return parseBlockStatement(nil), nil
		default:
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenComment, TokenIdentifier, TokenSymbol}}
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
