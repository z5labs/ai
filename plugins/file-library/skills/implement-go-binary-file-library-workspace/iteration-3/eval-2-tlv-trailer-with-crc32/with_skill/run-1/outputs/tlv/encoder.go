package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingWriter wraps an io.Writer and tracks the number of bytes written.
// It also folds every byte it sees into an attached hash, used by the
// encoder to compute the running CRC32 covering all bytes that precede the
// trailer.
type countingWriter struct {
	w io.Writer
	n int64
	// h, when non-nil, receives every byte written. The encoder uses
	// this to fold the body into the CRC32 hash, then nils it out
	// before writing the trailer so the trailer's own bytes are not
	// folded into the hash.
	h hash.Hash32
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	if c.h != nil && n > 0 {
		// hash.Hash.Write never returns an error per the contract.
		_, _ = c.h.Write(p[:n])
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
		w:         &countingWriter{w: w, h: crc32.NewIEEE()},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// Always called inside an `if err != nil` branch.
func (e *encoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}
}

// writeFile writes a TLV1 file. The current implementation only realizes the
// Trailer pipeline: every body byte (none, until Header/Record writers land)
// is folded into a running CRC32 via the countingWriter's hash, then the
// trailer is written. Once Header/Record writers exist, call them before
// writeTrailer — they'll write through the same countingWriter and the CRC
// will pick them up automatically.
func (e *encoder) writeFile(f *File) error {
	if err := e.writeTrailer(); err != nil {
		return err
	}
	return nil
}

// writeTrailer writes the 4-byte big-endian CRC32 trailer. The CRC is taken
// from the running hash (which has seen every byte written so far via the
// countingWriter), then the hash is detached so the trailer's own bytes are
// not folded back into it.
func (e *encoder) writeTrailer() error {
	crc := e.w.h.Sum32()
	// Detach the hash so the trailer's own bytes don't enter it. (The
	// hash isn't read again after this point, but detaching keeps the
	// invariant honest if anything else is later written.)
	e.w.h = nil

	var buf [4]byte
	e.byteOrder.PutUint32(buf[:], crc)
	if _, err := e.w.Write(buf[:]); err != nil {
		return e.wrapErr("Trailer.CRC32", err)
	}
	return nil
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
