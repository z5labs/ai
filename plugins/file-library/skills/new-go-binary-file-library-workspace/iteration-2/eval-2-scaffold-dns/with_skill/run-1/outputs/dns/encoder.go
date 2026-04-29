// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"encoding/binary"
	"io"
)

// countingWriter wraps an io.Writer and tallies the total number of bytes
// written. The encoder uses the running count as the offset reported in
// OffsetError so error sites don't have to do their own bookkeeping.
type countingWriter struct {
	w io.Writer
	n int64
}

// Write delegates to the underlying writer and increments n by the number
// of bytes actually written, regardless of whether an error was returned.
func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n)
	return n, err
}

// encoder is the symmetric counterpart of decoder: it owns the io.Writer
// and the byte order used to lay down multi-byte fields. Methods are named
// writeX (writeFile, writeHeader, ...).
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

// newEncoder wraps w in a countingWriter and defaults to network byte
// order (big-endian) — the same default as the decoder. They must always
// agree.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         &countingWriter{w: w},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr mirrors decoder.wrapErr: every encode error funnels through here
// so the chain stays FieldError -> OffsetError -> leaf, with the offset
// pulled from the counting writer.
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

// writeFile is a placeholder root writer. The real implementation will
// write the DNS header followed by the question / answer / authority /
// additional sections. Until then it returns the unimplemented sentinel
// through the canonical error chain.
func (e *encoder) writeFile(f *File) error {
	_ = f
	return e.wrapErr("File", errUnimplemented)
}

// Encode writes the DNS wire-format representation of f to w. It is the
// public entry point of the encoder.
func Encode(w io.Writer, f *File) error {
	e := newEncoder(w)
	return e.writeFile(f)
}
