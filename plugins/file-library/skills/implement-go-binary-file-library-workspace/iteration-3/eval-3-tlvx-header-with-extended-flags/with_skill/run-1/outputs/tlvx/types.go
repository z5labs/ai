package tlvx

import (
	"errors"
	"fmt"
)

// File is the top-level decoded representation of a TLVX file.
// Implementer fills this in as Header, Records, and Trailer are added.
type File struct {
	Header Header
}

// Header is the fixed-size 16-byte header that opens every TLVX file.
//
// Wire layout (big-endian):
//
//	offset 0:  Magic         [4]byte ASCII "TLVX"
//	offset 4:  Version       uint8   must be 1
//	offset 5:  Flags         uint8   bit field; see Flags constants
//	offset 6:  ChecksumAlg   uint8   see ChecksumAlg constants
//	offset 7:  Reserved1     uint8   must be 0
//	offset 8:  IndexCount    uint16
//	offset 10: ExtCount      uint16
//	offset 12: TrailerOffset uint32
type Header struct {
	Magic         [4]byte
	Version       uint8
	Flags         Flags
	ChecksumAlg   ChecksumAlg
	Reserved1     uint8
	IndexCount    uint16
	ExtCount      uint16
	TrailerOffset uint32
}

// MagicTLVX is the ASCII "TLVX" sequence that opens every TLVX file.
var MagicTLVX = [4]byte{'T', 'L', 'V', 'X'}

// Version1 is the only currently-defined TLVX format version.
const Version1 uint8 = 1

// Flags is the bit field stored in Header.Flags.
type Flags uint8

// Header.Flags bit definitions. Bit 7 (0x80) is reserved and must be 0.
const (
	FlagCompressed Flags = 0x01 // bit 0: record values are zlib-compressed
	FlagEncrypted  Flags = 0x02 // bit 1: record values are AES-encrypted
	FlagSigned     Flags = 0x04 // bit 2: trailer carries a 64-byte signature
	FlagIndexed    Flags = 0x08 // bit 3: must be set iff IndexCount > 0
	FlagExtended   Flags = 0x10 // bit 4: must be set iff ExtCount > 0
	FlagStrict     Flags = 0x20 // bit 5: unknown record types are a hard error
	FlagSealed     Flags = 0x40 // bit 6: no further records may be appended
)

// flagsReservedMask covers bit 7, which must always be zero.
const flagsReservedMask Flags = 0x80

// ChecksumAlg is the single-byte tag identifying the algorithm used for the
// body and trailer checksums.
type ChecksumAlg uint8

// Defined ChecksumAlg values. See SPEC.md "Encoding Tables → Checksum algorithms".
const (
	ChecksumAlgCRC32IEEE ChecksumAlg = 0x01
	ChecksumAlgCRC64ECMA ChecksumAlg = 0x02
	ChecksumAlgSHA256T32 ChecksumAlg = 0x03
	ChecksumAlgXXH64     ChecksumAlg = 0x04
	ChecksumAlgBLAKE3T32 ChecksumAlg = 0x05
)

// String renders the algorithm tag using the spec's short name.
func (a ChecksumAlg) String() string {
	switch a {
	case ChecksumAlgCRC32IEEE:
		return "CRC32_IEEE"
	case ChecksumAlgCRC64ECMA:
		return "CRC64_ECMA"
	case ChecksumAlgSHA256T32:
		return "SHA256_T32"
	case ChecksumAlgXXH64:
		return "XXH64"
	case ChecksumAlgBLAKE3T32:
		return "BLAKE3_T32"
	default:
		return fmt.Sprintf("ChecksumAlg(0x%02x)", uint8(a))
	}
}

// ErrInvalid is returned when the input bytes do not match the TLVX format.
var ErrInvalid = errors.New("tlvx: invalid file")

// errUnimplemented is returned by stub methods so tests can assert the error
// chain via errors.Is even before the real implementation lands.
var errUnimplemented = errors.New("tlvx: unimplemented")

// UnexpectedMagicError is returned when the four magic bytes that open the
// file do not match the TLVX magic.
type UnexpectedMagicError struct {
	Got [4]byte
}

func (e *UnexpectedMagicError) Error() string {
	return fmt.Sprintf("tlvx: unexpected magic %#v, want %#v", e.Got, MagicTLVX)
}

// UnknownVersionError is returned when Header.Version is not a recognised
// TLVX format version.
type UnknownVersionError struct {
	Version uint8
}

func (e *UnknownVersionError) Error() string {
	return fmt.Sprintf("tlvx: unknown version %d", e.Version)
}

// UnknownChecksumAlgError is returned when Header.ChecksumAlg is not a
// recognised algorithm tag.
type UnknownChecksumAlgError struct {
	Alg ChecksumAlg
}

func (e *UnknownChecksumAlgError) Error() string {
	return fmt.Sprintf("tlvx: unknown checksum algorithm %s", e.Alg)
}

// ReservedFieldNonZeroError is returned when a reserved field in the header
// or trailer carries a non-zero value.
type ReservedFieldNonZeroError struct {
	Field string
	Got   uint8
}

func (e *ReservedFieldNonZeroError) Error() string {
	return fmt.Sprintf("tlvx: reserved field %s must be 0, got %d", e.Field, e.Got)
}

// ReservedFlagBitError is returned when a reserved bit in a flags byte is set.
type ReservedFlagBitError struct {
	Field string
	Bit   uint8
}

func (e *ReservedFlagBitError) Error() string {
	return fmt.Sprintf("tlvx: reserved bit %d in %s is set", e.Bit, e.Field)
}

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
