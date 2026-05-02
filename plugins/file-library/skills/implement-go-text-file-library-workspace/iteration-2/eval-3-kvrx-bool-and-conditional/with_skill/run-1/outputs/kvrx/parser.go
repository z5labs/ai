package kvrx

import (
	"fmt"
	"io"
	"iter"
)

// File is the top-level AST node. Its Statements slice holds the top-level
// declarations of the source in declaration order. The empty file is a legal
// KVRX file (Parse("") yields &File{} per the Examples).
type File struct {
	Statements []Statement
}

// Statement is the marker interface for top-level statement AST nodes. The
// concrete kinds for the bool+conditional scope are *Record and *Conditional.
type Statement interface {
	isStatement()
}

// Type is the marker interface for AST node types in the KVRX AST.
// Concrete node types implement isType() to satisfy this interface.
type Type interface {
	isType()
}

// Expression is the marker interface for expression AST nodes. The
// bool+conditional scope only needs bool literals plus a tiny subset of
// expression nodes used by the conditional's condition (Reference,
// EqualExpr).
type Expression interface {
	isExpression()
}

// Record is `record TYPE KEY = EXPR`. Pos is the position of the `record`
// keyword. Type is the declared type identifier (e.g. "bool"). Key is the
// record's identifier. Value is the parsed right-hand-side expression.
type Record struct {
	Pos   Pos
	Type  string
	Key   string
	Value Expression
}

func (*Record) isStatement() {}

// BoolLiteral is `true` or `false`. Pos is the position of the literal's
// first rune.
type BoolLiteral struct {
	Pos   Pos
	Value bool
}

func (*BoolLiteral) isExpression() {}

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

// UnexpectedKeywordError is returned when the parser saw an identifier at
// statement position whose value is not one of the recognised statement-form
// keywords. Got is the offending token; Want lists the accepted keyword
// values (lowercase identifier values).
type UnexpectedKeywordError struct {
	Got  Token
	Want []string
}

func (e *UnexpectedKeywordError) Error() string {
	return fmt.Sprintf("unexpected keyword %q at %d:%d, expected one of %v", e.Got.Value, e.Got.Pos.Line, e.Got.Pos.Column, e.Want)
}

// parser drives the AST build. It pulls tokens from the tokenizer one at a
// time and runs action functions against the AST node currently being built.
// peeked is an at-most-one push-back slot so an action that pulled a token
// past its production can re-inject it for the next action.
type parser struct {
	next   func() (Token, error, bool)
	stop   func()
	peeked *Token
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

// peekNonTrivia advances past TokenNewline (and TokenComment, dropped silently
// in this scope) and returns the next significant token. It does NOT consume
// the returned token if put-back is needed; instead, the caller threads
// trivia-skipping through expect() — but for top-level dispatch we use a
// peek buffer on the parser. We model this by just consuming-and-replaying:
// since iter.Pull2 has no put-back, we instead make every consumer use
// expectSkipTrivia which loops past trivia tokens.
//
// A small wrapper: expect-with-skip-trivia.
func (p *parser) expectSkipTrivia(types ...TokenType) (Token, error) {
	for {
		tok, err, ok := p.next()
		if !ok {
			return Token{}, &UnexpectedEndOfTokensError{}
		}
		if err != nil {
			return Token{}, err
		}
		if tok.Type == TokenNewline || tok.Type == TokenComment {
			continue
		}
		for _, want := range types {
			if tok.Type == want {
				return tok, nil
			}
		}
		return Token{}, &UnexpectedTokenError{Got: tok, Want: types}
	}
}

// peekStatementOpener pulls tokens past trivia until it finds a non-trivia
// token. It returns the token plus a "putBack" function that re-injects the
// token so the next p.next() returns it. Implementation: store the token
// in p.peeked and have p.next() drain that first.
func (p *parser) peekStatementOpener() (Token, bool, error) {
	for {
		tok, err, ok := p.next()
		if !ok {
			return Token{}, false, nil
		}
		if err != nil {
			return Token{}, false, err
		}
		if tok.Type == TokenNewline || tok.Type == TokenComment {
			continue
		}
		// non-trivia — push back via the parser's one-slot buffer
		p.pushBack(tok)
		return tok, true, nil
	}
}

// pushBack stores a token to be returned by the next p.next() call. The
// parser allows at most one pending push-back.
func (p *parser) pushBack(tok Token) {
	p.peeked = &tok
}

// parserAction is one step in the parser state machine, generic over the AST
// node currently being built. Returning (nil, nil) completes successfully;
// (nil, err) terminates with error.
type parserAction[T any] func(p *parser, t T) (parserAction[T], error)

// parseFile is the top-level action. It dispatches on the first non-trivia
// token's value (per the Statement disambiguation table in SPEC §Structure).
func parseFile(p *parser, f *File) (parserAction[*File], error) {
	tok, ok, err := p.peekStatementOpener()
	if err != nil {
		return nil, err
	}
	if !ok {
		// EOF — done.
		return nil, nil
	}
	switch tok.Type {
	case TokenIdentifier:
		switch tok.Value {
		case "record":
			return parseRecord, nil
		case "if":
			return parseConditional, nil
		default:
			return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "if"}}
		}
	default:
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier}}
	}
}

