package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingReader wraps an io.Reader and tracks the number of bytes consumed.
// It also feeds every byte it reads into an optional running CRC32 hash so
// the decoder can verify the trailer without rebuffering the file.
type countingReader struct {
	r   io.Reader
	n   int64
	crc hash.Hash32
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 && c.crc != nil {
		// Hash.Write never returns an error.
		c.crc.Write(p[:n])
	}
	c.n += int64(n)
	return n, err
}

// decoder reads TLV1 structures from an underlying io.Reader.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         &countingReader{r: r, crc: crc32.NewIEEE()},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// All decode-time errors must go through this helper so the offset stays accurate.
func (d *decoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readTrailer reads the 4-byte big-endian CRC32 trailer and verifies it
// against the running CRC32 of every byte already consumed from the underlying
// reader. The expected CRC must be snapshotted before reading the trailer
// bytes themselves, otherwise the trailer bytes would be folded into the hash.
func (d *decoder) readTrailer() (Trailer, error) {
	expected := d.r.crc.Sum32()

	var buf [4]byte
	if _, err := io.ReadFull(d.r, buf[:]); err != nil {
		return Trailer{}, d.wrapErr("Trailer.CRC32", err)
	}
	got := d.byteOrder.Uint32(buf[:])
	if got != expected {
		return Trailer{}, d.wrapErr("Trailer.CRC32", ErrChecksumMismatch)
	}
	return Trailer{CRC32: got}, nil
}

func (d *decoder) readFile() (*File, error) {
	tr, err := d.readTrailer()
	if err != nil {
		return nil, err
	}
	return &File{Trailer: tr}, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
