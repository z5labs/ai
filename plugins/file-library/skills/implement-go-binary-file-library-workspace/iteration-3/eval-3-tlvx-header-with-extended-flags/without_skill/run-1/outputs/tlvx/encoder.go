package tlvx

import (
	"encoding/binary"
	"io"
)

// countingWriter wraps an io.Writer and tracks the number of bytes written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

// encoder writes TLVX structures to an underlying io.Writer.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{w: &countingWriter{w: w}, byteOrder: binary.BigEndian}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// Always called inside an `if err != nil` branch.
func (e *encoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}
}

// wrapAt returns a FieldError → OffsetError → err chain anchored at offset.
func (*encoder) wrapAt(field string, offset int64, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: offset, Err: err}}
}

// writeAll writes buf via the counting writer and returns the offset at which
// the field began (so callers can attribute errors to the start of the field).
func (e *encoder) writeAll(buf []byte) (int64, error) {
	start := e.w.n
	_, err := e.w.Write(buf)
	return start, err
}

func (e *encoder) writeHeader(h Header) error {
	// Magic: prefer the literal "TLVX" so encode-side enforces the format
	// regardless of whether the caller pre-populated h.Magic. The Header
	// passed in still has its zero-value Magic accepted; we substitute the
	// canonical value.
	magic := h.Magic
	if magic == ([4]byte{}) {
		magic = Magic
	}
	if off, err := e.writeAll(magic[:]); err != nil {
		return e.wrapAt("Header.Magic", off, err)
	}

	// Version.
	version := h.Version
	if version == 0 {
		version = Version
	}
	if off, err := e.writeAll([]byte{version}); err != nil {
		return e.wrapAt("Header.Version", off, err)
	}

	// Flags.
	if off, err := e.writeAll([]byte{byte(h.Flags)}); err != nil {
		return e.wrapAt("Header.Flags", off, err)
	}

	// ChecksumAlg.
	if off, err := e.writeAll([]byte{byte(h.ChecksumAlg)}); err != nil {
		return e.wrapAt("Header.ChecksumAlg", off, err)
	}

	// Reserved1: always written as 0 regardless of caller input, per spec.
	if off, err := e.writeAll([]byte{0}); err != nil {
		return e.wrapAt("Header.Reserved1", off, err)
	}

	// IndexCount.
	var b2 [2]byte
	e.byteOrder.PutUint16(b2[:], h.IndexCount)
	if off, err := e.writeAll(b2[:]); err != nil {
		return e.wrapAt("Header.IndexCount", off, err)
	}

	// ExtCount.
	e.byteOrder.PutUint16(b2[:], h.ExtCount)
	if off, err := e.writeAll(b2[:]); err != nil {
		return e.wrapAt("Header.ExtCount", off, err)
	}

	// TrailerOffset.
	var b4 [4]byte
	e.byteOrder.PutUint32(b4[:], h.TrailerOffset)
	if off, err := e.writeAll(b4[:]); err != nil {
		return e.wrapAt("Header.TrailerOffset", off, err)
	}

	return nil
}

func (e *encoder) writeFile(f *File) error {
	return e.writeHeader(f.Header)
}

// Encode writes f to w as a TLVX file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
