package tlv

import (
	"errors"
	"fmt"
	"strings"
)

// Magic is the 4-byte ASCII identifier "TLV1" that opens every TLV1 file.
var Magic = [4]byte{0x54, 0x4C, 0x56, 0x31}

// Version1 is the only currently defined TLV1 format version.
const Version1 uint8 = 1

// Flags is the bit field stored in Header.Flags. Each bit toggles an
// independent boolean flag; see the Flag* constants for the defined bits.
type Flags uint8

// Defined Header.Flags bits per SPEC.md "Header.Flags bit field".
const (
	// FlagCompressed: record values are zlib-compressed.
	FlagCompressed Flags = 0x01
	// FlagEncrypted: record values are AES-encrypted.
	FlagEncrypted Flags = 0x02
	// FlagSigned: the trailer carries a signature in addition to the CRC.
	FlagSigned Flags = 0x04
)

// flagsReservedMask covers bits 3-7 of the Flags byte. These bits must be
// zero on both write and read; a non-zero reserved bit surfaces as
// ErrReservedFlagsSet.
const flagsReservedMask Flags = 0xF8

// String renders Flags as a pipe-separated list of set flag names, e.g.
// "COMPRESSED|SIGNED". When no flags are set it returns "NONE". Unknown
// (reserved) bits are appended as "UNKNOWN(0xNN)".
func (f Flags) String() string {
	if f == 0 {
		return "NONE"
	}
	var parts []string
	if f&FlagCompressed != 0 {
		parts = append(parts, "COMPRESSED")
	}
	if f&FlagEncrypted != 0 {
		parts = append(parts, "ENCRYPTED")
	}
	if f&FlagSigned != 0 {
		parts = append(parts, "SIGNED")
	}
	if reserved := f & flagsReservedMask; reserved != 0 {
		parts = append(parts, fmt.Sprintf("UNKNOWN(0x%02X)", uint8(reserved)))
	}
	return strings.Join(parts, "|")
}

// Header is the fixed 8-byte header at the start of every TLV1 file.
//
// Wire layout (big-endian, total 8 bytes):
//
//	Magic    [4]byte // "TLV1"
//	Version  uint8   // must be Version1
//	Flags    Flags   // bit field
//	Reserved uint16  // must be 0
type Header struct {
	Magic    [4]byte
	Version  uint8
	Flags    Flags
	Reserved uint16
}

// File is the top-level decoded representation of a TLV1 file.
// Records and Trailer are not yet implemented; only Header is wired up.
type File struct {
	Header Header
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrInvalidMagic is returned when Header.Magic is not the ASCII bytes "TLV1".
var ErrInvalidMagic = errors.New("tlv: invalid magic")

// ErrUnsupportedVersion is returned when Header.Version is not a recognized
// TLV1 format version.
var ErrUnsupportedVersion = errors.New("tlv: unsupported version")

// ErrReservedFlagsSet is returned when any of the reserved bits (3-7) in
// Header.Flags is non-zero.
var ErrReservedFlagsSet = errors.New("tlv: reserved flag bits set")

// ErrReservedNonZero is returned when Header.Reserved is not zero.
var ErrReservedNonZero = errors.New("tlv: reserved field is non-zero")

// ErrSignedFlagUnsupported is returned when the SIGNED flag is set; the
// signed-trailer extension is reserved for a future revision and is not
// supported by this version of the decoder.
var ErrSignedFlagUnsupported = errors.New("tlv: SIGNED flag is reserved for future use")

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
