package tlv

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

// decoder reads TLV1 structures from an underlying io.Reader.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{r: &countingReader{r: r}, byteOrder: binary.BigEndian}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// All decode-time errors must go through this helper so the offset stays accurate.
func (d *decoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readFull reads exactly len(buf) bytes or returns an error.
func (d *decoder) readFull(buf []byte) error {
	_, err := io.ReadFull(d.r, buf)
	return err
}

func (d *decoder) readHeader() (Header, error) {
	var h Header
	var buf [8]byte
	if err := d.readFull(buf[:]); err != nil {
		return h, d.wrapErr("Header", err)
	}

	copy(h.Magic[:], buf[0:4])
	if h.Magic != Magic {
		return h, d.wrapErr("Header.Magic", ErrBadMagic)
	}

	h.Version = buf[4]
	if h.Version != Version {
		return h, d.wrapErr("Header.Version", ErrUnsupportedVersion)
	}

	h.Flags = Flags(buf[5])
	if h.Flags&flagsReservedMask != 0 {
		return h, d.wrapErr("Header.Flags", ErrReservedFlagSet)
	}
	if h.Flags.Signed() {
		return h, d.wrapErr("Header.Flags", ErrSignedFlagUnsupported)
	}

	h.Reserved = d.byteOrder.Uint16(buf[6:8])
	if h.Reserved != 0 {
		return h, d.wrapErr("Header.Reserved", ErrReservedNotZero)
	}

	return h, nil
}

func (d *decoder) readFile() (*File, error) {
	h, err := d.readHeader()
	if err != nil {
		return nil, err
	}
	return &File{Header: h}, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
