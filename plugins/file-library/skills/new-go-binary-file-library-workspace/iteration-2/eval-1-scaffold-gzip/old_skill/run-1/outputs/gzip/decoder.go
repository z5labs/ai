// Copyright (c) 2026 Z5labs and Contributors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// errUnimplemented is returned by stub methods that have not been
// filled in yet. It exists so the placeholder tests can flip from
// "expect error" to "expect success" once the real implementation
// is in place.
var errUnimplemented = errors.New("unimplemented")

// UnexpectedEOFError wraps io.ErrUnexpectedEOF with structural
// context describing which field ran out of bytes. It demonstrates
// the context-rich error wrapping pattern; readX methods should
// return this when binary.Read or io.ReadFull return short.
type UnexpectedEOFError struct {
	Field string
	Err   error
}

// Error implements the error interface.
func (e *UnexpectedEOFError) Error() string {
	return fmt.Sprintf("gzip: unexpected EOF reading %s: %v", e.Field, e.Err)
}

// Unwrap exposes the wrapped error so callers can use errors.Is to
// match against io.ErrUnexpectedEOF.
func (e *UnexpectedEOFError) Unwrap() error {
	return e.Err
}

// decoder owns an io.Reader and the byte order used to interpret
// every multi-byte field in the gzip wire format.
type decoder struct {
	r         io.Reader
	byteOrder binary.ByteOrder
}

// newDecoder constructs a decoder defaulting to binary.BigEndian.
// Note: the real gzip format is little-endian on the wire; the
// implementer should switch this to binary.LittleEndian when filling
// in readFile against the spec.
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         r,
		byteOrder: binary.BigEndian,
	}
}

// readFile is the top-level decode entry point. It is a stub today
// and returns an "unimplemented" error wrapped with the structural
// context the rest of the package uses.
func (d *decoder) readFile() (*File, error) {
	return &File{}, fmt.Errorf("decoding File: %w", errUnimplemented)
}

// Decode reads a gzip File from r.
func Decode(r io.Reader) (*File, error) {
	d := newDecoder(r)
	return d.readFile()
}
