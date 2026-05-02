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

// writeAll writes the entire buf, returning any underlying writer error.
func (e *encoder) writeAll(buf []byte) error {
	_, err := e.w.Write(buf)
	return err
}

// writeHeader emits the fixed 8-byte header. Validation mirrors the decoder so
// a struct that wouldn't decode also won't encode.
func (e *encoder) writeHeader(h Header) error {
	// Magic (offset 0, 4 bytes)
	magic := h.Magic
	// If the caller left Magic zero, fill in the canonical "TLV1".
	if magic == ([4]byte{}) {
		magic = Magic
	}
	if magic != Magic {
		return e.wrapErr("Header.Magic", ErrBadMagic)
	}
	if err := e.writeAll(magic[:]); err != nil {
		return e.wrapErr("Header.Magic", err)
	}

	// Version (offset 4, 1 byte) - default 0 → 1.
	version := h.Version
	if version == 0 {
		version = Version
	}
	if version != Version {
		return e.wrapErr("Header.Version", ErrUnsupportedVersion)
	}
	if err := e.writeAll([]byte{version}); err != nil {
		return e.wrapErr("Header.Version", err)
	}

	// Flags (offset 5, 1 byte)
	if h.Flags&flagsReservedMask != 0 {
		return e.wrapErr("Header.Flags", ErrReservedFlagBitsSet)
	}
	if err := e.writeAll([]byte{byte(h.Flags)}); err != nil {
		return e.wrapErr("Header.Flags", err)
	}

	// Reserved (offset 6, 2 bytes) - must be 0.
	if h.Reserved != 0 {
		return e.wrapErr("Header.Reserved", ErrReservedNonZero)
	}
	var reservedBuf [2]byte
	e.byteOrder.PutUint16(reservedBuf[:], h.Reserved)
	if err := e.writeAll(reservedBuf[:]); err != nil {
		return e.wrapErr("Header.Reserved", err)
	}

	return nil
}

func (e *encoder) writeFile(f *File) error {
	if f == nil {
		return e.wrapErr("File", ErrInvalid)
	}
	return e.writeHeader(f.Header)
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
