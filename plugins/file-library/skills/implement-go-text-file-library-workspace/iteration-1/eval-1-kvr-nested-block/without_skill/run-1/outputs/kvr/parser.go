package kvr

import (
	"fmt"
	"io"
	"iter"
)

// Value is the marker interface for typed record values.
type Value interface {
	isValue()
}

// StringValue is the AST node for a string-typed record value.
type StringValue struct {
	V string
}

func (StringValue) isValue() {}

// NumberValue is the AST node for a number-typed record value. The numeric
// text is preserved verbatim from the source; conversion into a numeric Go
// type is the consumer's responsibility.
type NumberValue struct {
	V string
}

func (NumberValue) isValue() {}

// Record is a single typed key-value declaration.
type Record struct {
	LeadingComments []string
	Type            string // "string" or "number"
	Key             string
	Value           Value
}

// Block is a named group of inner records.
type Block struct {
	LeadingComments []string
	Name            string
	Records         []Record
}

// File is the top-level AST node. It holds the file's top-level records and
// blocks as separate slices.
type File struct {
	Records []Record
	Blocks  []Block
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

// UnexpectedKeywordError is returned when an identifier-position keyword does
// not match what the grammar expects (e.g. a record where "record" is required).
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %s, expected one of %v", e.Got, e.Want)
}

// UnexpectedSymbolError is returned when a symbol token's value does not
// match what the grammar expects at that position.
type UnexpectedSymbolError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedSymbolError) Error() string {
	return fmt.Sprintf("unexpected symbol %s, expected one of %v", e.Got, e.Want)
}

// TypeMismatchError is returned when a record's value-token type does not
// agree with its declared type (e.g. `record string K = 42`).
type TypeMismatchError struct {
	Type string
	Got  Token
}

func (e *TypeMismatchError) Error() string {
	return fmt.Sprintf("type mismatch: declared %q, got %s", e.Type, e.Got)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time, with a single-token pushback buffer used for grammar-level peeks.
type parser struct {
	next    func() (Token, error, bool)
	stop    func()
	pushed  Token
	hasPush bool
}

// pull returns the next token, honouring any pushed-back token.
func (p *parser) pull() (Token, error, bool) {
	if p.hasPush {
		t := p.pushed
		p.hasPush = false
		p.pushed = Token{}
		return t, nil, true
	}
	return p.next()
}

// unread pushes a single token back onto the stream. At most one token may
// be in the pushback buffer at a time.
func (p *parser) unread(t Token) {
	p.pushed = t
	p.hasPush = true
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

// expectSymbol pulls the next token, verifies its type is TokenSymbol, and
// verifies its value matches one of the given symbol strings.
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
	return Token{}, &UnexpectedSymbolError{Got: tok, Want: values}
}

// expectKeyword pulls the next token, verifies it is an identifier, and
// verifies its value matches one of the given keyword strings.
func (p *parser) expectKeyword(values ...string) (Token, error) {
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

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level dispatch action. It collects any leading
// comments, then dispatches to record or block parsing based on the next
// identifier keyword.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	tok, err, ok := p.pull()
	if !ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var leading []string
	for tok.Type == TokenComment {
		leading = append(leading, tok.Value)
		tok, err, ok = p.pull()
		if !ok {
			// trailing free-floating comments at EOF are dropped — the format
			// only round-trips comments that attach to a following statement.
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
	default:
		return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "block"}}
	}
}

// parseRecordWithLeading returns a parserAction[*File] that parses a record
// (the "record" keyword has already been consumed) and appends it to f.Records
// with the given leading comments.
func parseRecordWithLeading(leading []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		rec, err := parseRecordBody(p, leading)
		if err != nil {
			return nil, err
		}
		f.Records = append(f.Records, rec)
		return parseFile, nil
	}
}

