package kvrx

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"strings"
)

// Pos identifies a 1-based line/column position in the source.
type Pos struct {
	Line   int
	Column int
}

// TokenType identifies the kind of a Token.
type TokenType int

const (
	TokenInvalid TokenType = iota
	TokenIdentifier
	TokenSymbol
	TokenString
	TokenNumber
	TokenComment
	TokenNewline
)

// String returns a human-readable name for a TokenType. Implementer extends
// this when new token types are added so test failures print readable names.
func (t TokenType) String() string {
	switch t {
	case TokenIdentifier:
		return "IDENT"
	case TokenSymbol:
		return "SYMBOL"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenComment:
		return "COMMENT"
	case TokenNewline:
		return "NEWLINE"
	default:
		return fmt.Sprintf("TokenType(%d)", int(t))
	}
}

// Token is one lexical element produced by the tokenizer.
type Token struct {
	Pos   Pos
	Type  TokenType
	Value string
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%d:%d", t.Type, t.Value, t.Pos.Line, t.Pos.Column)
}

// UnexpectedCharacterError is returned when the tokenizer encounters a rune
// that no action wanted.
type UnexpectedCharacterError struct {
	Pos  Pos
	Char rune
}

func (e *UnexpectedCharacterError) Error() string {
	return fmt.Sprintf("unexpected character %q at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// UnterminatedStringError is returned when a string literal opened with `"`
// reaches a literal newline or end-of-input before the closing `"`.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a `\` inside a string is followed by a
// rune that is not a recognised escape character.
type InvalidEscapeError struct {
	Pos  Pos
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape \\%c at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position. prevPos snapshots the
// position before the most recent next() so backup() can restore it
// (including across newline boundaries) — never reconstruct via column
// arithmetic, since that underflows when the previous next() reset Column
// to 1 after consuming '\n'.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos
	hasPrev bool
}

// next advances the cursor by one rune and updates pos. It snapshots pos
// into prevPos before mutating so backup() can restore it.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prevPos = t.pos
	t.hasPrev = true
	if r == '\n' {
		t.pos.Line++
		t.pos.Column = 1
	} else {
		t.pos.Column++
	}
	return r, nil
}

// backup rewinds the last rune read by next, restoring pos. backup may only
// be called once per next.
func (t *tokenizer) backup() {
	if err := t.r.UnreadRune(); err != nil {
		return
	}
	if t.hasPrev {
		t.pos = t.prevPos
		t.hasPrev = false
	}
}

// peek returns the next rune without consuming it. Returns 0, io.EOF at end.
func (t *tokenizer) peek() (rune, error) {
	r, err := t.next()
	if err != nil {
		return 0, err
	}
	t.backup()
	return r, nil
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldErr emits the error and stops the state machine.
func yieldErr(err error) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		yield(Token{}, err)
		return nil
	}
}

// isIdentStart reports whether r may begin an identifier.
func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isIdentCont reports whether r may continue an identifier.
func isIdentCont(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

// isDigit reports whether r is an ASCII digit.
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// tokenize is the top-level dispatch action. It peeks one rune and dispatches
// to a specialised action; specialised actions chain back here when done.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		yield(Token{}, err)
		return nil
	}
	switch {
	case r == '\n':
		tok := Token{Pos: t.prevPos, Type: TokenNewline, Value: "\n"}
		if !yield(tok, nil) {
			return nil
		}
		return tokenize
	case r == ' ' || r == '\t' || r == '\r':
		return tokenize
	case r == '"':
		return tokenizeString(t.prevPos)
	case r == '#':
		return tokenizeLineComment(t.prevPos)
	case isIdentStart(r):
		t.backup()
		return tokenizeIdentifier
	case isDigit(r):
		t.backup()
		return tokenizeNumber
	case isSingleSymbol(r):
		tok := Token{Pos: t.prevPos, Type: TokenSymbol, Value: string(r)}
		if !yield(tok, nil) {
			return nil
		}
		return tokenize
	case r == '=' || r == '!' || r == '<' || r == '>' || r == '&' || r == '|':
		return tokenizeMaybeTwoCharSymbol(r, t.prevPos)
	}
	yield(Token{}, &UnexpectedCharacterError{Pos: t.prevPos, Char: r})
	return nil
}

// isSingleSymbol matches the set of single-character symbols whose meaning
// never depends on the next rune.
func isSingleSymbol(r rune) bool {
	switch r {
	case '{', '}', '[', ']', '(', ')', ',', ';', ':', '+', '-', '*', '/':
		return true
	}
	return false
}

