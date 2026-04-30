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

// UnterminatedStringError is returned when a string literal does not close
// with `"` before end-of-file or before a literal newline.
type UnterminatedStringError struct {
	Pos Pos // position of the opening quote
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a `\` is followed by a rune that is
// not one of the recognised escape characters (`\\`, `\"`, `\n`, `\t`).
type InvalidEscapeError struct {
	Pos  Pos // position of the backslash
	Char rune
}

func (e *InvalidEscapeError) Error() string {
	return fmt.Sprintf("invalid escape \\%q at %d:%d", e.Char, e.Pos.Line, e.Pos.Column)
}

// tokenizer holds the reader and current position.
type tokenizer struct {
	r       *bufio.Reader
	pos     Pos
	prevPos Pos  // pos before the most recent next() call — used by backup
	lastR   rune // last rune read by next, for backup bookkeeping
	hasLast bool
}

// next advances the cursor by one rune and updates pos.
func (t *tokenizer) next() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	t.prevPos = t.pos
	t.lastR = r
	t.hasLast = true
	if r == '\n' {
		t.pos.Line++
		t.pos.Column = 1
	} else {
		t.pos.Column++
	}
	return r, nil
}

// backup rewinds the last rune read by next.
func (t *tokenizer) backup() {
	if !t.hasLast {
		return
	}
	_ = t.r.UnreadRune()
	t.pos = t.prevPos
	t.hasLast = false
}

// peek returns the next rune without consuming it (or io.EOF/err).
func (t *tokenizer) peek() (rune, error) {
	r, _, err := t.r.ReadRune()
	if err != nil {
		return 0, err
	}
	_ = t.r.UnreadRune()
	return r, nil
}

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

// tokenize is the top-level dispatch action. Reads one rune, dispatches on
// its category to a specialised action, then chains back to itself.
func tokenize(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
	// skip insignificant whitespace (newlines included — they are just whitespace)
	for {
		r, err := t.peek()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			yield(Token{}, err)
			return nil
		}
		if !unicode.IsSpace(r) {
			break
		}
		if _, err := t.next(); err != nil {
			yield(Token{}, err)
			return nil
		}
	}

	startPos := t.pos
	r, err := t.next()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		yield(Token{}, err)
		return nil
	}

	switch {
	case r == '#':
		return tokenizeComment(startPos)
	case r == '"':
		return tokenizeString(startPos)
	case r == '=' || r == '{' || r == '}' || r == ';':
		if !yield(Token{Pos: startPos, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	case isIdentStart(r):
		return tokenizeIdentifier(startPos, r)
	case unicode.IsDigit(r):
		return tokenizeNumber(startPos, r)
	}
	yield(Token{}, &UnexpectedCharacterError{Pos: startPos, Char: r})
	return nil
}

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentCont(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

// tokenizeIdentifier reads identifier continuation runes after `first` was
// already consumed at startPos.
func tokenizeIdentifier(startPos Pos, first rune) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		b.WriteRune(first)
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
			b.WriteRune(r)
		}
		if !yield(Token{Pos: startPos, Type: TokenIdentifier, Value: b.String()}, nil) {
			return nil
		}
		return tokenize
	}
}

// tokenizeNumber reads digit continuation runes after `first` was already
// consumed at startPos.
func tokenizeNumber(startPos Pos, first rune) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		b.WriteRune(first)
		for {
			r, err := t.next()
			if err == io.EOF {
				break
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !unicode.IsDigit(r) {
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
}

// tokenizeComment reads from after the `#` to the next `\n` or EOF. Leading
// horizontal whitespace and the trailing newline are excluded from Value.
func tokenizeComment(startPos Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		// skip leading horizontal whitespace
		for {
			r, err := t.peek()
			if err == io.EOF {
				if !yield(Token{Pos: startPos, Type: TokenComment, Value: ""}, nil) {
					return nil
				}
				return tokenize
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if r == '\n' {
				if !yield(Token{Pos: startPos, Type: TokenComment, Value: ""}, nil) {
					return nil
				}
				return tokenize
			}
			if r != ' ' && r != '\t' {
				break
			}
			if _, err := t.next(); err != nil {
				yield(Token{}, err)
				return nil
			}
		}
		var b strings.Builder
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
				// don't consume the newline as part of the comment value
				break
			}
			b.WriteRune(r)
		}
		if !yield(Token{Pos: startPos, Type: TokenComment, Value: b.String()}, nil) {
			return nil
		}
		return tokenize
	}
}

// tokenizeString reads runes after the opening `"` was already consumed at
// startPos, decoding `\\`, `\"`, `\n`, `\t` escapes.
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
				escapePos := t.pos
				escapePos.Column-- // position of the backslash
				er, err := t.next()
				if err == io.EOF {
					yield(Token{}, &UnterminatedStringError{Pos: startPos})
					return nil
				}
				if err != nil {
					yield(Token{}, err)
					return nil
				}
				switch er {
				case '\\':
					b.WriteRune('\\')
				case '"':
					b.WriteRune('"')
				case 'n':
					b.WriteRune('\n')
				case 't':
					b.WriteRune('\t')
				default:
					yield(Token{}, &InvalidEscapeError{Pos: escapePos, Char: er})
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

