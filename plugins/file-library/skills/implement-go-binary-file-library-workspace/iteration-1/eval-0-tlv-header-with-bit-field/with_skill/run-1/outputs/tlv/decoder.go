package tlv

import (
	"bytes"
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

// readHeader reads the fixed-size 8-byte TLV1 header and validates the
// invariant fields (Magic, Version, reserved bits).
func (d *decoder) readHeader() (Header, error) {
	var h Header
	if err := binary.Read(d.r, d.byteOrder, &h); err != nil {
		// binary.Read surfaces a short read as io.ErrUnexpectedEOF.
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return Header{}, d.wrapErr("Header", err)
	}

	if !bytes.Equal(h.Magic[:], Magic[:]) {
		return Header{}, d.wrapErr("Header.Magic", &UnexpectedValueError{
			Field: "Header.Magic",
			Got:   h.Magic,
		})
	}

	if h.Version != Version1 {
		return Header{}, d.wrapErr("Header.Version", ErrUnsupportedVersion)
	}

	if h.Flags&FlagsReservedMask != 0 {
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
