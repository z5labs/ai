package kvr

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node — a sequence of Records and Blocks in source
// order. (For comment round-trip, leading comments are attached to the next
// Record / Block via LeadingComments.)
type File struct {
	Records []Record
	Blocks  []Block
}

// Type is the marker interface for AST node types in the KVR AST.
// Concrete node types implement isType() to satisfy this interface.
type Type interface {
	isType()
}

// Record is a single typed key-value declaration.
type Record struct {
	// LeadingComments holds comment-token Values (without the leading '#' or
	// stripped horizontal whitespace) attached to this record from the
	// statements immediately preceding it. Order is source order.
	LeadingComments []string
	Type            string // "string" or "number"
	Key             string
	Value           string
}

func (Record) isType() {}

// Block is a named group of records.
type Block struct {
	// LeadingComments mirrors Record.LeadingComments.
	LeadingComments []string
	Name            string
	Records         []Record
}

func (Block) isType() {}

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

// UnexpectedKeywordError is returned when the parser saw an identifier whose
// value is not a valid keyword in the current grammar position.
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %s, expected one of %v", e.Got, e.Want)
}

// TypeMismatchError is returned when a record value's token type does not
// match the declared record type (e.g. a number value on a string-typed
// record).
type TypeMismatchError struct {
	Type string // declared type name ("string" or "number")
	Got  Token  // the value token whose type didn't match
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: record declared as %q but got %s", e.Type, e.Got)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
// One token of lookahead is buffered via peek/peeked* fields so dispatch
// actions can decide which sub-parser to launch without consuming the token.
type parser struct {
	next func() (Token, error, bool)
	stop func()

	hasPeek   bool
	peeked    Token
	peekedErr error
	peekedOk  bool
}

// expect pulls the next token and verifies its type matches one of the given
// types. Use it everywhere the grammar requires a specific token; never
// inline the type check.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.advance()
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

// advance pulls the next token, honouring any buffered peek.
func (p *parser) advance() (Token, error, bool) {
	if p.hasPeek {
		tok, err, ok := p.peeked, p.peekedErr, p.peekedOk
		p.hasPeek = false
		return tok, err, ok
	}
	return p.next()
}

// peek returns the next token without consuming it. The token is buffered
// and returned again by the next advance/expect call.
func (p *parser) peek() (Token, error, bool) {
	if p.hasPeek {
		return p.peeked, p.peekedErr, p.peekedOk
	}
	tok, err, ok := p.next()
	p.peeked = tok
	p.peekedErr = err
	p.peekedOk = ok
	p.hasPeek = true
	return tok, err, ok
}

// expectIdentValue pulls the next token, requires it to be a TokenIdentifier
// with one of the given values.
func (p *parser) expectIdentValue(values ...string) (Token, error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return Token{}, err
	}
	for _, want := range values {
		if tok.Value == want {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedKeywordError{Got: tok, Want: values}
}

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action. It dispatches to record/block/comment
// handlers based on the next token's value. EOF returns (nil, nil).
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	tok, err, ok := p.peek()
	if !ok {
		// EOF — nothing more to read.
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case TokenComment:
		return parseLeadingTrivia, nil
	case TokenIdentifier:
		switch tok.Value {
		case "record":
			return parseRecordStatement(nil), nil
		case "block":
			return parseBlockStatement(nil), nil
		}
		return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "block"}}
	}
	return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenComment, TokenIdentifier}}
}

// parseLeadingTrivia consumes a run of TokenComment values, accumulating them,
// then dispatches to the record or block parser passing the comments as
// leading trivia.
func parseLeadingTrivia(p *parser, f *File) (parserAction[*File], error) {
	var comments []string
	for {
		tok, err, ok := p.peek()
		if !ok {
			// EOF after a run of free-floating comments — drop them. (The
			// spec says comments attach to the immediately following
			// non-comment statement; without one, they have nowhere to live.)
			// Discard via consuming the peek then return.
			_, _, _ = p.advance()
			if len(comments) > 0 {
				// No following statement — the comments are orphaned.
				// Per the printer round-trip contract, the parser must drop
				// them rather than carry them on a phantom statement.
				return nil, nil
			}
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		if tok.Type != TokenComment {
			break
		}
		_, _, _ = p.advance()
		comments = append(comments, tok.Value)
	}
	// Now peek at the dispatch token. parseLeadingTrivia is only entered with
	// at least one buffered comment, but defensive: if the next token is
	// unexpected, surface UnexpectedTokenError.
	tok, err, ok := p.peek()
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
		return parseRecordStatement(comments), nil
	case "block":
		return parseBlockStatement(comments), nil
	}
	return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "block"}}
}

