package tlv

import (
	"errors"
	"fmt"
)

// File is the top-level decoded representation of a TLV1 file.
// Implementer fills this in as Header, Records, and Trailer are added.
type File struct {
	Header Header
}

// Header is the fixed 8-byte TLV1 file header.
//
// Wire layout (big-endian):
//
//	offset 0  Magic    [4]byte  ASCII "TLV1"
//	offset 4  Version  uint8    must be 1
//	offset 5  Flags    uint8    bit field; see Flags constants
//	offset 6  Reserved uint16   must be 0
type Header struct {
	Magic    [4]byte
	Version  uint8
	Flags    Flags
	Reserved uint16
}

// Flags is the TLV1 Header.Flags bit field.
type Flags uint8

const (
	// FlagCompressed indicates record values are zlib-compressed.
	FlagCompressed Flags = 0x01
	// FlagEncrypted indicates record values are AES-encrypted.
	FlagEncrypted Flags = 0x02
	// FlagSigned indicates the trailer carries a signature in addition to the CRC.
	FlagSigned Flags = 0x04

	// flagsReservedMask covers bits 3-7 which must be zero.
	flagsReservedMask Flags = 0xF8
)

// Has reports whether all bits in mask are set in f.
func (f Flags) Has(mask Flags) bool { return f&mask == mask }

// String returns a human-readable representation of the set flags.
func (f Flags) String() string {
	if f == 0 {
		return "0"
	}
	parts := ""
	add := func(name string) {
		if parts == "" {
			parts = name
			return
		}
		parts += "|" + name
	}
	if f.Has(FlagCompressed) {
		add("COMPRESSED")
	}
	if f.Has(FlagEncrypted) {
		add("ENCRYPTED")
	}
	if f.Has(FlagSigned) {
		add("SIGNED")
	}
	if rem := f &^ (FlagCompressed | FlagEncrypted | FlagSigned); rem != 0 {
		add(fmt.Sprintf("0x%02x", uint8(rem)))
	}
	return parts
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrInvalidMagic is returned when Header.Magic does not match the ASCII bytes "TLV1".
var ErrInvalidMagic = errors.New("tlv: invalid magic")

// ErrUnsupportedVersion is returned when Header.Version is not a recognized value.
var ErrUnsupportedVersion = errors.New("tlv: unsupported version")

// ErrReservedBitsSet is returned when reserved bits in Header.Flags or Header.Reserved are non-zero.
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