// parseRecord parses `record TYPE KEY = EXPR` using an inner action loop, as
// required for any complex type per references/architecture.md.
func parseRecord(p *parser, f *File) (parserAction[*File], error) {
	rec := &Record{}
	var err error
	for action := parseRecordKeyword; action != nil && err == nil; {
		action, err = action(p, rec)
	}
	if err != nil {
		return nil, err
	}
	f.Statements = append(f.Statements, rec)
	return parseFile, nil
}

// parseRecordKeyword consumes the `record` identifier and stores its
// position on the record.
func parseRecordKeyword(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expectSkipTrivia(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	if tok.Value != "record" {
		return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record"}}
	}
	rec.Pos = tok.Pos
	return parseRecordType, nil
}

// parseRecordType consumes the type identifier (`bool`, `string`, etc.).
// The bool+conditional scope only validates that the type identifier is
// present; full Type parsing (ListType, MapType, alias resolution) is out
// of scope for this slice.
func parseRecordType(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expectSkipTrivia(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	rec.Type = tok.Value
	return parseRecordKey, nil
}

// parseRecordKey consumes the record's key identifier.
func parseRecordKey(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expectSkipTrivia(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	rec.Key = tok.Value
	return parseRecordEquals, nil
}

// parseRecordEquals consumes the `=` symbol.
func parseRecordEquals(p *parser, rec *Record) (parserAction[*Record], error) {
	tok, err := p.expectSkipTrivia(TokenSymbol)
	if err != nil {
		return nil, err
	}
	if tok.Value != "=" {
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenSymbol}}
	}
	return parseRecordValue, nil
}

// parseRecordValue parses the right-hand-side expression. For the bool+
// conditional scope the only supported value form is a bool literal
// (true/false). String/number/reference values can be added in a later
// phase along with the rest of the Expression grammar.
func parseRecordValue(p *parser, rec *Record) (parserAction[*Record], error) {
	expr, err := parseExpression(p)
	if err != nil {
		return nil, err
	}
	rec.Value = expr
	return nil, nil
}

// parseExpression is a stub that handles only the expression forms required
// by the bool+conditional scope: bool literals, string literals, references,
// and `==` comparison (used by the conditional's condition). Full operator-
// precedence parsing is intentionally out of scope.
//
// Grammar (this stub):
//
//	Expression  = Primary [ "==" Primary ] .
//	Primary     = "true" | "false" | TokenString | "&" Identifier .
func parseExpression(p *parser) (Expression, error) {
	left, err := parsePrimary(p)
	if err != nil {
		return nil, err
	}
	// Look ahead for `==`.
	tok, err, ok := p.next()
	if !ok {
		return left, nil
	}
	if err != nil {
		return nil, err
	}
	if tok.Type == TokenSymbol && tok.Value == "==" {
		right, err := parsePrimary(p)
		if err != nil {
			return nil, err
		}
		return &EqualExpr{Pos: tok.Pos, Left: left, Right: right}, nil
	}
	// not a `==` operator — push the token back so the caller sees it.
	p.pushBack(tok)
	return left, nil
}

// parsePrimary parses a single primary expression. See parseExpression's
// grammar for the supported forms.
func parsePrimary(p *parser) (Expression, error) {
	tok, err, ok := p.next()
	if !ok {
		return nil, &UnexpectedEndOfTokensError{}
	}
	if err != nil {
		return nil, err
	}
	switch tok.Type {
	case TokenIdentifier:
		switch tok.Value {
		case "true":
			return &BoolLiteral{Pos: tok.Pos, Value: true}, nil
		case "false":
			return &BoolLiteral{Pos: tok.Pos, Value: false}, nil
		default:
			return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"true", "false"}}
		}
	case TokenString:
		return &StringLiteral{Pos: tok.Pos, Value: tok.Value}, nil
	case TokenSymbol:
		if tok.Value == "&" {
			ref, err := p.expectSkipTrivia(TokenIdentifier)
			if err != nil {
				return nil, err
			}
			return &Reference{Pos: tok.Pos, Name: ref.Value}, nil
		}
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenString}}
	default:
		return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier, TokenString}}
	}
}

