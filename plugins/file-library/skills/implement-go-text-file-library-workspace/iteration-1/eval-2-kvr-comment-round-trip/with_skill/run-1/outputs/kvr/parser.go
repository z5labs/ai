package kvr

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node. Its Statements slice carries Records and
// Blocks in the order they appeared in the source.
type File struct {
	Statements []Type
}

// Type is the marker interface for AST node types in the KVR AST.
// Concrete node types implement isType() to satisfy this interface.
type Type interface {
	isType()
}

// Record is a single typed key-value declaration.
//
// LeadingComments holds the run of comments that immediately preceded this
// record in the source (in order). They survive a round-trip through Print.
// Type is "string" or "number". Value is the decoded string content (for
// string records) or the digit text (for number records). ValueKind is the
// token kind that produced Value, used by the printer to re-quote strings.
type Record struct {
	LeadingComments []string
	Type            string
	Key             string
	Value           string
	ValueKind       TokenType
}

func (*Record) isType() {}

// Block is a named group of records.
//
// LeadingComments holds the run of comments that immediately preceded this
// block. Records carries the inner records (each may have its own
// LeadingComments).
type Block struct {
	LeadingComments []string
	Name            string
	Records         []Record
}

func (*Block) isType() {}

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

// UnexpectedKeywordError is returned when the parser saw an identifier
// that is not one of the expected keywords for the current position.
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %q at %d:%d, expected one of %v",
		e.Got.Value, e.Got.Pos.Line, e.Got.Pos.Column, e.Want)
}

// TypeMismatchError is returned when a record value's token kind does not
// match the declared type (e.g. `record string K = 42`).
type TypeMismatchError struct {
	Type string // declared type ("string" or "number")
	Got  Token  // value token whose kind disagreed
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: declared %q but value %s at %d:%d",
		e.Type, e.Got.Type, e.Got.Pos.Line, e.Got.Pos.Column)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
type parser struct {
	pull func() (Token, error, bool)
	stop func()

	// peeked, if hasPeek, is returned by the next call to pullToken instead
	// of advancing the underlying iterator. This is the parser's one-token
	// lookahead, used wherever the grammar needs to inspect a token without
	// committing to a specific action.
	peeked   Token
	peekErr  error
	peekOk   bool
	hasPeek  bool
}

// pullToken returns the next (token, err, ok) triple, honouring any pushed-
// back peek.
func (p *parser) pullToken() (Token, error, bool) {
	if p.hasPeek {
		t, e, ok := p.peeked, p.peekErr, p.peekOk
		p.peeked, p.peekErr, p.peekOk, p.hasPeek = Token{}, nil, false, false
		return t, e, ok
	}
	return p.pull()
}

// pushBack stores tok/err/ok so the next pullToken returns them.
func (p *parser) pushBack(tok Token, err error, ok bool) {
	p.peeked, p.peekErr, p.peekOk, p.hasPeek = tok, err, ok, true
}

// peek returns the next token without consuming it.
func (p *parser) peek() (Token, error, bool) {
	tok, err, ok := p.pullToken()
	p.pushBack(tok, err, ok)
	return tok, err, ok
}

// expect pulls the next token and verifies its type matches one of the given
// types. Use it everywhere the grammar requires a specific token; never
// inline the type check.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.pullToken()
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
// the value to be one of the given keywords.
func (p *parser) expectKeyword(keywords ...string) (Token, error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return Token{}, err
	}
	for _, k := range keywords {
		if tok.Value == k {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedKeywordError{Got: tok, Want: keywords}
}

// expectSymbol pulls the next token, requires TokenSymbol, and requires the
// value to equal sym.
func (p *parser) expectSymbol(sym string) (Token, error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return Token{}, err
	}
	if tok.Value != sym {
		return Token{}, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
	}
	return tok, nil
}

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action. It collects pending leading comments,
// then dispatches to parseRecord or parseBlock when it sees the matching
// keyword. End-of-tokens completes successfully.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	tok, err, ok := p.pullToken()
	if !ok {
		// trailing free-floating comments are dropped — the user task only
		// requires comments-above-statement to round-trip.
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// gather a run of leading comments
	var leading []string
	for tok.Type == TokenComment {
		leading = append(leading, tok.Value)
		tok, err, ok = p.pullToken()
		if !ok {
			// orphaned trailing comments — drop per simplification above.
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
	}

	if tok.Type != TokenIdentifier {
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier}}
	}
	switch tok.Value {
	case "record":
		return parseRecordWithLeading(leading), nil
	case "block":
		return parseBlockWithLeading(leading), nil
	}
	return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "block"}}
}

