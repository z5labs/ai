package tlvx

import (
	"encoding/binary"
	"io"
)

// countingReader wraps an io.Reader and tracks the number of bytes consumed.
type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// decoder reads TLVX structures from an underlying io.Reader.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{r: &countingReader{r: r}, byteOrder: binary.BigEndian}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// Always called inside an `if err != nil` branch, so it doesn't guard against
// nil errors itself. The OffsetError.Offset is captured at the moment of the
// call, which is the byte position immediately after the failing read.
func (d *decoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readFull reads exactly len(buf) bytes from the underlying reader. It snaps
// the field-start offset before the read so callers can wrap the returned
// error with the offset of the first byte of the failing field.
func (d *decoder) readFull(buf []byte) (int64, error) {
	start := d.r.n
	_, err := io.ReadFull(d.r, buf)
	return start, err
}

// wrapAt returns a FieldError → OffsetError → err chain anchored at offset.
func (*decoder) wrapAt(field string, offset int64, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: offset, Err: err}}
}

func (d *decoder) readHeader() (Header, error) {
	var h Header

	// Magic: 4 bytes.
	var magic [4]byte
	off, err := d.readFull(magic[:])
	if err != nil {
		return h, d.wrapAt("Header.Magic", off, err)
	}
	if magic != Magic {
		return h, d.wrapAt("Header.Magic", off, ErrMagicMismatch)
	}
	h.Magic = magic

	// Version: 1 byte.
	var b1 [1]byte
	off, err = d.readFull(b1[:])
	if err != nil {
		return h, d.wrapAt("Header.Version", off, err)
	}
	if b1[0] != Version {
		return h, d.wrapAt("Header.Version", off, ErrUnsupportedVersion)
	}
	h.Version = b1[0]

	// Flags: 1 byte.
	off, err = d.readFull(b1[:])
	if err != nil {
		return h, d.wrapAt("Header.Flags", off, err)
	}
	h.Flags = Flags(b1[0])

	// ChecksumAlg: 1 byte. Accept any value here; trailer-time validation
	// (out of scope for this iteration) is responsible for surfacing
	// UnknownChecksumAlgError.
	off, err = d.readFull(b1[:])
	if err != nil {
		return h, d.wrapAt("Header.ChecksumAlg", off, err)
	}
	h.ChecksumAlg = ChecksumAlg(b1[0])

	// Reserved1: 1 byte. Must be zero.
	off, err = d.readFull(b1[:])
	if err != nil {
		return h, d.wrapAt("Header.Reserved1", off, err)
	}
	if b1[0] != 0 {
		return h, d.wrapAt("Header.Reserved1", off, ErrReservedNonZero)
	}
	h.Reserved1 = 0

	// IndexCount: 2 bytes.
	var b2 [2]byte
	off, err = d.readFull(b2[:])
	if err != nil {
		return h, d.wrapAt("Header.IndexCount", off, err)
	}
	h.IndexCount = d.byteOrder.Uint16(b2[:])

	// ExtCount: 2 bytes.
	off, err = d.readFull(b2[:])
	if err != nil {
		return h, d.wrapAt("Header.ExtCount", off, err)
	}
	h.ExtCount = d.byteOrder.Uint16(b2[:])

	// TrailerOffset: 4 bytes.
	var b4 [4]byte
	off, err = d.readFull(b4[:])
	if err != nil {
		return h, d.wrapAt("Header.TrailerOffset", off, err)
	}
	h.TrailerOffset = d.byteOrder.Uint32(b4[:])

	return h, nil
}

func (d *decoder) readFile() (*File, error) {
	h, err := d.readHeader()
	if err != nil {
		return nil, err
	}
	return &File{Header: h}, nil
}

// Decode reads a TLVX file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
