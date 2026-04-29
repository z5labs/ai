package dns

import "io"

// Printer writes a Message AST out as a DNS wire-format byte stream.
type Printer struct {
	w io.Writer
}

// NewPrinter constructs a Printer that writes to w.
func NewPrinter(w io.Writer) *Printer {
	return &Printer{w: w}
}

// Print serializes m to the underlying writer in DNS wire format.
//
// Boilerplate: implementation is filled in by the
// implement-binary-file-library agent.
func (p *Printer) Print(m Message) error {
	return nil
}

// Encode is a convenience wrapper that serializes m to w.
func Encode(w io.Writer, m Message) error {
	return NewPrinter(w).Print(m)
}
