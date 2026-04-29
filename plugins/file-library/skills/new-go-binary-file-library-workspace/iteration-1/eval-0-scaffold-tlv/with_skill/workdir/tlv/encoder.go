// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

import (
	"encoding/binary"
	"fmt"
	"io"
)

// encoder is the internal writer. It owns the io.Writer and the byte order
// used for every binary.Write call. The byte order MUST match the decoder's;
// they are read and write inverses of the same wire format.
type encoder struct {
	w         io.Writer
	byteOrder binary.ByteOrder
}

// newEncoder constructs an encoder writing to w. Default byte order is
// binary.BigEndian — keep this in sync with newDecoder.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         w,
		byteOrder: binary.BigEndian,
	}
}

// writeFile is the per-structure writer for the top-level File type. Replace
// the unimplemented stub with real encoding logic — the inverse of readFile,
// using binary.Write for fixed-size fields and explicit length prefixes plus
// e.w.Write for variable-length payloads.
func (e *encoder) writeFile(f *File) error {
	return fmt.Errorf("encoding File: %w", errUnimplemented)
}

// Encode writes a TLV File to w.
func Encode(w io.Writer, f *File) error {
	e := newEncoder(w)
	return e.writeFile(f)
}
