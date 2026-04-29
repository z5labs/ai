package tlv

import (
	"errors"
	"fmt"
)

// Magic is the 4-byte ASCII identifier at the start of every TLV1 file ("TLV1").
var Magic = [4]byte{'T', 'L', 'V', '1'}

// Version is the only TLV file format version this package recognizes.
const Version uint8 = 1

// Flags is the bit field carried in Header.Flags.
//
// Bit layout:
//
//	bit 0 (0x01) COMPRESSED - record values are zlib-compressed
//	bit 1 (0x02) ENCRYPTED  - record values are AES-encrypted
//	bit 2 (0x04) SIGNED     - the trailer carries a signature in addition to the CRC
//	bits 3-7   (0xF8)       - reserved, must be zero
type Flags uint8

const (
	// FlagCompressed indicates record values are zlib-compressed.
	FlagCompressed Flags = 0x01
	// FlagEncrypted indicates record values are AES-encrypted.
	FlagEncrypted Flags = 0x02
	// FlagSigned indicates the trailer carries a signature in addition to the CRC.
	FlagSigned Flags = 0x04

	// flagsReservedMask masks the reserved (must-be-zero) bits of Flags.
	flagsReservedMask Flags = 0xF8
)

// Compressed reports whether the COMPRESSED flag is set.
func (f Flags) Compressed() bool { return f&FlagCompressed != 0 }

// Encrypted reports whether the ENCRYPTED flag is set.
func (f Flags) Encrypted() bool { return f&FlagEncrypted != 0 }

// Signed reports whether the SIGNED flag is set.
func (f Flags) Signed() bool { return f&FlagSigned != 0 }

// Header is the fixed-size 8-byte preamble of a TLV1 file.
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

// errUnimplemented is returned by stub methods so tests can assert the error
// chain via errors.Is even before the real implementation lands.
var errUnimplemented = errors.New("tlv: unimplemented")

// Sentinel errors for typed leaf failures so callers can branch with errors.Is.
var (
	// ErrBadMagic is the leaf error for a header whose Magic bytes are not "TLV1".
	ErrBadMagic = errors.New("tlv: bad magic")
	// ErrUnsupportedVersion is the leaf error for a header whose Version is not 1.
	ErrUnsupportedVersion = errors.New("tlv: unsupported version")
	// ErrReservedFlagSet is the leaf error for a header whose Flags has reserved bits set.
	ErrReservedFlagSet = errors.New("tlv: reserved flag bit set")
	// ErrSignedFlagUnsupported is the leaf error for a SIGNED flag we do not yet support.
	ErrSignedFlagUnsupported = errors.New("tlv: SIGNED flag is not supported by this version")
	// ErrReservedNotZero is the leaf error for a header whose Reserved field is non-zero.
	ErrReservedNotZero = errors.New("tlv: header reserved field must be zero")
)

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
