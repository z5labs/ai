package dns

import "io"

// Parser consumes Token values produced by a Tokenizer and builds a Message
// AST.
type Parser struct {
	t *Tokenizer
}

// NewParser constructs a Parser that reads tokens from t.
func NewParser(t *Tokenizer) *Parser {
	return &Parser{t: t}
}

// Parse reads tokens until the end of the input and returns the resulting
// Message.
//
// Boilerplate: implementation is filled in by the
// implement-binary-file-library agent.
func (p *Parser) Parse() (Message, error) {
	return Message{}, nil
}

// Decode is a convenience wrapper that reads a complete DNS message from r.
func Decode(r io.Reader) (Message, error) {
	return NewParser(NewTokenizer(r)).Parse()
}
