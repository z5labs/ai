package tlv

import (
	"errors"
	"fmt"
)

// Magic is the four-byte ASCII identifier "TLV1" carried at the start of every
// TLV1 file header.
var Magic = [4]byte{'T', 'L', 'V', '1'}

// Version1 is the only TLV1 format version currently defined.
const Version1 uint8 = 1

// File is the top-level decoded representation of a TLV1 file.
// Records and Trailer are added in subsequent phases.
type File struct {
	Header Header
}

// Header is the fixed-size 8-byte header that begins every TLV1 file.
//
// Field order matches wire order (big-endian) so the struct lines up against
// the byte layout described in SPEC.md's Field Definitions section.
type Header struct {
	Magic    [4]byte
	Version  uint8
	Flags    Flags
	Reserved uint16
}

// Flags is the bit field carried in Header.Flags. The underlying byte holds
// three independent boolean flags; the upper five bits are reserved and must
// be zero on encode.
type Flags uint8

// Header.Flags bit field constants. Each constant is the mask for its bit.
const (
	FlagCompressed Flags = 0x01 // Record values are zlib-compressed.
	FlagEncrypted  Flags = 0x02 // Record values are AES-encrypted.
	FlagSigned     Flags = 0x04 // Trailer carries a signature in addition to the CRC.

	// FlagsReservedMask covers bits 3-7. Must be zero.
	FlagsReservedMask Flags = 0xF8
)

// Has reports whether all of the bits in mask are set in f.
func (f Flags) Has(mask Flags) bool { return f&mask == mask }

// String returns a human-readable rendering of the set flags. Useful when a
// hex-dump test fails.
func (f Flags) String() string {
	if f == 0 {
		return "0"
	}
	out := ""
	if f.Has(FlagCompressed) {
		out += "COMPRESSED|"
	}
	if f.Has(FlagEncrypted) {
		out += "ENCRYPTED|"
	}
	if f.Has(FlagSigned) {
		out += "SIGNED|"
	}
	if reserved := f & FlagsReservedMask; reserved != 0 {
		out += fmt.Sprintf("RESERVED(0x%02x)|", uint8(reserved))
	}
	if out == "" {
		return fmt.Sprintf("0x%02x", uint8(f))
	}
	return out[:len(out)-1]
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrUnsupportedVersion is returned when Header.Version is not a recognized
// TLV1 version. Future revisions will accept additional values.
var ErrUnsupportedVersion = errors.New("tlv: unsupported version")

// ErrReservedBitsSet is returned when reserved bits in a flag/reserved field
// are non-zero. Decoders treat this as a hard error so future-version files
// are not silently accepted.
var ErrReservedBitsSet = errors.New("tlv: reserved bits set")

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

// UnexpectedValueError reports that a fixed-value field carried something other
// than the spec-required value (e.g. Header.Magic != "TLV1").
type UnexpectedValueError struct {
	Field string
	Got   any
}

func (e *UnexpectedValueError) Error() string {
	return fmt.Sprintf("unexpected value for %s: %v", e.Field, e.Got)
}
