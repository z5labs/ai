// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

import (
	"encoding/binary"
	"io"
)

// countingWriter wraps an io.Writer and tallies the number of bytes written.
// The encoder uses the running total as the byte offset reported in
// OffsetError, mirroring countingReader on the decode side.
type countingWriter struct {
	w io.Writer
	n int64
}

// Write delegates to the underlying writer and increments n by the number of
// bytes actually written.
func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

// encoder is the internal push-based writer. It owns the io.Writer (wrapped
// in a countingWriter for offset tracking) and the byte order. Per-structure
// writes live as writeX methods.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

// newEncoder constructs an encoder around w, defaulting to binary.BigEndian.
// Must match the decoder's byte order — round-trip tests will fail in
// confusing ways if the two disagree.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         &countingWriter{w: w},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr is the encode-side mirror of decoder.wrapErr. Same chain shape
// (FieldError -> OffsetError -> leaf), offset sourced from the counting
// writer instead of the counting reader.
func (e *encoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{
		Field: field,
		Err: &OffsetError{
			Offset: e.w.n,
			Err:    err,
		},
	}
}

// writeFile writes the top-level File structure. This is the stub; replace
// with real writes against the wire format described in SPEC.md.
func (e *encoder) writeFile(f *File) error {
	return e.wrapErr("File", errUnimplemented)
}

// Encode writes f to w as a TLV file.
//
// While the package is still scaffolded, Encode always returns an error whose
// chain is FieldError{Field: "File"} -> OffsetError{Offset: 0} ->
// errUnimplemented. The chain shape is real even though the bytes aren't.
func Encode(w io.Writer, f *File) error {
	e := newEncoder(w)
	return e.writeFile(f)
}
