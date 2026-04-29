package dns

import "io"

// TokenKind enumerates the kinds of tokens produced by the tokenizer when
// scanning a DNS wire-format byte stream.
type TokenKind int

const (
	// TokenInvalid is the zero-value kind and indicates an uninitialized
	// or malformed token.
	TokenInvalid TokenKind = iota
	// TokenHeader is the 12-byte fixed header at the start of every DNS
	// message.
	TokenHeader
	// TokenQuestion is a single entry in the question section.
	TokenQuestion
	// TokenResourceRecord is a single entry in the answer, authority, or
	// additional section.
	TokenResourceRecord
	// TokenEOF signals the end of the input stream.
	TokenEOF
)

// Token is a single unit of input produced by the tokenizer.
type Token struct {
	Kind  TokenKind
	Bytes []byte
}

// Tokenizer reads a DNS wire-format byte stream and yields Token values.
type Tokenizer struct {
	r io.Reader
}

// NewTokenizer constructs a Tokenizer that reads from r.
func NewTokenizer(r io.Reader) *Tokenizer {
	return &Tokenizer{r: r}
}

// Next returns the next Token from the input. When the input is exhausted
// it returns a TokenEOF token and io.EOF.
//
// Boilerplate: implementation is filled in by the
// implement-binary-file-library agent.
func (t *Tokenizer) Next() (Token, error) {
	return Token{Kind: TokenEOF}, io.EOF
}
