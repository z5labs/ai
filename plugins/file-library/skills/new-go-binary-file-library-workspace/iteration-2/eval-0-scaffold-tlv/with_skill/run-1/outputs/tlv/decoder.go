// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

import (
	"encoding/binary"
	"io"
)

// countingReader wraps an io.Reader and tallies the number of bytes read.
// The decoder uses the running total as the byte offset reported in
// OffsetError, so individual readX methods don't have to thread an offset
// counter manually.
type countingReader struct {
	r io.Reader
	n int64
}

// Read delegates to the underlying reader and increments n by the number of
// bytes actually read (per io.Reader's contract, even on a non-nil error).
func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// decoder is the internal pull-based reader. It owns the io.Reader (wrapped
// in a countingReader for offset tracking) and the byte order. Per-structure
// reads live as readX methods.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

// newDecoder constructs a decoder around r, defaulting to binary.BigEndian.
// Most network and container formats are big-endian; change this constant if
// the real TLV spec turns out to be little-endian. The encoder must agree.
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         &countingReader{r: r},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every decode error through a uniform FieldError ->
// OffsetError -> leaf chain. Call sites use d.wrapErr("Header.Length", err)
// rather than constructing FieldError/OffsetError directly, so the offset
// always reflects the counting reader's current position.
func (d *decoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{
		Field: field,
		Err: &OffsetError{
			Offset: d.r.n,
			Err:    err,
		},
	}
}

// readFile reads the top-level File structure. This is the stub; replace with
// real reads against the wire format described in SPEC.md.
func (d *decoder) readFile() (*File, error) {
	return nil, d.wrapErr("File", errUnimplemented)
}

// Decode reads a single TLV file from r and returns the typed AST.
//
// While the package is still scaffolded, Decode always returns an error whose
// chain is FieldError{Field: "File"} -> OffsetError{Offset: 0} ->
// errUnimplemented. The chain shape is real even though the bytes aren't.
func Decode(r io.Reader) (*File, error) {
	d := newDecoder(r)
	return d.readFile()
}