// StringLiteral, Reference, EqualExpr are declared up-front because
// parseExpression / parsePrimary refer to them. Their full role (used by the
// conditional's expression) is documented when the conditional sub-unit
// wires them in.
type StringLiteral struct {
	Pos   Pos
	Value string
}

func (*StringLiteral) isExpression() {}

type Reference struct {
	Pos  Pos
	Name string
}

func (*Reference) isExpression() {}

type EqualExpr struct {
	Pos   Pos
	Left  Expression
	Right Expression
}

func (*EqualExpr) isExpression() {}

// Conditional is `if (...) { ... } { elif (...) { ... } } [ else { ... } ]`.
// Branches preserves every clause in source order so the printer can replay
// them. Active is the index of the branch whose condition evaluated to true
// at parse time, or -1 when no branch took (no else, or only-false ifs).
type Conditional struct {
	Pos      Pos
	Branches []Branch
	Active   int
}

func (*Conditional) isStatement() {}

// Branch is one clause of a Conditional. Keyword is "if", "elif", or "else".
// Condition is nil for the "else" clause; non-nil for "if" and "elif".
// Body holds the inner statements of this clause in declaration order.
type Branch struct {
	Pos       Pos
	Keyword   string
	Condition Expression
	Body      []Statement
}

// UndeclaredReferenceError is returned when a `&NAME` reference resolves to
// no record in the enclosing scope chain.
type UndeclaredReferenceError struct {
	Pos  Pos
	Name string
}

func (e *UndeclaredReferenceError) Error() string {
	return fmt.Sprintf("undeclared reference &%s at %d:%d", e.Name, e.Pos.Line, e.Pos.Column)
}

// NonStaticConditionalError is returned when a conditional's expression
// cannot be reduced to a literal bool at parse time.
type NonStaticConditionalError struct {
	Pos Pos
}

func (e *NonStaticConditionalError) Error() string {
	return fmt.Sprintf("non-static conditional at %d:%d", e.Pos.Line, e.Pos.Column)
}

// parseConditional parses an if/elif*/else? chain via an inner action loop.
func parseConditional(p *parser, f *File) (parserAction[*File], error) {
	cond := &Conditional{Active: -1}
	var err error
	for action := parseConditionalIf; action != nil && err == nil; {
		action, err = action(p, cond)
	}
	if err != nil {
		return nil, err
	}
	// Validate references in every branch's condition against the file's
	// already-declared records (per spec §References — undeclared refs
	// fail at parse time).
	for _, br := range cond.Branches {
		if br.Condition == nil {
			continue
		}
		if err := validateReferences(br.Condition, f); err != nil {
			return nil, err
		}
	}
	// Resolve the active branch using the file's already-parsed records.
	cond.Active = pickActiveBranch(cond, f)
	f.Statements = append(f.Statements, cond)
	return parseFile, nil
}

