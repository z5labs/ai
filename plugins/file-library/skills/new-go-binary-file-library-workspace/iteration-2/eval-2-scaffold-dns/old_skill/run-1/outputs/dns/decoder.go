// Copyright (c) 2026 z5labs
//
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file or at https://opensource.org/licenses/MIT.

package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// errUnimplemented is returned by stub methods until the implementer fills in
// the real wire-format logic.
var errUnimplemented = errors.New("dns: unimplemented")

// UnexpectedEOFError reports a read that ended before the field was fully
// consumed. It wraps the underlying I/O error (typically io.ErrUnexpectedEOF)
// so callers can use errors.Is.
type UnexpectedEOFError struct {
	// Field is the dotted struct path of the field being read when the
	// reader ran out of bytes, e.g. "Header.Length".
	Field string
	// Want is the number of bytes expected.
	Want int
	// Got is the number of bytes actually read.
	Got int
	// Err is the underlying reader error.
	Err error
}

// Error implements the error interface.
func (e *UnexpectedEOFError) Error() string {
	return fmt.Sprintf(
		"dns: unexpected EOF reading %s: wanted %d bytes, got %d: %v",
		e.Field, e.Want, e.Got, e.Err,
	)
}

// Unwrap returns the underlying I/O error.
func (e *UnexpectedEOFError) Unwrap() error {
	return e.Err
}

// decoder pulls DNS messages from an io.Reader. It is internal: callers go
// through the package-level Decode function.
type decoder struct {
	r         io.Reader
	byteOrder binary.ByteOrder
}

// newDecoder constructs a decoder over r using big-endian byte order, which
// matches the DNS wire format (network byte order).
func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         r,
		byteOrder: binary.BigEndian,
	}
}

// readFile reads a complete DNS message from the underlying reader.
//
// This is currently a stub. Replace the body with real per-section reads
// (Header, Questions, Answers, Authorities, Additionals) when the spec is
// implemented.
func (d *decoder) readFile() (*File, error) {
	return &File{}, fmt.Errorf("decoding File: %w", errUnimplemented)
}

// Decode reads a DNS message from r and returns the decoded *File.
func Decode(r io.Reader) (*File, error) {
	d := newDecoder(r)
	return d.readFile()
}