// recordParseState is the state passed between actions of the record inner
// loop. It accumulates the record fields as each action runs.
type recordParseState struct {
	rec *Record
}

// parseRecordWithLeading returns a parserAction[*File] that parses one
// record, attaches the given leading comments, appends the record to f,
// and returns control to parseFile.
//
// The `record` keyword has already been consumed before this action runs.
func parseRecordWithLeading(leading []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		state := &recordParseState{rec: &Record{LeadingComments: leading}}
		var err error
		for action := parseRecordType; action != nil && err == nil; {
			action, err = action(p, state)
		}
		if err != nil {
			return nil, err
		}
		f.Statements = append(f.Statements, state.rec)
		return parseFile, nil
	}
}

func parseRecordType(p *parser, s *recordParseState) (parserAction[*recordParseState], error) {
	tok, err := p.expectKeyword("string", "number")
	if err != nil {
		return nil, err
	}
	s.rec.Type = tok.Value
	return parseRecordKey, nil
}

func parseRecordKey(p *parser, s *recordParseState) (parserAction[*recordParseState], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	s.rec.Key = tok.Value
	return parseRecordEquals, nil
}

func parseRecordEquals(p *parser, s *recordParseState) (parserAction[*recordParseState], error) {
	if _, err := p.expectSymbol("="); err != nil {
		return nil, err
	}
	return parseRecordValue, nil
}

func parseRecordValue(p *parser, s *recordParseState) (parserAction[*recordParseState], error) {
	tok, err := p.expect(TokenString, TokenNumber)
	if err != nil {
		return nil, err
	}
	switch s.rec.Type {
	case "string":
		if tok.Type != TokenString {
			return nil, &TypeMismatchError{Type: s.rec.Type, Got: tok}
		}
	case "number":
		if tok.Type != TokenNumber {
			return nil, &TypeMismatchError{Type: s.rec.Type, Got: tok}
		}
	}
	s.rec.Value = tok.Value
	s.rec.ValueKind = tok.Type
	return nil, nil
}

// blockParseState is the state passed between actions of the block inner
// loop.
type blockParseState struct {
	blk *Block
}

// parseBlockWithLeading returns a parserAction[*File] that parses one block.
// The `block` keyword has already been consumed.
func parseBlockWithLeading(leading []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		state := &blockParseState{blk: &Block{LeadingComments: leading}}
		var err error
		for action := parseBlockName; action != nil && err == nil; {
			action, err = action(p, state)
		}
		if err != nil {
			return nil, err
		}
		f.Statements = append(f.Statements, state.blk)
		return parseFile, nil
	}
}

func parseBlockName(p *parser, s *blockParseState) (parserAction[*blockParseState], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	s.blk.Name = tok.Value
	return parseBlockOpen, nil
}

func parseBlockOpen(p *parser, s *blockParseState) (parserAction[*blockParseState], error) {
	if _, err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	return parseBlockBody, nil
}

// parseBlockBody parses zero or more inner statements, each terminated by `;`.
// It returns nil when the closing `}` is consumed.
func parseBlockBody(p *parser, s *blockParseState) (parserAction[*blockParseState], error) {
	tok, err, ok := p.pullToken()
	if !ok {
		return nil, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return nil, err
	}

	// gather leading comments for the next inner statement
	var leading []string
	for tok.Type == TokenComment {
		leading = append(leading, tok.Value)
		tok, err, ok = p.pullToken()
		if !ok {
			return nil, &UnexpectedEndOfTokensError{}
		}
		if err != nil {
			return nil, err
		}
	}

	// closing `}` ends the block. If we collected leading comments, drop them
	// (no statement to attach to inside the block).
	if tok.Type == TokenSymbol && tok.Value == "}" {
		return nil, nil
	}

	if tok.Type != TokenIdentifier || tok.Value != "record" {
		return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record"}}
	}

	rec, err := parseInnerRecord(p, leading)
	if err != nil {
		return nil, err
	}
	if _, err := p.expectSymbol(";"); err != nil {
		return nil, err
	}
	s.blk.Records = append(s.blk.Records, *rec)
	return parseBlockBody, nil
}

// parseInnerRecord parses a record inside a block. The `record` keyword has
// already been consumed. Leading comments collected by the caller are
// attached.
func parseInnerRecord(p *parser, leading []string) (*Record, error) {
	state := &recordParseState{rec: &Record{LeadingComments: leading}}
	var err error
	for action := parseRecordType; action != nil && err == nil; {
		action, err = action(p, state)
	}
	if err != nil {
		return nil, err
	}
	return state.rec, nil
}

// Parse reads a KVR file from r.
func Parse(r io.Reader) (*File, error) {
	next, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{
		pull: func() (Token, error, bool) {
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
