package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingWriter wraps an io.Writer, tallies bytes written, and (optionally)
// feeds every written byte into a running CRC32 hash so the encoder can emit
// a trailer covering everything written before it.
type countingWriter struct {
	w   io.Writer
	n   int64
	crc hash.Hash32
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if n > 0 && c.crc != nil {
		// Hash.Write never returns an error.
		c.crc.Write(p[:n])
	}
	c.n += int64(n)
	return n, err
}

// encoder writes TLV1 structures to an underlying io.Writer.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w:         &countingWriter{w: w, crc: crc32.NewIEEE()},
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

// writeTrailer snapshots the running CRC32 and writes it as a 4-byte
// big-endian integer. The snapshot must happen BEFORE the bytes are written
// so the trailer itself is excluded from the checksum.
func (e *encoder) writeTrailer() error {
	crc := e.w.crc.Sum32()
	var buf [4]byte
	e.byteOrder.PutUint32(buf[:], crc)
	if _, err := e.w.Write(buf[:]); err != nil {
		return e.wrapErr("Trailer.CRC32", err)
	}
	return nil
}

func (e *encoder) writeFile(_ *File) error {
	return e.writeTrailer()
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
