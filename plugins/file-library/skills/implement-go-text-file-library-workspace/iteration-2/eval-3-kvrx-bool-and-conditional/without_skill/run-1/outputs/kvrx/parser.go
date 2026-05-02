package kvrx

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node.
type File struct {
	Statements []Statement
}

// Statement is the marker interface for top-level statement AST nodes.
type Statement interface {
	isStatement()
}

// Type is the marker interface for AST node types.
type Type interface {
	isType()
}

// Expression is the marker interface for expression AST nodes.
type Expression interface {
	isExpression()
}

// NamedType is the simple `string` / `bool` / `number` form.
type NamedType struct {
	Pos  Pos
	Name string
}

func (NamedType) isType() {}

// Record is `record TYPE KEY = EXPR`.
type Record struct {
	Pos   Pos
	Type  Type
	Key   string
	Value Expression
}

func (Record) isStatement() {}

// BoolLiteral represents `true` / `false`.
type BoolLiteral struct {
	Pos   Pos
	Value bool
}

func (BoolLiteral) isExpression() {}

// Reference represents `&NAME`.
type Reference struct {
	Pos  Pos
	Name string
}

func (Reference) isExpression() {}

// BinaryExpr is a comparison or other binary operation. For this slice only
// `==` is exercised at runtime; the field is kept open so the AST shape can
// extend without a breaking change.
type BinaryExpr struct {
	Pos   Pos
	Op    string
	Left  Expression
	Right Expression
}

func (BinaryExpr) isExpression() {}

// ConditionalBranch is one `if (...) {...}` / `elif (...) {...}` / `else {...}`.
// `Kind` is "if", "elif", or "else"; an "else" branch's Cond is nil.
type ConditionalBranch struct {
	Pos        Pos
	Kind       string
	Cond       Expression
	Body       []Statement
}

// Conditional is a chain of branches. The branches list always has the leading
// `if` branch as element 0; any `elif` follow; the optional `else` is last.
type Conditional struct {
	Pos      Pos
	Branches []ConditionalBranch
}

func (Conditional) isStatement() {}

// UnexpectedEndOfTokensError is returned when the parser ran out of tokens.
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
// value didn't match the keyword(s) it expected.
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %q at %d:%d, expected one of %v", e.Got.Value, e.Got.Pos.Line, e.Got.Pos.Column, e.Want)
}

// UndeclaredReferenceError is returned when a `&NAME` reference cannot be
// resolved to a previously-declared record.
type UndeclaredReferenceError struct {
	Pos  Pos
	Name string
}

func (e *UndeclaredReferenceError) Error() string {
	return fmt.Sprintf("undeclared reference %q at %d:%d", e.Name, e.Pos.Line, e.Pos.Column)
}

// NonStaticConditionalError is returned when a conditional's expression
// cannot be reduced to a static bool at parse time.
type NonStaticConditionalError struct {
	Pos Pos
}

func (e *NonStaticConditionalError) Error() string {
	return fmt.Sprintf("conditional expression is not static at %d:%d", e.Pos.Line, e.Pos.Column)
}

// parser drives the AST build.
type parser struct {
	next     func() (Token, error, bool)
	stop     func()
	peeked   *Token
	hasPeek  bool
	peekErr  error
	peekDone bool

	// records holds previously-declared top-level records for use by the
	// conditional resolver. Inner-block records would extend this scope chain;
	// for this slice the flat list is sufficient.
	records []*Record
}

// nextNonTrivia pulls the next token, skipping newlines and comments.
// Comments are dropped because this slice's targeted fixtures do not assert
// on leading-comment attachment.
func (p *parser) nextNonTrivia() (Token, error, bool) {
	for {
		var tok Token
		var err error
		var ok bool
		if p.hasPeek {
			tok, err, ok = *p.peeked, p.peekErr, !p.peekDone
			p.hasPeek = false
			p.peeked = nil
			p.peekErr = nil
			p.peekDone = false
		} else {
			tok, err, ok = p.next()
		}
		if !ok {
			return Token{}, err, false
		}
		if err != nil {
			return Token{}, err, true
		}
		if tok.Type == TokenNewline || tok.Type == TokenComment {
			continue
		}
		return tok, nil, true
	}
}

// peekNonTrivia stores the next non-trivia token in the parser's peek slot
// without consuming it.
func (p *parser) peekNonTrivia() (Token, error, bool) {
	if p.hasPeek {
		return *p.peeked, p.peekErr, !p.peekDone
	}
	tok, err, ok := p.nextNonTrivia()
	pTok := tok
	p.peeked = &pTok
	p.peekErr = err
	p.peekDone = !ok
	p.hasPeek = true
	return tok, err, ok
}

