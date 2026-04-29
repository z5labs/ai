// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package gzip

import (
	"encoding/binary"
	"io"
)

// countingReader wraps an io.Reader and tallies the total number of bytes read
// in n. The decoder uses n as the current byte offset when wrapping leaf
// errors with OffsetError, so individual readX methods don't need to do any
// manual accounting.
type countingReader struct {
	r io.Reader
	n int64
}

// Read delegates to the underlying reader and increments n by the number of
// bytes successfully read. The byte count is incremented even when err is
// non-nil (e.g. a partial read followed by io.ErrUnexpectedEOF) so the offset
// reported in OffsetError points at the boundary the read failed on.
func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// decoder is the internal pull-based reader. It owns the io.Reader (wrapped
// in a countingReader so we always know the current byte offset) and the byte
// order used for every binary.Read call.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

// newDecoder constructs a decoder reading from r. Default byte order is
// binary.BigEndian. NOTE: the real gzip format is little-endian; flip this
// (and the matching constant in newEncoder) once SPEC.md is filled in. The
// decoder and encoder must agree, otherwise round-trip tests will fail.
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         &countingReader{r: r},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every leaf error through the FieldError -> OffsetError ->
// leaf chain. Every readX method should return its errors via this helper —
// constructing FieldError or OffsetError directly outside this method risks
// the offset drifting out of sync with countingReader.n.
//
// Pass a dotted field path ("Header.Length") so callers can grep, and pass
// the underlying error untouched so errors.Is(err, sentinel) keeps working.
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

// readFile is the per-structure reader for the top-level File type. Replace
// the unimplemented stub with real decoding logic — typically a sequence of
// binary.Read calls for fixed-size fields and io.ReadFull calls for
// length-prefixed payloads, each routed through d.wrapErr.
func (d *decoder) readFile() (*File, error) {
	return nil, d.wrapErr("File", errUnimplemented)
}

// Decode reads a gzip File from r. It is the only public entry point on the
// decoder; if the format ever needs options (a strictness flag, a member
// limit), add a separate DecodeWithOptions function rather than overloading
// this one.
func Decode(r io.Reader) (*File, error) {
	d := newDecoder(r)
	return d.readFile()
}
