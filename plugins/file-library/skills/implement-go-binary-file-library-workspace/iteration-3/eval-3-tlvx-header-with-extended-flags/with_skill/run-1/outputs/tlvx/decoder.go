package tlvx

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
	// binary.Read uses io.ReadFull internally; surface partial reads as
	// io.ErrUnexpectedEOF the same way the standard library does.
	if errors.Is(err, io.EOF) && n > 0 {
		err = io.ErrUnexpectedEOF
	}
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
// nil errors itself.
func (d *decoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readFile reads a complete TLVX file. Currently only the Header is wired up.
func (d *decoder) readFile() (*File, error) {
	hdr, err := d.readHeader()
	if err != nil {
		return nil, err
	}
	return &File{Header: *hdr}, nil
}

// readHeader reads the fixed 16-byte TLVX file header. The magic, version,
// reserved field, reserved flag bit, and checksum algorithm are all validated
// in-place; failures funnel through wrapErr so callers see the standard
// FieldError → OffsetError → leaf chain.
//
// The magic is read in its own binary.Read call so a magic-mismatch failure
// reports the offset of the magic field (4) rather than the offset after the
// entire 16-byte header has been slurped (16).
func (d *decoder) readHeader() (*Header, error) {
	var hdr Header

	// Magic first, on its own — wrong magic means we are looking at a
	// non-TLVX file, so further validation would be misleading. Reporting
	// the failure at offset 4 (right after the magic was consumed) lines up
	// with the field's position in the wire layout.
	if err := binary.Read(d.r, d.byteOrder, &hdr.Magic); err != nil {
		return nil, d.wrapErr("Header.Magic", err)
	}
	if hdr.Magic != MagicTLVX {
		return nil, d.wrapErr("Header.Magic", &UnexpectedMagicError{Got: hdr.Magic})
	}

	// The remaining 12 bytes form a fixed-size sub-record we can read in one
	// call; per-field validation runs after it lands.
	rest := struct {
		Version       uint8
		Flags         Flags
		ChecksumAlg   ChecksumAlg
		Reserved1     uint8
		IndexCount    uint16
		ExtCount      uint16
		TrailerOffset uint32
	}{}
	if err := binary.Read(d.r, d.byteOrder, &rest); err != nil {
		return nil, d.wrapErr("Header", err)
	}
	hdr.Version = rest.Version
	hdr.Flags = rest.Flags
	hdr.ChecksumAlg = rest.ChecksumAlg
	hdr.Reserved1 = rest.Reserved1
	hdr.IndexCount = rest.IndexCount
	hdr.ExtCount = rest.ExtCount
	hdr.TrailerOffset = rest.TrailerOffset

	if hdr.Version != Version1 {
		return nil, d.wrapErr("Header.Version", &UnknownVersionError{Version: hdr.Version})
	}

	// Reserved bit 7 of Flags must be zero per the spec.
	if hdr.Flags&flagsReservedMask != 0 {
		return nil, d.wrapErr("Header.Flags", &ReservedFlagBitError{Field: "Header.Flags", Bit: 7})
	}

	if hdr.Reserved1 != 0 {
		return nil, d.wrapErr("Header.Reserved1", &ReservedFieldNonZeroError{Field: "Header.Reserved1", Got: hdr.Reserved1})
	}

	// Validate the checksum algorithm tag against the defined enum range.
	switch hdr.ChecksumAlg {
	case ChecksumAlgCRC32IEEE,
		ChecksumAlgCRC64ECMA,
		ChecksumAlgSHA256T32,
		ChecksumAlgXXH64,
		ChecksumAlgBLAKE3T32:
		// ok
	default:
		return nil, d.wrapErr("Header.ChecksumAlg", &UnknownChecksumAlgError{Alg: hdr.ChecksumAlg})
	}

	return &hdr, nil
}

// Decode reads a TLVX file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
