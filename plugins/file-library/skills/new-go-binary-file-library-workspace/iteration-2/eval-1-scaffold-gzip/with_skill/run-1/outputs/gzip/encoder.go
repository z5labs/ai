// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package gzip

import (
	"encoding/binary"
	"io"
)

// countingWriter wraps an io.Writer and tallies the total number of bytes
// written in n. The encoder uses n as the current byte offset when wrapping
// leaf errors with OffsetError, mirroring the decoder's countingReader.
type countingWriter struct {
	w io.Writer
	n int64
}

// Write delegates to the underlying writer and increments n by the number of
// bytes successfully written. The byte count is incremented even on a short
// write so the offset reported in OffsetError points at the boundary the
// write failed on.
func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

// encoder is the internal push-based writer. It owns the io.Writer (wrapped
// in a countingWriter so we always know the current byte offset) and the byte
// order used for every binary.Write call. The byte order MUST match the
// decoder's, otherwise round-trip tests will fail in confusing ways.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

// newEncoder constructs an encoder writing to w. Default byte order is
// binary.BigEndian. NOTE: the real gzip format is little-endian; flip this
// (and the matching constant in newDecoder) once SPEC.md is filled in.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         &countingWriter{w: w},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every leaf error through the FieldError -> OffsetError ->
// leaf chain, using e.w.n as the current offset. Every writeX method should
// return its errors via this helper; constructing FieldError or OffsetError
// directly outside this method risks the offset drifting out of sync with
// countingWriter.n.
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

// writeFile is the per-structure writer for the top-level File type. Replace
// the unimplemented stub with real encoding logic — typically a sequence of
// binary.Write calls for fixed-size fields and e.w.Write calls for
// length-prefixed payloads, each routed through e.wrapErr.
func (e *encoder) writeFile(_ *File) error {
	return e.wrapErr("File", errUnimplemented)
}

// Encode writes a gzip File to w. It is the only public entry point on the
// encoder; if the format ever needs options (a compression level, a flag),
// add a separate EncodeWithOptions function rather than overloading this one.
func Encode(w io.Writer, f *File) error {
	e := newEncoder(w)
	return e.writeFile(f)
}