// validateReferences walks expr and returns an UndeclaredReferenceError on
// the first reference whose Name does not resolve in f.
func validateReferences(expr Expression, f *File) error {
	switch e := expr.(type) {
	case *Reference:
		if _, ok := lookupRecord(f, e.Name); !ok {
			return &UndeclaredReferenceError{Pos: e.Pos, Name: e.Name}
		}
	case *EqualExpr:
		if err := validateReferences(e.Left, f); err != nil {
			return err
		}
		if err := validateReferences(e.Right, f); err != nil {
			return err
		}
	}
	return nil
}

// parseConditionalIf consumes the leading `if` keyword and its branch.
func parseConditionalIf(p *parser, cond *Conditional) (parserAction[*Conditional], error) {
	tok, err := p.expectSkipTrivia(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	if tok.Value != "if" {
		return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"if"}}
	}
	cond.Pos = tok.Pos
	br, err := parseBranchTail(p, tok.Pos, "if", true)
	if err != nil {
		return nil, err
	}
	cond.Branches = append(cond.Branches, br)
	return parseConditionalNext, nil
}

// parseConditionalNext peeks for `elif`, `else`, or end-of-conditional. The
// peek mechanism uses the parser's pushBack slot.
func parseConditionalNext(p *parser, cond *Conditional) (parserAction[*Conditional], error) {
	tok, ok, err := p.peekStatementOpener()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil // EOF — chain ends
	}
	if tok.Type != TokenIdentifier {
		return nil, nil // not a continuation keyword
	}
	switch tok.Value {
	case "elif":
		// consume the peeked token
		_, err := p.expectSkipTrivia(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		br, err := parseBranchTail(p, tok.Pos, "elif", true)
		if err != nil {
			return nil, err
		}
		cond.Branches = append(cond.Branches, br)
		return parseConditionalNext, nil
	case "else":
		_, err := p.expectSkipTrivia(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		br, err := parseBranchTail(p, tok.Pos, "else", false)
		if err != nil {
			return nil, err
		}
		cond.Branches = append(cond.Branches, br)
		return nil, nil // else terminates the chain
	default:
		return nil, nil
	}
}

// parseBranchTail parses the part of a branch following its keyword:
// `( Expression )` (when hasCondition) then `{ Statement* }`. pos is the
// position of the keyword.
func parseBranchTail(p *parser, pos Pos, keyword string, hasCondition bool) (Branch, error) {
	br := Branch{Pos: pos, Keyword: keyword}
	if hasCondition {
		open, err := p.expectSkipTrivia(TokenSymbol)
		if err != nil {
			return br, err
		}
		if open.Value != "(" {
			return br, &UnexpectedTokenError{Got: open, Want: []TokenType{TokenSymbol}}
		}
		expr, err := parseExpression(p)
		if err != nil {
			return br, err
		}
		br.Condition = expr
		close, err := p.expectSkipTrivia(TokenSymbol)
		if err != nil {
			return br, err
		}
		if close.Value != ")" {
			return br, &UnexpectedTokenError{Got: close, Want: []TokenType{TokenSymbol}}
		}
	}
	open, err := p.expectSkipTrivia(TokenSymbol)
	if err != nil {
		return br, err
	}
	if open.Value != "{" {
		return br, &UnexpectedTokenError{Got: open, Want: []TokenType{TokenSymbol}}
	}
	body, err := parseBranchBody(p)
	if err != nil {
		return br, err
	}
	br.Body = body
	return br, nil
}

// parseBranchBody parses `{ Statement ";" } "}"`. Inside a block (per spec)
// statements are explicitly terminated with `;`. The bool+conditional scope
// supports `record bool` statements and (recursively) nested conditionals
// inside a branch body.
func parseBranchBody(p *parser) ([]Statement, error) {
	var stmts []Statement
	for {
		tok, ok, err := p.peekStatementOpener()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, &UnexpectedEndOfTokensError{}
		}
		if tok.Type == TokenSymbol && tok.Value == "}" {
			// consume the closing brace
			_, err := p.expectSkipTrivia(TokenSymbol)
			if err != nil {
				return nil, err
			}
			return stmts, nil
		}
		if tok.Type != TokenIdentifier {
			return nil, &UnexpectedTokenError{Got: tok, Want: []TokenType{TokenIdentifier}}
		}
		switch tok.Value {
		case "record":
			rec, err := parseRecordStandalone(p)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, rec)
		case "if":
			inner, err := parseConditionalStandalone(p)
			if err != nil {
				return nil, err
			}
			stmts = append(stmts, inner)
		default:
			return nil, &UnexpectedKeywordError{Got: tok, Want: []string{"record", "if"}}
		}
		// require `;` after the statement (per Block grammar in spec).
		semi, err := p.expectSkipTrivia(TokenSymbol)
		if err != nil {
			return nil, err
		}
		if semi.Value != ";" {
			return nil, &UnexpectedTokenError{Got: semi, Want: []TokenType{TokenSymbol}}
		}
	}
}

