// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// errUnimplemented is the placeholder error returned by the scaffold's stub
// methods. Real implementations should remove this and return real wire-level
// errors instead.
var errUnimplemented = errors.New("unimplemented")

// UnexpectedEOFError wraps io.ErrUnexpectedEOF with the name of the field
// that ran out of bytes. This is the recommended pattern for surfacing
// truncated-input errors with structural context.
type UnexpectedEOFError struct {
	Field string
}

// Error implements the error interface.
func (e *UnexpectedEOFError) Error() string {
	return fmt.Sprintf("dns: unexpected EOF reading %s", e.Field)
}

// Unwrap returns io.ErrUnexpectedEOF so callers can match with errors.Is.
func (e *UnexpectedEOFError) Unwrap() error { return io.ErrUnexpectedEOF }

// decoder is the internal pull-based reader for the DNS wire format. It owns
// the io.Reader and the byte order, mirroring the encoder.
type decoder struct {
	r         io.Reader
	byteOrder binary.ByteOrder
}

// newDecoder constructs a decoder defaulting to the network byte order
// (big-endian), matching how DNS messages are framed on the wire (RFC 1035
// section 2.3.2).
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         r,
		byteOrder: binary.BigEndian,
	}
}

// readFile is the entry point for decoding a complete DNS message. The
// scaffold returns an unimplemented error so the test suite has a stable
// failure mode to invert once the real decoder lands.
func (d *decoder) readFile() (*File, error) {
	return &File{}, fmt.Errorf("decoding File: %w", errUnimplemented)
}

// Decode reads a DNS wire-format message from r and returns the decoded
// File. It is the public entry point of the decoder.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