// tokenizeIdentifier reads an identifier (letters, digits, underscores).
func tokenizeIdentifier(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err != nil {
		// shouldn't happen because dispatch already saw a start rune
		yield(Token{}, err)
		return nil
	}
	start := t.prevPos
	var sb strings.Builder
	sb.WriteRune(r)
	for {
		r, err := t.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if !isIdentCont(r) {
			t.backup()
			break
		}
		sb.WriteRune(r)
	}
	tok := Token{Pos: start, Type: TokenIdentifier, Value: sb.String()}
	if !yield(tok, nil) {
		return nil
	}
	return tokenize
}

// tokenizeNumber reads a (decimal) number. Hex/oct/binary/float are not
// required for the bool+conditional scope; if more forms are needed later,
// add them here.
func tokenizeNumber(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err != nil {
		yield(Token{}, err)
		return nil
	}
	start := t.prevPos
	var sb strings.Builder
	sb.WriteRune(r)
	for {
		r, err := t.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if !isDigit(r) {
			t.backup()
			break
		}
		sb.WriteRune(r)
	}
	tok := Token{Pos: start, Type: TokenNumber, Value: sb.String()}
	if !yield(tok, nil) {
		return nil
	}
	return tokenize
}

// tokenizeString reads a double-quoted string with backslash escapes. start
// is the position of the opening quote.
func tokenizeString(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		for {
			r, err := t.next()
			if err == io.EOF {
				yield(Token{}, &UnterminatedStringError{Pos: start})
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if r == '\n' {
				yield(Token{}, &UnterminatedStringError{Pos: start})
				return nil
			}
			if r == '"' {
				tok := Token{Pos: start, Type: TokenString, Value: sb.String()}
				if !yield(tok, nil) {
					return nil
				}
				return tokenize
			}
			if r == '\\' {
				esc, err := t.next()
				if err == io.EOF {
					yield(Token{}, &UnterminatedStringError{Pos: start})
					return nil
				}
				if err != nil {
					yield(Token{}, err)
					return nil
				}
				switch esc {
				case '\\':
					sb.WriteRune('\\')
				case '"':
					sb.WriteRune('"')
				case 'n':
					sb.WriteRune('\n')
				case 't':
					sb.WriteRune('\t')
				case 'r':
					sb.WriteRune('\r')
				case '0':
					sb.WriteRune(0)
				default:
					yield(Token{}, &InvalidEscapeError{Pos: t.prevPos, Char: esc})
					return nil
				}
				continue
			}
			sb.WriteRune(r)
		}
	}
}

// tokenizeLineComment reads from after `#` to the next newline (or EOF).
// start is the position of the `#`. The newline itself is NOT consumed —
// dispatch picks it up next and emits TokenNewline.
func tokenizeLineComment(start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var sb strings.Builder
		// drop leading horizontal whitespace
		leading := true
		for {
			r, err := t.next()
			if err == io.EOF {
				break
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if r == '\n' {
				t.backup()
				break
			}
			if leading && (r == ' ' || r == '\t') {
				continue
			}
			leading = false
			sb.WriteRune(r)
		}
		tok := Token{Pos: start, Type: TokenComment, Value: sb.String()}
		if !yield(tok, nil) {
			return nil
		}
		return tokenize
	}
}

// tokenizeMaybeTwoCharSymbol handles symbols whose meaning depends on whether
// the next rune extends them: `==`, `!=`, `<=`, `>=`, `&&`, `||`. start is
// the position of the first rune; first is that rune.
func tokenizeMaybeTwoCharSymbol(first rune, start Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		r, err := t.next()
		if err == io.EOF {
			tok := Token{Pos: start, Type: TokenSymbol, Value: string(first)}
			if !yield(tok, nil) {
				return nil
			}
			return tokenize
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		var two string
		switch {
		case first == '=' && r == '=':
			two = "=="
		case first == '!' && r == '=':
			two = "!="
		case first == '<' && r == '=':
			two = "<="
		case first == '>' && r == '=':
			two = ">="
		case first == '&' && r == '&':
			two = "&&"
		case first == '|' && r == '|':
			two = "||"
		}
		if two != "" {
			tok := Token{Pos: start, Type: TokenSymbol, Value: two}
			if !yield(tok, nil) {
				return nil
			}
			return tokenize
		}
		// not a two-char form — back up the second rune and emit single-char
		t.backup()
		tok := Token{Pos: start, Type: TokenSymbol, Value: string(first)}
		if !yield(tok, nil) {
			return nil
		}
		return tokenize
	}
}

// Tokenize streams tokens from r as an iter.Seq2[Token, error].
func Tokenize(r io.Reader) iter.Seq2[Token, error] {
	return func(yield func(Token, error) bool) {
		t := &tokenizer{r: bufio.NewReader(r), pos: Pos{Line: 1, Column: 1}}
		for action := tokenize; action != nil; {
			action = action(t, yield)
		}
	}
}

// silence unused import warnings if helpers go unused in lean builds.
var _ = yieldErr
