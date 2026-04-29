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

// encoder is the internal writer counterpart to decoder. It owns the
// io.Writer and the byte order, which must match the decoder's.
type encoder struct {
	w         io.Writer
	byteOrder binary.ByteOrder
}

// newEncoder constructs an encoder that defaults to binary.BigEndian, the
// scaffolding default. Switch to binary.LittleEndian when implementing the
// real gzip wire format.
func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         w,
		byteOrder: binary.BigEndian,
	}
}

// writeFile is the per-structure writer for the root File type. Real
// implementations should mirror readFile in decoder.go and emit the same
// fields in the same wire order.
func (e *encoder) writeFile(f *File) error {
	err := errors.New("unimplemented")
	return fmt.Errorf("encoding File: %w", err)
}

// Encode writes f to w as a gzip file. It is the minimal public surface;
// format options should go on a separate EncodeWithOptions function rather
// than overloading this one.
func Encode(w io.Writer, f *File) error {
	e := newEncoder(w)
	return e.writeFile(f)
}
