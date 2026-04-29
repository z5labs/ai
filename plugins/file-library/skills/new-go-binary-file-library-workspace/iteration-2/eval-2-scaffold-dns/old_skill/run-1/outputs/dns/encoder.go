// Copyright (c) 2026 z5labs
//
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file or at https://opensource.org/licenses/MIT.

package dns

import (
	"encoding/binary"
	"fmt"
	"io"
)

// encoder pushes DNS messages to an io.Writer. It is internal: callers go
// through the package-level Encode function.
type encoder struct {
	w         io.Writer
	byteOrder binary.ByteOrder
}

// newEncoder constructs an encoder over w using big-endian byte order, which
// matches the DNS wire format (network byte order).
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         w,
		byteOrder: binary.BigEndian,
	}
}

// writeFile writes a complete DNS message to the underlying writer.
//
// This is currently a stub. Replace the body with real per-section writes
// (Header, Questions, Answers, Authorities, Additionals) when the spec is
// implemented.
func (e *encoder) writeFile(f *File) error {
	return fmt.Errorf("encoding File: %w", errUnimplemented)
}

// Encode writes the DNS message f to w.
func Encode(w io.Writer, f *File) error {
	e := newEncoder(w)
	return e.writeFile(f)
}
