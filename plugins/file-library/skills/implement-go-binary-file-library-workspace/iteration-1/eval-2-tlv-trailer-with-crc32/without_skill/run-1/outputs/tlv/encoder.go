package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingWriter wraps an io.Writer and tracks the number of bytes written.
// It also feeds every byte written into a running CRC32 so the trailer can be
// emitted as the IEEE CRC32 of all preceding bytes.
type countingWriter struct {
	w    io.Writer
	n    int64
	hash hash.Hash32
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if n > 0 {
		_, _ = c.hash.Write(p[:n])
		c.n += int64(n)
	}
	return n, err
}

// encoder writes TLV1 structures to an underlying io.Writer.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         &countingWriter{w: w, hash: crc32.NewIEEE()},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
func (e *encoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}
}

// writeFile writes a TLV1 file to w.
//
// Currently only the trailer is implemented: the body is empty (header/records
// will be filled in by future work) and the trailer is the IEEE CRC32 of every
// byte already written, encoded big-endian.
func (e *encoder) writeFile(f *File) error {
	// Future: header + records would be written here, each going through
	// e.w.Write so the running CRC stays consistent with the bytes on the wire.

	// Trailer: snapshot CRC of everything written so far, then emit.
	sum := e.w.hash.Sum32()

	var buf [4]byte
	e.byteOrder.PutUint32(buf[:], sum)

	// Use the underlying writer directly so we don't fold the trailer back
	// into the running CRC.
	if _, err := e.w.w.Write(buf[:]); err != nil {
		return e.wrapErr("Trailer.CRC32", err)
	}
	e.w.n += 4
	return nil
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
