package kvr

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

// UnterminatedStringError is returned when a quoted string runs to end-of-file
// or a literal newline before its closing quote. Pos points to the opening
// quote.
type UnterminatedStringError struct {
	Pos Pos
}

func (e *UnterminatedStringError) Error() string {
	return fmt.Sprintf("unterminated string starting at %d:%d", e.Pos.Line, e.Pos.Column)
}

// InvalidEscapeError is returned when a backslash inside a string literal is
// followed by an unrecognised escape character. Pos points to the backslash.
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

// tokenizerAction is a step in the tokenizer state machine.
// Returning nil ends iteration.
type tokenizerAction func(t *tokenizer, yield func(Token, error) bool) tokenizerAction

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentCont(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHorizontalWS(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

// tokenize is the top-level dispatch action. The implementer extends the switch
// to recognise new token types and returns the appropriate specialised action.
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
	case r == '\n' || isHorizontalWS(r):
		return tokenize
	case isIdentStart(r):
		startPos := t.prevPos
		return tokenizeIdent(startPos, string(r))
	case r == '"':
		startPos := t.prevPos
		return tokenizeString(startPos)
	case r == '=' || r == '{' || r == '}' || r == ';':
		startPos := t.prevPos
		if !yield(Token{Pos: startPos, Type: TokenSymbol, Value: string(r)}, nil) {
			return nil
		}
		return tokenize
	case isDigit(r):
		startPos := t.prevPos
		return tokenizeNumber(startPos, string(r))
	case r == '#':
		startPos := t.prevPos
		return tokenizeComment(startPos)
	default:
		yield(Token{}, &UnexpectedCharacterError{Pos: t.prevPos, Char: r})
		return nil
	}
}

// tokenizeIdent collects an identifier starting at startPos. The first rune
// has already been consumed and is in acc.
func tokenizeIdent(startPos Pos, acc string) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		b.WriteString(acc)
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: startPos, Type: TokenIdentifier, Value: b.String()}, nil) {
					return nil
				}
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isIdentCont(r) {
				t.backup()
				if !yield(Token{Pos: startPos, Type: TokenIdentifier, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			b.WriteRune(r)
		}
	}
}

// tokenizeNumber collects digits starting at startPos. The first rune has
// already been consumed and is in acc.
func tokenizeNumber(startPos Pos, acc string) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		b.WriteString(acc)
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: startPos, Type: TokenNumber, Value: b.String()}, nil) {
					return nil
				}
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if !isDigit(r) {
				t.backup()
				if !yield(Token{Pos: startPos, Type: TokenNumber, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			b.WriteRune(r)
		}
	}
}

// tokenizeString collects a quoted string. The opening quote has already been
// consumed; startPos is the position of that opening quote.
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
			switch r {
			case '"':
				if !yield(Token{Pos: startPos, Type: TokenString, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			case '\n':
				yield(Token{}, &UnterminatedStringError{Pos: startPos})
				return nil
			case '\\':
				escPos := t.prevPos
				next, err := t.next()
				if err == io.EOF {
					yield(Token{}, &UnterminatedStringError{Pos: startPos})
					return nil
				}
				if err != nil {
					yield(Token{}, err)
					return nil
				}
				switch next {
				case '\\':
					b.WriteRune('\\')
				case '"':
					b.WriteRune('"')
				case 'n':
					b.WriteRune('\n')
				case 't':
					b.WriteRune('\t')
				default:
					yield(Token{}, &InvalidEscapeError{Pos: escPos, Char: next})
					return nil
				}
			default:
				b.WriteRune(r)
			}
		}
	}
}

// tokenizeComment collects a line comment. The leading '#' has already been
// consumed; startPos is the position of that '#'. Leading horizontal whitespace
// after the '#' is stripped from Value, and the trailing newline is not part
// of Value.
func tokenizeComment(startPos Pos) tokenizerAction {
	return func(t *tokenizer, yield func(Token, error) bool) tokenizerAction {
		var b strings.Builder
		strippedLeading := false
		for {
			r, err := t.next()
			if err == io.EOF {
				if !yield(Token{Pos: startPos, Type: TokenComment, Value: b.String()}, nil) {
					return nil
				}
				return nil
			}
			if err != nil {
				yield(Token{}, err)
				return nil
			}
			if r == '\n' {
				if !yield(Token{Pos: startPos, Type: TokenComment, Value: b.String()}, nil) {
					return nil
				}
				return tokenize
			}
			if !strippedLeading && (r == ' ' || r == '\t') {
				continue
			}
			strippedLeading = true
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
