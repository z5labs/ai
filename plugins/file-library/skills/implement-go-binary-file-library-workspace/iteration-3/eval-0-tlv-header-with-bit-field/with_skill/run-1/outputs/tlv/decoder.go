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

func (d *decoder) readFile() (*File, error) {
	h, err := d.readHeader()
	if err != nil {
		return nil, err
	}
	return &File{Header: h}, nil
}

// readHeader reads the fixed 8-byte TLV1 header.
//
// Wire layout (big-endian):
//
//	Magic    [4]byte
//	Version  uint8
//	Flags    uint8
//	Reserved uint16
//
// On a short read the failure surfaces as
// FieldError{Field:"Header"} → OffsetError → io.ErrUnexpectedEOF.
// Per-field validation errors (invalid magic, unsupported version, reserved
// bits set, SIGNED flag, non-zero Reserved) surface with a more specific
// field path like "Header.Magic" / "Header.Version" / "Header.Flags" /
// "Header.Reserved" wrapping a typed sentinel.
func (d *decoder) readHeader() (Header, error) {
	var h Header
	if err := binary.Read(d.r, d.byteOrder, &h); err != nil {
		// Translate io.EOF — which binary.Read can return when no bytes are
		// available at all — to io.ErrUnexpectedEOF so a partial-or-empty
		// read on the header always surfaces the same leaf sentinel.
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return Header{}, d.wrapErr("Header", err)
	}

	if h.Magic != Magic {
		return Header{}, d.wrapErr("Header.Magic", ErrInvalidMagic)
	}
	if h.Version != Version1 {
		return Header{}, d.wrapErr("Header.Version", ErrUnsupportedVersion)
	}
	if h.Flags&flagsReservedMask != 0 {
		return Header{}, d.wrapErr("Header.Flags", ErrReservedFlagsSet)
	}
	if h.Flags&FlagSigned != 0 {
		return Header{}, d.wrapErr("Header.Flags", ErrSignedFlagUnsupported)
	}
	if h.Reserved != 0 {
		return Header{}, d.wrapErr("Header.Reserved", ErrReservedNonZero)
	}
	return h, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