// parseRecordBody parses the tail of a record statement: the type, key, `=`,
// and value. The `record` keyword is assumed to have been consumed already.
func parseRecordBody(p *parser, leading []string) (Record, error) {
	typeTok, err := p.expectKeyword("string", "number")
	if err != nil {
		return Record{}, err
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
	rec := Record{
		LeadingComments: leading,
		Type:            typeTok.Value,
		Key:             keyTok.Value,
	}
	switch typeTok.Value {
	case "string":
		if valTok.Type != TokenString {
			return Record{}, &TypeMismatchError{Type: "string", Got: valTok}
		}
		rec.Value = StringValue{V: valTok.Value}
	case "number":
		if valTok.Type != TokenNumber {
			return Record{}, &TypeMismatchError{Type: "number", Got: valTok}
		}
		rec.Value = NumberValue{V: valTok.Value}
	}
	return rec, nil
}

// parseBlockWithLeading returns a parserAction[*File] that parses a block
// (the "block" keyword has already been consumed) and appends it to f.Blocks
// with the given leading comments.
func parseBlockWithLeading(leading []string) parserAction[*File] {
	return func(p *parser, f *File) (parserAction[*File], error) {
		nameTok, err := p.expect(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		blk := &Block{LeadingComments: leading, Name: nameTok.Value}

		// Inner action loop over *Block. Each grammar position is its own
		// parserAction[*Block]: open brace, body dispatch, member, separator,
		// close brace. Per CLAUDE.md, complex types must use this pattern
		// instead of a flat for-with-switch.
		action := parseBlockOpen
		for action != nil {
			action, err = action(p, blk)
			if err != nil {
				return nil, err
			}
		}
		f.Blocks = append(f.Blocks, *blk)
		return parseFile, nil
	}
}

// parseBlockOpen consumes the `{` that opens a block.
func parseBlockOpen(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	return parseBlockBody, nil
}

// parseBlockBody peeks the next token to decide between parsing another
// member or closing the block. Comments are accumulated as leading comments
// on the next member.
func parseBlockBody(p *parser, blk *Block) (parserAction[*Block], error) {
	tok, err, ok := p.pull()
	if !ok {
		return nil, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return nil, err
	}
	if tok.Type == TokenSymbol && tok.Value == "}" {
		// Hand off to parseBlockClose with the brace already pulled — but our
		// rule says close-brace is its own action, so push the token back and
		// let parseBlockClose consume it cleanly.
		p.unread(tok)
		return parseBlockClose, nil
	}
	// Otherwise we expect a member; push the token back so parseBlockMember
	// can read it and any preceding comments. We accumulate any TokenComment
	// run here into the next member's leading comments.
	if tok.Type == TokenComment {
		leading := []string{tok.Value}
		for {
			next, err, ok := p.pull()
			if !ok {
				return nil, &UnexpectedEndOfTokensError{}
			}
			if err != nil {
				return nil, err
			}
			if next.Type == TokenComment {
				leading = append(leading, next.Value)
				continue
			}
			p.unread(next)
			return parseBlockMemberWithLeading(leading), nil
		}
	}
	p.unread(tok)
	return parseBlockMember, nil
}

// parseBlockMember parses one inner statement (currently always a record) and
// appends it to blk.Records.
func parseBlockMember(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectKeyword("record"); err != nil {
		return nil, err
	}
	rec, err := parseRecordBody(p, nil)
	if err != nil {
		return nil, err
	}
	blk.Records = append(blk.Records, rec)
	return parseBlockSeparator, nil
}

// parseBlockMemberWithLeading parses one inner record and attaches the given
// leading comments. Used after parseBlockBody has accumulated a comment run.
func parseBlockMemberWithLeading(leading []string) parserAction[*Block] {
	return func(p *parser, blk *Block) (parserAction[*Block], error) {
		if _, err := p.expectKeyword("record"); err != nil {
			return nil, err
		}
		rec, err := parseRecordBody(p, leading)
		if err != nil {
			return nil, err
		}
		blk.Records = append(blk.Records, rec)
		return parseBlockSeparator, nil
	}
}

// parseBlockSeparator consumes the `;` that follows every inner record.
func parseBlockSeparator(p *parser, blk *Block) (parserAction[*Block], error) {
	if _, err := p.expectSymbol(";"); err != nil {
		return nil, err
	}
	return parseBlockBody, nil
}

// parseBlockClose consumes the `}` that closes a block.
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
