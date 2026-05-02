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
// Always called inside an `if err != nil` branch, so it doesn't guard against
// nil errors itself.
func (d *decoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readFull reads exactly len(buf) bytes from the underlying reader.
// On a short read it surfaces the underlying error (or io.ErrUnexpectedEOF).
func (d *decoder) readFull(buf []byte) error {
	_, err := io.ReadFull(d.r, buf)
	return err
}

// readHeader decodes the fixed 8-byte header. The offset reported in any
// returned error refers to the first byte of the field that failed to decode.
func (d *decoder) readHeader() (Header, error) {
	var h Header

	// Magic (offset 0, 4 bytes)
	if err := d.readFull(h.Magic[:]); err != nil {
		return Header{}, d.wrapErr("Header.Magic", err)
	}
	if h.Magic != Magic {
		return Header{}, d.wrapErr("Header.Magic", ErrBadMagic)
	}

	// Version (offset 4, 1 byte)
	var versionBuf [1]byte
	if err := d.readFull(versionBuf[:]); err != nil {
		return Header{}, d.wrapErr("Header.Version", err)
	}
	h.Version = versionBuf[0]
	if h.Version != Version {
		return Header{}, d.wrapErr("Header.Version", ErrUnsupportedVersion)
	}

	// Flags (offset 5, 1 byte)
	var flagsBuf [1]byte
	if err := d.readFull(flagsBuf[:]); err != nil {
		return Header{}, d.wrapErr("Header.Flags", err)
	}
	h.Flags = Flags(flagsBuf[0])
	if h.Flags&flagsReservedMask != 0 {
		return Header{}, d.wrapErr("Header.Flags", ErrReservedFlagBitsSet)
	}

	// Reserved (offset 6, 2 bytes)
	var reservedBuf [2]byte
	if err := d.readFull(reservedBuf[:]); err != nil {
		return Header{}, d.wrapErr("Header.Reserved", err)
	}
	h.Reserved = d.byteOrder.Uint16(reservedBuf[:])
	if h.Reserved != 0 {
		return Header{}, d.wrapErr("Header.Reserved", ErrReservedNonZero)
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
