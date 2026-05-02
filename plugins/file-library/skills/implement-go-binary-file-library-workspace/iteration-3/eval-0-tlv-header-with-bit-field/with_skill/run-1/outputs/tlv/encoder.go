package tlv

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

// encoder writes TLV1 structures to an underlying io.Writer.
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

func (e *encoder) writeFile(f *File) error {
	return e.writeHeader(&f.Header)
}

// writeHeader writes the fixed 8-byte TLV1 header. Validation runs before any
// bytes are written so a malformed header doesn't leave a partial file behind.
func (e *encoder) writeHeader(h *Header) error {
	if h.Magic != Magic {
		return e.wrapErr("Header.Magic", ErrInvalidMagic)
	}
	if h.Version != Version1 {
		return e.wrapErr("Header.Version", ErrUnsupportedVersion)
	}
	if h.Flags&flagsReservedMask != 0 {
		return e.wrapErr("Header.Flags", ErrReservedFlagsSet)
	}
	if h.Flags&FlagSigned != 0 {
		return e.wrapErr("Header.Flags", ErrSignedFlagUnsupported)
	}
	if h.Reserved != 0 {
		return e.wrapErr("Header.Reserved", ErrReservedNonZero)
	}

	if err := binary.Write(e.w, e.byteOrder, h); err != nil {
		return e.wrapErr("Header", err)
	}
	return nil
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
