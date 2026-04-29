// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// errUnimplemented is a placeholder error used by the scaffolded readFile()
// stub. Remove it once the real decoder logic is implemented.
var errUnimplemented = errors.New("unimplemented")

// UnexpectedEOFError carries field-level context for an io.ErrUnexpectedEOF
// encountered while decoding. It demonstrates the context-rich error wrapping
// pattern used throughout the decoder: every non-trivial error should identify
// the field that was being read when the failure occurred.
type UnexpectedEOFError struct {
	// Field is the dotted path to the field whose read was truncated.
	Field string
	// Err is the underlying I/O error (typically io.ErrUnexpectedEOF).
	Err error
}

// Error implements the error interface.
func (e *UnexpectedEOFError) Error() string {
	return fmt.Sprintf("unexpected EOF reading %s: %v", e.Field, e.Err)
}

// Unwrap exposes the underlying error so errors.Is and errors.As work.
func (e *UnexpectedEOFError) Unwrap() error {
	return e.Err
}

// decoder is the internal pull-based reader. It owns the io.Reader and the
// byte order used for every binary.Read call.
type decoder struct {
	r         io.Reader
	byteOrder binary.ByteOrder
}

// newDecoder constructs a decoder reading from r. Default byte order is
// binary.BigEndian, which matches most network and container formats.
// Change here once SPEC.md confirms the format's byte order.
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         r,
		byteOrder: binary.BigEndian,
	}
}

// readFile is the per-structure reader for the top-level File type. Replace
// the unimplemented stub with real decoding logic — typically a sequence of
// binary.Read calls for fixed-size fields and io.ReadFull calls for
// length-prefixed payloads.
func (d *decoder) readFile() (*File, error) {
	return &File{}, fmt.Errorf("decoding File: %w", errUnimplemented)
}

// Decode reads a TLV File from r. It is the only public entry point on the
// decoder; if the format ever needs options (a strictness flag, a version
// pin), add a separate DecodeWithOptions function rather than overloading this
// one.
func Decode(r io.Reader) (*File, error) {
	d := newDecoder(r)
	return d.readFile()
}