// parseRecordStandalone reuses the record-parsing inner loop without going
// through parseFile, so a record inside a branch body lands in the branch's
// stmts slice instead of File.Statements.
func parseRecordStandalone(p *parser) (*Record, error) {
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

// parseConditionalStandalone parses a nested conditional inside a branch
// body. It does NOT resolve Active recursively against the enclosing file
// (the inner branches' statements are not visible at the file level until
// a consumer walks Active), so the inner conditional's Active is left as
// -1 — a future scope can extend resolution. For the user-prompt scope,
// nested conditionals don't need to be evaluated at parse time.
func parseConditionalStandalone(p *parser) (*Conditional, error) {
	cond := &Conditional{Active: -1}
	var err error
	for action := parseConditionalIf; action != nil && err == nil; {
		action, err = action(p, cond)
	}
	if err != nil {
		return nil, err
	}
	return cond, nil
}

// pickActiveBranch evaluates each branch's condition against f's
// previously-declared records (last-wins per spec §Lookup precedence) and
// returns the first index that evaluates to true. Returns -1 if no branch
// takes. The "else" branch (Condition == nil) always matches when reached.
func pickActiveBranch(cond *Conditional, f *File) int {
	for i, br := range cond.Branches {
		if br.Condition == nil {
			return i // else branch
		}
		v, ok := evalBoolExpression(br.Condition, f)
		if !ok {
			// non-static — skip per the user-prompt's "stub" allowance.
			continue
		}
		if v {
			return i
		}
	}
	return -1
}

// evalBoolExpression evaluates expr to a literal bool using only the
// previously-declared records in f. Supported forms:
//
//   - BoolLiteral
//   - Reference to a record whose value is a BoolLiteral (last-wins)
//   - EqualExpr where both sides reduce to BoolLiteral
//
// Returns (value, true) on success or (false, false) if expr is not a
// constant-foldable bool. This is intentionally a stub: the user prompt
// scopes "full Expression resolution" out.
func evalBoolExpression(expr Expression, f *File) (bool, bool) {
	switch e := expr.(type) {
	case *BoolLiteral:
		return e.Value, true
	case *Reference:
		rec, ok := lookupRecord(f, e.Name)
		if !ok {
			return false, false
		}
		bl, ok := rec.Value.(*BoolLiteral)
		if !ok {
			return false, false
		}
		return bl.Value, true
	case *EqualExpr:
		l, ok := evalBoolExpression(e.Left, f)
		if !ok {
			return false, false
		}
		r, ok := evalBoolExpression(e.Right, f)
		if !ok {
			return false, false
		}
		return l == r, true
	}
	return false, false
}

// lookupRecord returns the most recently declared *Record with key name in
// f.Statements (last-wins per spec §Record key uniqueness). It does not
// descend into Conditional branches — only top-level records are visible
// to a top-level conditional in this scope.
func lookupRecord(f *File, name string) (*Record, bool) {
	for i := len(f.Statements) - 1; i >= 0; i-- {
		rec, ok := f.Statements[i].(*Record)
		if !ok {
			continue
		}
		if rec.Key == name {
			return rec, true
		}
	}
	return nil, false
}


// Parse reads a KVRX file from r.
func Parse(r io.Reader) (*File, error) {
	next, stop := iter.Pull2(Tokenize(r))
	defer stop()

	p := &parser{
		stop: stop,
	}
	p.next = func() (Token, error, bool) {
		if p.peeked != nil {
			tok := *p.peeked
			p.peeked = nil
			return tok, nil, true
		}
		tok, err, ok := next()
		return tok, err, ok
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
