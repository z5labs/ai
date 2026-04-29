// Copyright (c) 2026 Z5labs and Contributors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"encoding/binary"
	"fmt"
	"io"
)

// encoder owns an io.Writer and the byte order used to serialize
// every multi-byte field in the gzip wire format. The byte order
// must match the decoder's.
type encoder struct {
	w         io.Writer
	byteOrder binary.ByteOrder
}

// newEncoder constructs an encoder defaulting to binary.BigEndian.
// Note: the real gzip format is little-endian on the wire; the
// implementer should switch this to binary.LittleEndian when filling
// in writeFile against the spec, in lockstep with the decoder.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         w,
		byteOrder: binary.BigEndian,
	}
}

// writeFile is the top-level encode entry point. It is a stub today
// and returns an "unimplemented" error wrapped with the structural
// context the rest of the package uses.
func (e *encoder) writeFile(f *File) error {
	return fmt.Errorf("encoding File: %w", errUnimplemented)
}

// Encode writes f to w in the gzip wire format.
func Encode(w io.Writer, f *File) error {
	e := newEncoder(w)
	return e.writeFile(f)
}
