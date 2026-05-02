package kvr

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"strings"
	"unicode"
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

// UnterminatedStringError is returned when a string literal is not closed
// before EOF or before a literal newline.
type UnterminatedStringError struct {
	Pos Pos // position of the opening quote
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a string literal contains a backslash
// followed by a rune that is not a recognised escape character.
type InvalidEscapeError struct {
	Pos  Pos // position of the backslash
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape sequence \\%c at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
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

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// yieldErr is the standard error-and-stop ending. It surfaces err to the
// consumer and ends the iteration.
func yieldErr(err error) tokenizerAction {
	return func(_ *tokenizer, yield func(Token, error) bool) tokenizerAction {
		yield(Token{}, err)
		return nil
	}
}

// isIdentStart reports whether r can start an identifier.
func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isIdentCont reports whether r can continue an identifier.
func isIdentCont(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

// tokenize is the top-level dispatch action. It peeks one rune and delegates
// to a specialised action (or returns nil at EOF).
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
	case r == '\n' || unicode.IsSpace(r):
		// Skip whitespace and chain back to dispatch. End-of-line is just
		// whitespace per SPEC.md Overview.
		return tokenize
	case r == '#':
		return tokenizeComment(t.prevPos)
	case r == '"':
		return tokenizeString(t.prevPos)
	case r == '=' || r == '{' || r == '}' || r == ';':
		pos := t.prevPos
		val := string(r)
		if !yield(Token{Pos: pos, Type: TokenSymbol, Value: val}, nil) {
			return nil
		}
		return tokenize
	case r >= '0' && r <= '9':
		t.backup()
		return tokenizeNumber
	case isIdentStart(r):
		t.backup()
		return tokenizeIdentifier
	}
	return yieldErr(&UnexpectedCharacterError{Pos: t.prevPos, Char: r})
}

// tokenizeIdentifier reads runs of identifier-continuation runes and yields
// a single TokenIdentifier.
func tokenizeIdentifier(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err != nil {
		// EOF before any rune is impossible — dispatch already saw one.
		return yieldErr(err)
	}
	startPos := t.prevPos
	var b strings.Builder
	b.WriteRune(r)
	for {
		r, err := t.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return yieldErr(err)
		}
		if !isIdentCont(r) {
			t.backup()
			break
		}
		b.WriteRune(r)
	}
	if !yield(Token{Pos: startPos, Type: TokenIdentifier, Value: b.String()}, nil) {
		return nil
	}
	return tokenize
}

// tokenizeNumber reads a run of ASCII digits and yields a single TokenNumber.
func tokenizeNumber(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	r, err := t.next()
	if err != nil {
		return yieldErr(err)
	}
	startPos := t.prevPos
	var b strings.Builder
	b.WriteRune(r)
	for {
		r, err := t.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return yieldErr(err)
		}
		if r < '0' || r > '9' {
			t.backup()
			break
		}
		b.WriteRune(r)
	}
	if !yield(Token{Pos: startPos, Type: TokenNumber, Value: b.String()}, nil) {
		return nil
	}
	return tokenize
}

// tokenizeComment captures everything from the rune after '#' up to (but not
// including) the next newline or EOF, with leading horizontal whitespace
// stripped from the value.
func tokenizeComment(startPos Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		// Strip leading horizontal whitespace.
		stripping := true
		for {
			r, err := t.next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return yieldErr(err)
			}
			if r == '\n' {
				// Don't consume the newline as part of the comment value;
				// rewind so dispatch sees it as whitespace.
				t.backup()
				break
			}
			if stripping && (r == ' ' || r == '\t') {
				continue
			}
			stripping = false
			b.WriteRune(r)
		}
		if !yield(Token{Pos: startPos, Type: TokenComment, Value: b.String()}, nil) {
			return nil
		}
		return tokenize
	}
}

// tokenizeString reads a quoted-string literal, handling backslash escapes,
// rejecting literal newlines and EOF inside the string.
func tokenizeString(startPos Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		for {
			r, err := t.next()
			if err == io.EOF {
				return yieldErr(&UnterminatedStringError{Pos: startPos})
			}
			if err != nil {
				return yieldErr(err)
			}
			if r == '\n' {
				return yieldErr(&UnterminatedStringError{Pos: startPos})
			}
			if r == '"' {
				if !yield(Token{Pos: startPos, Type: TokenString, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if r == '\\' {
				escPos := t.prevPos
				esc, err := t.next()
				if err == io.EOF {
					return yieldErr(&UnterminatedStringError{Pos: startPos})
				}
				if err != nil {
					return yieldErr(err)
				}
				switch esc {
				case '\\':
					b.WriteRune('\\')
				case '"':
					b.WriteRune('"')
				case 'n':
					b.WriteRune('\n')
				case 't':
					b.WriteRune('\t')
				default:
					return yieldErr(&InvalidEscapeError{Pos: escPos, Char: esc})
				}
				continue
			}
			b.WriteRune(r)
		}
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
