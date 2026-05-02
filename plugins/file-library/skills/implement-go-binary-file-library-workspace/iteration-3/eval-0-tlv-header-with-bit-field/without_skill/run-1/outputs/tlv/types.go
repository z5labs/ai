package tlv

import (
	"errors"
	"fmt"
)

// Magic is the 4-byte ASCII identifier "TLV1" that starts every TLV1 file.
var Magic = [4]byte{'T', 'L', 'V', '1'}

// Version is the only currently-defined format version.
const Version uint8 = 1

// Flags is a bit field carried in the header. See the SPEC.md
// "Header.Flags bit field" section.
type Flags uint8

const (
	// FlagCompressed (bit 0, mask 0x01) indicates record values are zlib-compressed.
	FlagCompressed Flags = 1 << 0
	// FlagEncrypted (bit 1, mask 0x02) indicates record values are AES-encrypted.
	FlagEncrypted Flags = 1 << 1
	// FlagSigned (bit 2, mask 0x04) indicates the trailer carries a signature
	// in addition to the CRC.
	FlagSigned Flags = 1 << 2

	// flagsReservedMask covers the bits that must be zero (bits 3-7).
	flagsReservedMask Flags = 0xF8
)

// Has reports whether all bits in mask are set in f.
func (f Flags) Has(mask Flags) bool { return f&mask == mask }

// Header is the fixed 8-byte header of a TLV1 file.
type Header struct {
	Magic    [4]byte
	Version  uint8
	Flags    Flags
	Reserved uint16
}

// File is the top-level decoded representation of a TLV1 file.
// Implementer fills this in as Header, Records, and Trailer are added.
type File struct {
	Header Header
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrBadMagic is returned when the header's Magic field is not "TLV1".
var ErrBadMagic = errors.New("tlv: bad magic")

// ErrUnsupportedVersion is returned when the header's Version field is not 1.
var ErrUnsupportedVersion = errors.New("tlv: unsupported version")

// ErrReservedNonZero is returned when the header's Reserved field is non-zero.
var ErrReservedNonZero = errors.New("tlv: reserved field non-zero")

// ErrReservedFlagBitsSet is returned when reserved bits 3-7 are set in Flags.
var ErrReservedFlagBitsSet = errors.New("tlv: reserved flag bits set")

// errUnimplemented is returned by stub methods so tests can assert the error
// chain via errors.Is even before the real implementation lands.
var errUnimplemented = errors.New("tlv: unimplemented")

// OffsetError records the byte offset at which a decode or encode error occurred.
type OffsetError struct {
	Offset int64
	Err    error
}

func (e *OffsetError) Error() string { return fmt.Sprintf("at byte %d: %v", e.Offset, e.Err) }
func (e *OffsetError) Unwrap() error { return e.Err }

// FieldError records the field path (e.g. "Header.Length") at which an error occurred.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string { return fmt.Sprintf("decoding %s: %v", e.Field, e.Err) }
func (e *FieldError) Unwrap() error { return e.Err }
