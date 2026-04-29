package tlv

import (
	"encoding/binary"
	"errors"
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

// magicBytes is the required value of Header.Magic — ASCII "TLV1".
var magicBytes = [4]byte{'T', 'L', 'V', '1'}

// supportedVersion is the only TLV version this package currently understands.
const supportedVersion uint8 = 1

func (d *decoder) readHeader() (Header, error) {
	var h Header
	// Read all 8 bytes in one shot. binary.Read converts io.EOF into
	// io.ErrUnexpectedEOF for partial reads, which is exactly the leaf
	// sentinel our truncation test asserts.
	if err := binary.Read(d.r, d.byteOrder, &h); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return Header{}, d.wrapErr("Header", err)
	}
	if h.Magic != magicBytes {
		return Header{}, d.wrapErr("Header.Magic", ErrInvalidMagic)
	}
	if h.Version != supportedVersion {
		return Header{}, d.wrapErr("Header.Version", ErrUnsupportedVersion)
	}
	if h.Flags&flagsReservedMask != 0 {
		return Header{}, d.wrapErr("Header.Flags", ErrReservedBitsSet)
	}
	if h.Reserved != 0 {
		return Header{}, d.wrapErr("Header.Reserved", ErrReservedBitsSet)
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
