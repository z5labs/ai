package kvr

import (
	"fmt"
	"io"
	"iter"
)

// RecordType identifies the declared type of a record's value.
type RecordType int

const (
	RecordTypeInvalid RecordType = iota
	RecordTypeString
	RecordTypeNumber
)

// String returns the source-level keyword for a RecordType.
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

// Record is a single typed key/value declaration.
type Record struct {
	LeadingComments []string
	Type            RecordType
	Key             string
	Value           string
}

// Block is a named group of records.
type Block struct {
	LeadingComments []string
	Name            string
	Records         []Record
}

// File is the top-level AST node.
type File struct {
	Records []Record
	Blocks  []Block
}

// Type is the marker interface for AST node types in the KVR AST.
// Concrete node types implement isType() to satisfy this interface.
type Type interface {
	isType()
}

func (Record) isType() {}
func (Block) isType()  {}
func (File) isType()   {}

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
// value was not in the acceptable set of keywords at the current position.
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %s, expected one of %v", e.Got, e.Want)
}

// UnexpectedSymbolError is returned when the parser saw a symbol whose value
// was not the one required by the current grammar position.
type UnexpectedSymbolError struct {
	Got  Token
	Want string
}

func (e *UnexpectedSymbolError) Error() string {
	return fmt.Sprintf("unexpected symbol %s, expected %q", e.Got, e.Want)
}

// TypeMismatchError is returned when a record's declared type does not match
// the kind of value token following the `=`.
type TypeMismatchError struct {
	Pos  Pos
	Type RecordType
	Got  TokenType
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch at %d:%d: declared %s, got %s", e.Pos.Line, e.Pos.Column, e.Type, e.Got)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
//
// peeked holds a single look-ahead token so that an action can inspect the
// next token without consuming it (used for record termination at the top
// level — a record ends when the next valid statement opener appears).
type parser struct {
	next      func() (Token, error, bool)
	stop      func()
	peeked    Token
	hasPeeked bool
	peekedEnd bool
	peekedErr error
}

// peek returns the next token without consuming it. ok=false means the stream
// is exhausted (with err nil) or returned err.
func (p *parser) peek() (Token, error, bool) {
	if p.hasPeeked {
		if p.peekedEnd {
			return Token{}, p.peekedErr, false
		}
		return p.peeked, nil, true
	}
	tok, err, ok := p.next()
	p.hasPeeked = true
	if !ok {
		p.peekedEnd = true
		p.peekedErr = err
		return Token{}, err, false
	}
	if err != nil {
		p.peekedEnd = true
		p.peekedErr = err
		return Token{}, err, false
	}
	p.peeked = tok
	return tok, nil, true
}

// take consumes the look-ahead token (or pulls a fresh one if none).
func (p *parser) take() (Token, error, bool) {
	if p.hasPeeked {
		p.hasPeeked = false
		if p.peekedEnd {
			return Token{}, p.peekedErr, false
		}
		return p.peeked, nil, true
	}
	return p.next()
}

// expect pulls the next token and verifies its type matches one of the given
// types. Use it everywhere the grammar requires a specific token; never
// inline the type check.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.take()
	if !ok {
		if err != nil {
			return Token{}, err
		}
		return Token{}, &UnexpectedEndOfTokensError{}
	}
	for _, want := range types {
		if tok.Type == want {
			return tok, nil
		}
	}
	return Token{}, &UnexpectedTokenError{Got: tok, Want: types}
}

// expectSymbol pulls the next token and verifies it is the given symbol.
func (p *parser) expectSymbol(sym string) (Token, error) {
	tok, err := p.expect(TokenSymbol)
	if err != nil {
		return Token{}, err
	}
	if tok.Value != sym {
		return Token{}, &UnexpectedSymbolError{Got: tok, Want: sym}
	}
	return tok, nil
}

// expectKeyword pulls the next identifier token and verifies its value matches
// one of the given keywords.
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

// parseFile is the top-level action. Dispatches on the next token: a `record`
// keyword starts a record, a `block` keyword starts a block, end-of-stream
// finishes parsing, anything else is an error.
//
// Comments encountered at the top level become leading comments for the next
// non-comment statement.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	return parseFileWithComments(nil)(p, f)
}

