package kvr

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node. It owns top-level records and named blocks.
type File struct {
	Records []Record
	Blocks  []Block
}

// Record is a single typed key-value declaration. LeadingComments holds any
// run of comments that immediately preceded the record in the source.
type Record struct {
	LeadingComments []string
	Type            string
	Key             string
	Value           string
}

func (Record) isType() {}

// Block is a named group of records. LeadingComments holds any run of comments
// that immediately preceded the block in the source.
type Block struct {
	LeadingComments []string
	Name            string
	Records         []Record
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

// UnknownKeywordError is returned when the parser encounters an identifier
// in a position that requires a specific keyword (e.g. top-level statement
// must start with `record` or `block`) and the value did not match.
type UnknownKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnknownKeywordError) Error() string {
	return fmt.Sprintf("unknown keyword %q at %d:%d, expected one of %v",
		e.Got.Value, e.Got.Pos.Line, e.Got.Pos.Column, e.Want)
}

// TypeMismatchError is returned when a record's value-token type does not
// match the declared record type. Pos is the value token's position.
type TypeMismatchError struct {
	Type string
	Got  Token
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: declared %s but value at %d:%d is %s",
		e.Type, e.Got.Pos.Line, e.Got.Pos.Column, e.Got.Type)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
// A one-token lookahead buffer (held) lets actions peek without consuming.
type parser struct {
	pull   func() (Token, error, bool)
	stop   func()
	held   Token
	hasHeld bool
	heldErr error
	heldOk  bool
}

// next returns the next token, consuming any held lookahead first.
func (p *parser) next() (Token, error, bool) {
	if p.hasHeld {
		t, e, ok := p.held, p.heldErr, p.heldOk
		p.hasHeld = false
		p.held = Token{}
		p.heldErr = nil
		p.heldOk = false
		return t, e, ok
	}
	return p.pull()
}

// peek returns the next token without consuming it.
func (p *parser) peek() (Token, error, bool) {
	if p.hasHeld {
		return p.held, p.heldErr, p.heldOk
	}
	t, e, ok := p.pull()
	p.held = t
	p.heldErr = e
	p.heldOk = ok
	p.hasHeld = true
	return t, e, ok
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

// expectSymbol is a convenience for expect(TokenSymbol) plus a value check.
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

// parseFile is the top-level dispatch. It peeks at the next token: an
// identifier whose value is `record` or `block` opens the corresponding
// statement; a comment is buffered as a leading-comment run for the next
// statement; EOF ends the parse.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	var pendingComments []string
	for {
		tok, err, ok := p.peek()
		if !ok {
			// Trailing comments with no following statement are dropped per
			// the spec (only attached comments round-trip). Done.
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		switch {
		case tok.Type == TokenComment:
			// Consume the comment and accumulate it for the next statement.
			_, _, _ = p.next()
			pendingComments = append(pendingComments, tok.Value)
			continue
		case tok.Type == TokenIdentifier && tok.Value == "record":
			return parseTopLevelRecord(pendingComments), nil
		case tok.Type == TokenIdentifier && tok.Value == "block":
			return parseTopLevelBlock(pendingComments), nil
		case tok.Type == TokenIdentifier:
			return nil, &UnknownKeywordError{Got: tok, Want: []string{"record", "block"}}
		default:
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenComment}}
		}
	}
}

// parseTopLevelRecord runs the inner record-action loop and appends the
// resulting Record to f.Records.
func parseTopLevelRecord(leading []string) parserAction[*File] {
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

// parseTopLevelBlock runs the inner block-action loop and appends the
// resulting Block to f.Blocks.
func parseTopLevelBlock(leading []string) parserAction[*File] {
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

// --- Record inner action loop --------------------------------------------------

// parseRecordKeyword consumes the literal `record` identifier.
func parseRecordKeyword(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	if tok.Value != "record" {
		return nil, &UnknownKeywordError{Got: tok, Want: []string{"record"}}
	}
	return parseRecordType, nil
}

// parseRecordType consumes the type-name identifier (`string` or `number`).
func parseRecordType(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	if tok.Value != "string" && tok.Value != "number" {
		return nil, &UnknownKeywordError{Got: tok, Want: []string{"string", "number"}}
	}
	rec.Type = tok.Value
	return parseRecordKey, nil
}

// parseRecordKey consumes the record key identifier.
func parseRecordKey(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	rec.Key = tok.Value
	return parseRecordEquals, nil
}

// parseRecordEquals consumes the `=` symbol.
func parseRecordEquals(p *parser, rec *Record) (parserAction[*Record], error) {
	if _, err := p.expectSymbol("="); err != nil {
		return nil, err
	}
	return parseRecordValue, nil
}

// parseRecordValue consumes the value token and verifies its type matches the
// declared record type.
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

// --- Block inner action loop ---------------------------------------------------

// parseBlockKeyword consumes the literal `block` identifier.
func parseBlockKeyword(p *parser, blk *Block) (parserAction[*Block], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	if tok.Value != "block" {
		return nil, &UnknownKeywordError{Got: tok, Want: []string{"block"}}
	}
	return parseBlockName, nil
}

// parseBlockName consumes the block-name identifier.
func parseBlockName(p *parser, blk *Block) (parserAction[*Block], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	blk.Name = tok.Value
	return parseBlockOpen, nil
}

// parseBlockOpen consumes the `{` symbol.
func parseBlockOpen(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	return parseBlockMember, nil
}

// parseBlockMember peeks at the next token. If it's `}`, the block ends
// (parseBlockClose handles it). Otherwise it parses one inner record, then
// requires a `;`, then loops back to parseBlockMember.
func parseBlockMember(p *parser, blk *Block) (parserAction[*Block], error) {
	tok, err, ok := p.peek()
	if !ok {
		return nil, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return nil, err
	}
	if tok.Type == TokenSymbol && tok.Value == "}" {
		return parseBlockClose, nil
	}
	// Inner statements must be records (the spec's inner Statement reduces
	// to Record for now; comments can be added without changing this shape).
	rec := &Record{}
	var rerr error
	for action := parseRecordKeyword; action != nil && rerr == nil; {
		action, rerr = action(p, rec)
	}
	if rerr != nil {
		return nil, rerr
	}
	blk.Records = append(blk.Records, *rec)
	return parseBlockSeparator, nil
}

// parseBlockSeparator consumes the `;` that follows every inner statement.
func parseBlockSeparator(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol(";"); err != nil {
		return nil, err
	}
	return parseBlockMember, nil
}

// parseBlockClose consumes the `}` symbol and ends the block.
func parseBlockClose(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol("}"); err != nil {
		return nil, err
	}
	return nil, nil
}

// Parse reads a KVR file from r.
func Parse(r io.Reader) (*File, error) {
	pull, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{
		pull: pull,
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
