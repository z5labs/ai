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

// UnterminatedStringError is returned when a string literal lacks its closing
// quote — either end-of-file was reached or a literal newline was encountered
// inside the quotes. Pos carries the opening-quote position.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when the tokenizer sees `\` followed by a
// character that is not one of the recognised escapes (\\, \", \n, \t).
type InvalidEscapeError struct {
	Pos  Pos
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

// yieldErr returns an action that yields a single error token and stops the
// stream. Used by every error path so the convention is consistent.
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

// isIdentPart reports whether r may continue an identifier.
func isIdentPart(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

// tokenize is the top-level dispatch action.
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
	case unicode.IsSpace(r):
		return tokenize
	case isIdentStart(r):
		t.backup()
		return tokenizeIdentifier
	case r == '"':
		return tokenizeString(t.prevPos)
	case r == '=' || r == '{' || r == '}' || r == ';':
		pos := t.prevPos
		ch := r
		if !yield(Token{Pos: pos, Type: TokenSymbol, Value: string(ch)}, nil) {
			return nil
		}
		return tokenize
	}
	return yieldErr(&UnexpectedCharacterError{Pos: t.prevPos, Char: r})
}

// tokenizeIdentifier consumes a run of identifier characters and yields a
// single TokenIdentifier. Caller must have backed up so the first identifier
// rune is still in the reader.
func tokenizeIdentifier(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	var b strings.Builder
	startPos := t.pos
	for {
		r, err := t.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if !isIdentPart(r) {
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

// tokenizeString returns a closure that consumes the body of a string literal
// (the opening `"` has already been consumed) and yields a single TokenString
// with the decoded value. startPos is the position of the opening quote.
func tokenizeString(startPos Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		for {
			r, err := t.next()
			if err == io.EOF {
				yield(Token{}, &UnterminatedStringError{Pos: startPos})
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if r == '\n' {
				yield(Token{}, &UnterminatedStringError{Pos: startPos})
				return nil
			}
			if r == '"' {
				if !yield(Token{Pos: startPos, Type: TokenString, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if r == '\\' {
				esc, err := t.next()
				if err == io.EOF {
					yield(Token{}, &UnterminatedStringError{Pos: startPos})
					return nil
				}
				if err != nil {
					yield(Token{}, err)
					return nil
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
					yield(Token{}, &InvalidEscapeError{Pos: t.prevPos, Char: esc})
					return nil
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