// parseRecordStatement parses one top-level record using the inner action
// loop pattern. The closure captures any leading-trivia comments accumulated
// upstream.
func parseRecordStatement(leading []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		rec := &Record{LeadingComments: leading}
		var err error
		for action := parseRecordKeyword; action != nil && err == nil; {
			action, err = action(p, rec)
		}
		if err != nil {
			return nil, err
		}
		f.Records = append(f.Records, *rec)
		return parseFile, nil
	}
}

// parseRecordKeyword consumes the literal "record" identifier.
func parseRecordKeyword(p *parser, rec *Record) (parserAction[*Record], error) {
	if _, err := p.expectIdentValue("record"); err != nil {
		return nil, err
	}
	return parseRecordType, nil
}

// parseRecordType consumes the type name ("string" or "number").
func parseRecordType(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expectIdentValue("string", "number")
	if err != nil {
		return nil, err
	}
	rec.Type = tok.Value
	return parseRecordKey, nil
}

// parseRecordKey consumes the key identifier.
func parseRecordKey(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	rec.Key = tok.Value
	return parseRecordEquals, nil
}

// parseRecordEquals consumes the '=' symbol.
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

// parseRecordValue consumes the typed value, enforcing type/value agreement.
func parseRecordValue(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenString, TokenNumber)
	if err != nil {
		return nil, err
	}
	switch rec.Type {
	case "string":
		if tok.Type != TokenString {
			return nil, &TypeMismatchError{Type: rec.Type, Got: tok}
		}
	case "number":
		if tok.Type != TokenNumber {
			return nil, &TypeMismatchError{Type: rec.Type, Got: tok}
		}
	}
	rec.Value = tok.Value
	return nil, nil
}

// parseBlockStatement parses one top-level block using the inner action loop
// pattern.
func parseBlockStatement(leading []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		blk := &Block{LeadingComments: leading}
		var err error
		for action := parseBlockKeyword; action != nil && err == nil; {
			action, err = action(p, blk)
		}
		if err != nil {
			return nil, err
		}
		f.Blocks = append(f.Blocks, *blk)
		return parseFile, nil
	}
}

// parseBlockKeyword consumes the literal "block" identifier.
func parseBlockKeyword(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectIdentValue("block"); err != nil {
		return nil, err
	}
	return parseBlockName, nil
}

// parseBlockName consumes the block name identifier.
func parseBlockName(p *parser, blk *Block) (parserAction[*Block], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	blk.Name = tok.Value
	return parseBlockOpen, nil
}

// parseBlockOpen consumes the '{' symbol.
func parseBlockOpen(p *parser, blk *Block) (parserAction[*Block], error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return nil, err
	}
	if tok.Value != "{" {
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
	}
	return parseBlockBody, nil
}

// parseBlockBody parses zero-or-more inner statements, each terminated by
// ';', then consumes the closing '}'. Inner statements can be records with
// optional leading comments.
func parseBlockBody(p *parser, blk *Block) (parserAction[*Block], error) {
	for {
		tok, err, ok := p.peek()
		if !ok {
			return nil, &UnexpectedEndOfTokensError{}
		}
		if err != nil {
			return nil, err
		}
		// Block-close ends the body.
		if tok.Type == TokenSymbol && tok.Value == "}" {
			_, _, _ = p.advance()
			return nil, nil
		}
		// Otherwise expect a (possibly comment-prefixed) inner record.
		var inner []string
		for tok.Type == TokenComment {
			_, _, _ = p.advance()
			inner = append(inner, tok.Value)
			tok, err, ok = p.peek()
			if !ok {
				return nil, &UnexpectedEndOfTokensError{}
			}
			if err != nil {
				return nil, err
			}
		}
		if tok.Type != TokenIdentifier || tok.Value != "record" {
			return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record"}}
		}
		rec := &Record{LeadingComments: inner}
		var aerr error
		for action := parseRecordKeyword; action != nil && aerr == nil; {
			action, aerr = action(p, rec)
		}
		if aerr != nil {
			return nil, aerr
		}
		// Trailing ';' is required after every inner statement, including the
		// last one before '}'.
		semi, err := p.expect(TokenSymbol)
		if err != nil {
			return nil, err
		}
		if semi.Value != ";" {
			return nil, &UnexpectedTokenError{Got: semi, Want: []TokenType{TokenSymbol}}
		}
		blk.Records = append(blk.Records, *rec)
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