// expect pulls the next non-trivia token and verifies its type matches.
func (p *parser) expect(types ...TokenType) (Token, error) {
	tok, err, ok := p.nextNonTrivia()
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

// expectSymbol pulls the next non-trivia token and verifies it is the given
// symbol value.
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

// expectKeyword pulls the next non-trivia token and verifies it is an
// identifier whose value is one of the given keywords.
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

// parserAction is one step in the parser state machine.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action.
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	tok, err, ok := p.peekNonTrivia()
	if !ok {
		if err != nil {
			return nil, err
		}
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
		return parseRecordStatement, nil
	case "if":
		return parseConditionalStatement, nil
	default:
		return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "if"}}
	}
}

// parseRecordStatement parses one top-level `record TYPE KEY = EXPR`.
func parseRecordStatement(p *parser, f *File) (parserAction[*File], error) {
	rec, err := parseRecord(p)
	if err != nil {
		return nil, err
	}
	f.Statements = append(f.Statements, *rec)
	p.records = append(p.records, rec)
	return parseFile, nil
}

// parseRecord parses one record (used both at the top level and inside a
// conditional body).
func parseRecord(p *parser) (*Record, error) {
	rec := &Record{}
	var err error
	for action := parseRecordKeyword; action != nil && err == nil; {
		action, err = action(p, rec)
	}
	if err != nil {
		return nil, err
	}
	return rec, nil
}

func parseRecordKeyword(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expectKeyword("record")
	if err != nil {
		return nil, err
	}
	r.Pos = tok.Pos
	return parseRecordType, nil
}

func parseRecordType(p *parser, r *Record) (parserAction[*Record], error) {
	tok, err := p.expect(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	r.Type = NamedType{Pos: tok.Pos, Name: tok.Value}
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
	if _, err := p.expectSymbol("="); err != nil {
		return nil, err
	}
	return parseRecordValue, nil
}

func parseRecordValue(p *parser, r *Record) (parserAction[*Record], error) {
	expr, err := parseExpression(p)
	if err != nil {
		return nil, err
	}
	r.Value = expr
	return nil, nil
}

// parseExpression parses one expression. For this slice the supported shapes
// are: BoolLiteral, Reference, and a single `==` comparison between any two
// of those.
func parseExpression(p *parser) (Expression, error) {
	left, err := parsePrimary(p)
	if err != nil {
		return nil, err
	}
	tok, errPeek, ok := p.peekNonTrivia()
	if !ok {
		return left, nil
	}
	if errPeek != nil {
		return nil, errPeek
	}
	if tok.Type == TokenSymbol && tok.Value == "==" {
		// consume
		_, _, _ = p.nextNonTrivia()
		right, err := parsePrimary(p)
		if err != nil {
			return nil, err
		}
		return BinaryExpr{Pos: posOfExpression(left), Op: "==", Left: left, Right: right}, nil
	}
	return left, nil
}

// parsePrimary parses a `&NAME` reference or a `true`/`false` bool literal.
func parsePrimary(p *parser) (Expression, error) {
	tok, err, ok := p.peekNonTrivia()
	if !ok {
		return nil, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return nil, err
	}
	switch {
	case tok.Type == TokenSymbol && tok.Value == "&":
		_, _, _ = p.nextNonTrivia() // consume &
		ident, err := p.expect(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		return Reference{Pos: tok.Pos, Name: ident.Value}, nil
	case tok.Type == TokenIdentifier && (tok.Value == "true" || tok.Value == "false"):
		_, _, _ = p.nextNonTrivia()
		return BoolLiteral{Pos: tok.Pos, Value: tok.Value == "true"}, nil
	default:
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenSymbol}}
	}
}

func posOfExpression(e Expression) Pos {
	switch v := e.(type) {
	case BoolLiteral:
		return v.Pos
	case Reference:
		return v.Pos
	case BinaryExpr:
		return v.Pos
	}
	return Pos{}
}

// parseConditionalStatement handles the full `if (Expr) { ... } { elif ... } [ else ... ]`.
func parseConditionalStatement(p *parser, f *File) (parserAction[*File], error) {
	cond := &Conditional{}
	var err error
	for action := parseConditionalIf; action != nil && err == nil; {
		action, err = action(p, cond)
	}
	if err != nil {
		return nil, err
	}
	// Verify the conditional's `if` branch evaluates to a static bool.
	// Per spec we still want to flag NonStaticConditionalError if the
	// resolver can't reduce the expression — but the inactive branch
	// expressions are not type-checked. We only check the chain's
	// expressions can be evaluated; if the active branch is found we
	// keep going. If none can be evaluated the resolver returns false
	// for them and we fall through to else.
	for _, br := range cond.Branches {
		if br.Kind == "else" {
			continue
		}
		if _, err := evalCondition(p, br.Cond, br.Pos); err != nil {
			return nil, err
		}
	}
	f.Statements = append(f.Statements, *cond)
	return parseFile, nil
}

func parseConditionalIf(p *parser, c *Conditional) (parserAction[*Conditional], error) {
	tok, err := p.expectKeyword("if")
	if err != nil {
		return nil, err
	}
	c.Pos = tok.Pos
	br, err := parseConditionalIfElifBranch(p, "if", tok.Pos)
	if err != nil {
		return nil, err
	}
	c.Branches = append(c.Branches, br)
	return parseConditionalContinuation, nil
}

