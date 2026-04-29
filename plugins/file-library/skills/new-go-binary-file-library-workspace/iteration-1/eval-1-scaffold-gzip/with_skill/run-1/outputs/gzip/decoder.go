// Copyright (c) 2026 Z5labs and Contributors
//
// Licensed under the MIT License. See LICENSE file in the project root
// for full license information.

package gzip

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// UnexpectedEOFError wraps io.ErrUnexpectedEOF with structural context so
// the caller knows which field ran out of bytes during decoding.
type UnexpectedEOFError struct {
	Field string
	Err   error
}

// Error implements the error interface.
func (e *UnexpectedEOFError) Error() string {
	return fmt.Sprintf("gzip: unexpected EOF while reading %s: %v", e.Field, e.Err)
}

// Unwrap returns the wrapped error so errors.Is(err, io.ErrUnexpectedEOF) holds.
func (e *UnexpectedEOFError) Unwrap() error {
	return e.Err
}

// decoder is the internal pull-based reader. It owns the io.Reader and the
// byte order so individual reads stay terse.
type decoder struct {
	r         io.Reader
	byteOrder binary.ByteOrder
}

// newDecoder constructs a decoder that defaults to binary.BigEndian. The
// gzip wire format itself is little-endian, but BigEndian is the default
// per the scaffolding skill so the implementer makes the byte-order choice
// explicitly when wiring up real reads.
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         r,
		byteOrder: binary.BigEndian,
	}
}

// readFile is the per-structure reader for the root File type. Real
// implementations should read the gzip header, optional fields, compressed
// data stream, and trailer here.
func (d *decoder) readFile() (*File, error) {
	err := errors.New("unimplemented")
	return &File{}, fmt.Errorf("decoding File: %w", err)
}

// Decode reads a gzip file from r and returns the parsed File. It is the
// minimal public surface; format options should go on a separate
// DecodeWithOptions function rather than overloading this one.
func Decode(r io.Reader) (*File, error) {
	d := newDecoder(r)
	return d.readFile()
}
