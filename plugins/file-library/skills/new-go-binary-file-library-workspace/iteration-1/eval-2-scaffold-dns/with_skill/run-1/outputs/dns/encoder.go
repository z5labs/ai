// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"encoding/binary"
	"fmt"
	"io"
)

// encoder is the inverse of decoder: it owns the io.Writer and byte order
// and serializes Go values back into the DNS wire format.
type encoder struct {
	w         io.Writer
	byteOrder binary.ByteOrder
}

// newEncoder constructs an encoder defaulting to the network byte order
// (big-endian). The byte order must match the decoder for round-trip
// correctness.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         w,
		byteOrder: binary.BigEndian,
	}
}

// writeFile is the entry point for encoding a DNS message. The scaffold
// returns an unimplemented error so the test suite has a stable failure mode
// to invert once the real encoder lands.
func (e *encoder) writeFile(f *File) error {
	return fmt.Errorf("encoding File: %w", errUnimplemented)
}

// Encode serializes f to w in the DNS wire format. It is the public entry
// point of the encoder.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