func parseConditionalContinuation(p *parser, c *Conditional) (parserAction[*Conditional], error) {
	tok, err, ok := p.peekNonTrivia()
	if !ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if tok.Type != TokenIdentifier {
		return nil, nil
	}
	switch tok.Value {
	case "elif":
		_, _, _ = p.nextNonTrivia()
		br, err := parseConditionalIfElifBranch(p, "elif", tok.Pos)
		if err != nil {
			return nil, err
		}
		c.Branches = append(c.Branches, br)
		return parseConditionalContinuation, nil
	case "else":
		_, _, _ = p.nextNonTrivia()
		br, err := parseConditionalElseBranch(p, tok.Pos)
		if err != nil {
			return nil, err
		}
		c.Branches = append(c.Branches, br)
		return nil, nil
	default:
		return nil, nil
	}
}

func parseConditionalIfElifBranch(p *parser, kind string, pos Pos) (ConditionalBranch, error) {
	if _, err := p.expectSymbol("("); err != nil {
		return ConditionalBranch{}, err
	}
	cond, err := parseExpression(p)
	if err != nil {
		return ConditionalBranch{}, err
	}
	if _, err := p.expectSymbol(")"); err != nil {
		return ConditionalBranch{}, err
	}
	body, err := parseConditionalBody(p)
	if err != nil {
		return ConditionalBranch{}, err
	}
	return ConditionalBranch{Pos: pos, Kind: kind, Cond: cond, Body: body}, nil
}

func parseConditionalElseBranch(p *parser, pos Pos) (ConditionalBranch, error) {
	body, err := parseConditionalBody(p)
	if err != nil {
		return ConditionalBranch{}, err
	}
	return ConditionalBranch{Pos: pos, Kind: "else", Body: body}, nil
}

// parseConditionalBody parses `{ Statement ";" ... }`. Each inner statement
// must be a record (per the slice's scope).
func parseConditionalBody(p *parser) ([]Statement, error) {
	if _, err := p.expectSymbol("{"); err != nil {
		return nil, err
	}
	var body []Statement
	for {
		tok, err, ok := p.peekNonTrivia()
		if !ok {
			return nil, &UnexpectedEndOfTokensError{}
		}
		if err != nil {
			return nil, err
		}
		if tok.Type == TokenSymbol && tok.Value == "}" {
			_, _, _ = p.nextNonTrivia()
			return body, nil
		}
		// Inside a conditional body only record statements are accepted by
		// this slice. The grammar permits any Statement; we widen as needed.
		if tok.Type != TokenIdentifier || tok.Value != "record" {
			return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record"}}
		}
		rec, err := parseRecord(p)
		if err != nil {
			return nil, err
		}
		body = append(body, *rec)
		// Require trailing ;
		if _, err := p.expectSymbol(";"); err != nil {
			return nil, err
		}
	}
}

// evalCondition reduces a conditional's expression to a static bool. The
// supported shapes are:
//   - BoolLiteral
//   - Reference (must resolve to a previously-declared record whose value
//     is a BoolLiteral)
//   - BinaryExpr with Op == "==" between any two of the above
//
// Anything else returns NonStaticConditionalError.
func evalCondition(p *parser, e Expression, pos Pos) (bool, error) {
	switch v := e.(type) {
	case BoolLiteral:
		return v.Value, nil
	case Reference:
		val, ok := lookupBool(p, v.Name)
		if !ok {
			return false, &UndeclaredReferenceError{Pos: v.Pos, Name: v.Name}
		}
		return val, nil
	case BinaryExpr:
		if v.Op != "==" {
			return false, &NonStaticConditionalError{Pos: pos}
		}
		l, err := evalOperand(p, v.Left, pos)
		if err != nil {
			return false, err
		}
		r, err := evalOperand(p, v.Right, pos)
		if err != nil {
			return false, err
		}
		return l == r, nil
	default:
		return false, &NonStaticConditionalError{Pos: pos}
	}
}

// evalOperand reduces a primary expression to a bool for use inside a
// comparison. Only BoolLiteral and Reference (resolving to a bool record)
// are supported.
func evalOperand(p *parser, e Expression, pos Pos) (bool, error) {
	switch v := e.(type) {
	case BoolLiteral:
		return v.Value, nil
	case Reference:
		val, ok := lookupBool(p, v.Name)
		if !ok {
			return false, &UndeclaredReferenceError{Pos: v.Pos, Name: v.Name}
		}
		return val, nil
	default:
		return false, &NonStaticConditionalError{Pos: pos}
	}
}

// lookupBool walks the parser's record list (last-wins) for a bool record
// matching name.
func lookupBool(p *parser, name string) (bool, bool) {
	for i := len(p.records) - 1; i >= 0; i-- {
		r := p.records[i]
		if r.Key != name {
			continue
		}
		bl, ok := r.Value.(BoolLiteral)
		if !ok {
			return false, false
		}
		return bl.Value, true
	}
	return false, false
}

// Parse reads a KVRX file from r.
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
