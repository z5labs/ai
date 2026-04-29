// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"encoding/binary"
	"io"
)

// countingReader wraps an io.Reader and tallies the total number of bytes
// read. The decoder uses the running count as the offset reported in
// OffsetError so error sites don't have to do their own bookkeeping.
type countingReader struct {
	r io.Reader
	n int64
}

// Read delegates to the underlying reader and increments n by the number of
// bytes actually read, regardless of whether an error was returned.
func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.n += int64(n)
	return n, err
}

// decoder owns the io.Reader and the byte order used to interpret
// multi-byte fields. Methods are named readX where X is the structure
// being read (readFile, readHeader, ...).
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

// newDecoder wraps r in a countingReader and defaults to network byte
// order (big-endian), which is what DNS uses per RFC 1035.
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         &countingReader{r: r},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr is the single funnel for every decode error. It produces the
// uniform chain FieldError -> OffsetError -> leaf so callers can use
// errors.Is on the leaf, errors.As on *FieldError for the field path, and
// errors.As on *OffsetError for the byte offset. Returns nil when err is
// nil so call sites can stay terse.
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

// readFile is a placeholder root reader. The real implementation will read
// the DNS header followed by the question / answer / authority / additional
// sections. Until then it returns the unimplemented sentinel through the
// canonical error chain.
func (d *decoder) readFile() (*File, error) {
	return nil, d.wrapErr("File", errUnimplemented)
}

// Decode reads a DNS wire-format message from r and returns the parsed
// File. It is the public entry point of the decoder.
func Decode(r io.Reader) (*File, error) {
	d := newDecoder(r)
	return d.readFile()
}
