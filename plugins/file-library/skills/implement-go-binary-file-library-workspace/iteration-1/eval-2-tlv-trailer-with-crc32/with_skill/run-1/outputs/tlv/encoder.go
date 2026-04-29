package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingWriter wraps an io.Writer and tracks the number of bytes written.
// It also tees every successful write into a running CRC32 hash so the trailer
// can be emitted as the IEEE CRC32 of every byte written so far.
type countingWriter struct {
	w    io.Writer
	n    int64
	hash hash.Hash32
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	if c.hash != nil && n > 0 {
		// hash.Hash.Write never returns an error per its contract.
		_, _ = c.hash.Write(p[:n])
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

// writeFile emits the payload bytes followed by a 4-byte big-endian CRC32 (IEEE)
// trailer covering exactly those payload bytes.
func (e *encoder) writeFile(f *File) error {
	if f == nil {
		return e.wrapErr("File", ErrInvalid)
	}

	if _, err := e.w.Write(f.Payload); err != nil {
		return e.wrapErr("File.Payload", err)
	}

	// Snapshot the running CRC32 BEFORE the trailer bytes are written — the
	// trailer must cover only the bytes preceding it.
	return e.writeTrailer()
}

// writeTrailer writes the 4-byte big-endian IEEE CRC32 of every byte written so
// far. It must be called after every other byte of the file has been emitted.
func (e *encoder) writeTrailer() error {
	sum := e.w.hash.Sum32()

	var buf [4]byte
	e.byteOrder.PutUint32(buf[:], sum)

	if _, err := e.w.Write(buf[:]); err != nil {
		return e.wrapErr("Trailer.CRC32", err)
	}
	return nil
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