// parseFileWithComments returns a top-level action that carries already-seen
// leading comments forward to the next statement.
func parseFileWithComments(comments []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		tok, err, ok := p.peek()
		if !ok {
			if err != nil {
				return nil, err
			}
			// Trailing comments without a statement after them are dropped per spec
			// (comments attach to the *following* non-comment statement).
			return nil, nil
		}
		switch tok.Type {
		case TokenComment:
			// consume and accumulate
			_, _, _ = p.take()
			return parseFileWithComments(append(comments, tok.Value)), nil
		case TokenIdentifier:
			switch tok.Value {
			case "record":
				return parseTopRecord(comments), nil
			case "block":
				return parseTopBlock(comments), nil
			default:
				return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "block"}}
			}
		default:
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenComment}}
		}
	}
}

// parseTopRecord parses a single record at file top level and appends it to f.
func parseTopRecord(comments []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		rec, err := parseRecord(p, comments)
		if err != nil {
			return nil, err
		}
		f.Records = append(f.Records, rec)
		return parseFile, nil
	}
}

// parseTopBlock parses a block at file top level and appends it to f.
//
// The block's body is built by an inner action loop with separate
// parserAction[*Block] functions for the open brace, each member, the
// separator, and the close brace.
func parseTopBlock(comments []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		blk := &Block{LeadingComments: comments}
		var err error
		for action := parseBlockOpen; action != nil && err == nil; {
			action, err = action(p, blk)
		}
		if err != nil {
			return nil, err
		}
		f.Blocks = append(f.Blocks, *blk)
		return parseFile, nil
	}
}

// parseRecord consumes a record statement (`record TYPE KEY = VALUE`) and
// returns the resulting Record. It does NOT consume any terminator; the
// caller decides what follows.
func parseRecord(p *parser, comments []string) (Record, error) {
	if _, err := p.expectKeyword("record"); err != nil {
		return Record{}, err
	}
	typeTok, err := p.expectKeyword("string", "number")
	if err != nil {
		return Record{}, err
	}
	var rt RecordType
	switch typeTok.Value {
	case "string":
		rt = RecordTypeString
	case "number":
		rt = RecordTypeNumber
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
	switch rt {
	case RecordTypeString:
		if valTok.Type != TokenString {
			return Record{}, &TypeMismatchError{Pos: valTok.Pos, Type: rt, Got: valTok.Type}
		}
	case RecordTypeNumber:
		if valTok.Type != TokenNumber {
			return Record{}, &TypeMismatchError{Pos: valTok.Pos, Type: rt, Got: valTok.Type}
		}
	}
	return Record{
		LeadingComments: comments,
		Type:            rt,
		Key:             keyTok.Value,
		Value:           valTok.Value,
	}, nil
}

// --- Block inner action loop ---

// parseBlockOpen consumes the `block NAME {` prefix. Note: the leading
// `block` keyword has not yet been consumed when this action runs.
func parseBlockOpen(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectKeyword("block"); err != nil {
		return nil, err
	}
	nameTok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	blk.Name = nameTok.Value
	if _, err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	return parseBlockMember(nil), nil
}

// parseBlockMember handles one inner statement. It either:
//   - sees `}` (empty block or after final `;` close) and finishes via
//     parseBlockClose;
//   - sees a comment and accumulates it for the next member;
//   - sees `record` and parses a record, then transitions to parseBlockSeparator.
func parseBlockMember(comments []string) parserAction[*Block] {
	return func(p *parser, blk *Block) (parserAction[*Block], error) {
		tok, err, ok := p.peek()
		if !ok {
			if err != nil {
				return nil, err
			}
			return nil, &UnexpectedEndOfTokensError{}
		}
		switch tok.Type {
		case TokenSymbol:
			if tok.Value == "}" {
				if len(comments) > 0 {
					// Dangling comments at end of block — drop, per spec
					// (comments attach to a following non-comment statement).
				}
				return parseBlockClose, nil
			}
			return nil, &UnexpectedSymbolError{Got: tok, Want: "}"}
		case TokenComment:
			_, _, _ = p.take()
			return parseBlockMember(append(comments, tok.Value)), nil
		case TokenIdentifier:
			if tok.Value != "record" {
				return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record"}}
			}
			rec, err := parseRecord(p, comments)
			if err != nil {
				return nil, err
			}
			blk.Records = append(blk.Records, rec)
			return parseBlockSeparator, nil
		default:
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenComment, TokenSymbol}}
		}
	}
}

// parseBlockSeparator consumes the mandatory `;` between (and after every)
// inner statement, then loops back to parseBlockMember.
func parseBlockSeparator(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol(";"); err != nil {
		return nil, err
	}
	return parseBlockMember(nil), nil
}

// parseBlockClose consumes the closing `}` and ends the block's inner loop.
func parseBlockClose(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol("}"); err != nil {
		return nil, err
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
