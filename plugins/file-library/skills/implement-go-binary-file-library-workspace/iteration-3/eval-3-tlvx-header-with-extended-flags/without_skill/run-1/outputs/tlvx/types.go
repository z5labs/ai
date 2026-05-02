package tlvx

import (
	"errors"
	"fmt"
)

// Format-wide constants pulled from SPEC.md "Header" and the file's overall
// shape.
const (
	// HeaderSize is the fixed on-disk size of the TLVX header in bytes.
	HeaderSize = 16

	// Version is the only TLVX format version currently defined.
	Version uint8 = 1
)

// Magic is the four-byte ASCII tag at the start of every TLVX file
// ("TLVX" — 0x54 0x4C 0x56 0x58).
var Magic = [4]byte{'T', 'L', 'V', 'X'}

// Flags is the bit field stored in Header.Flags. See SPEC.md
// "Header.Flags bit field". Bit 7 is reserved and must remain zero.
type Flags uint8

// Header.Flags bits, matching SPEC.md.
const (
	FlagCompressed Flags = 1 << 0 // 0x01
	FlagEncrypted  Flags = 1 << 1 // 0x02
	FlagSigned     Flags = 1 << 2 // 0x04
	FlagIndexed    Flags = 1 << 3 // 0x08
	FlagExtended   Flags = 1 << 4 // 0x10
	FlagStrict     Flags = 1 << 5 // 0x20
	FlagSealed     Flags = 1 << 6 // 0x40
)

// Has reports whether every bit in mask is set in f.
func (f Flags) Has(mask Flags) bool { return f&mask == mask }

// ChecksumAlg is the single-byte tag identifying the checksum algorithm used
// for the body and trailer checksums. See SPEC.md "Encoding Tables" →
// "Checksum algorithms".
type ChecksumAlg uint8

// Checksum algorithm tags, matching SPEC.md.
const (
	ChecksumCRC32IEEE ChecksumAlg = 0x01
	ChecksumCRC64ECMA ChecksumAlg = 0x02
	ChecksumSHA256T32 ChecksumAlg = 0x03
	ChecksumXXH64     ChecksumAlg = 0x04
	ChecksumBLAKE3T32 ChecksumAlg = 0x05
)

// Header is the fixed-size 16-byte TLVX header. It identifies the file,
// carries format-wide flags, and points at the trailer.
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

// File is the top-level decoded representation of a TLVX file.
// Implementer fills this in further as Index, Records, ExtTab, and Trailer
// are added.
type File struct {
	Header Header
}

// ErrInvalid is returned when the input bytes do not match the TLVX format.
var ErrInvalid = errors.New("tlvx: invalid file")

// ErrMagicMismatch is the leaf error returned when the four magic bytes do
// not equal "TLVX".
var ErrMagicMismatch = errors.New("tlvx: magic mismatch")

// ErrUnsupportedVersion is the leaf error returned when Header.Version is
// not 1.
var ErrUnsupportedVersion = errors.New("tlvx: unsupported version")

// ErrReservedNonZero is the leaf error returned when a reserved field is
// non-zero on read.
var ErrReservedNonZero = errors.New("tlvx: reserved field is non-zero")

// errUnimplemented is retained for stub paths that have not yet been wired
// up; current Header implementation does not return it.
var errUnimplemented = errors.New("tlvx: unimplemented")

// OffsetError records the byte offset at which a decode or encode error occurred.
type OffsetError struct {
	Offset int64
	Err    error
}

func (e *OffsetError) Error() string { return fmt.Sprintf("at byte %d: %v", e.Offset, e.Err) }
func (e *OffsetError) Unwrap() error { return e.Err }

// FieldError records the field path (e.g. "Header.Magic") at which an error occurred.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string { return fmt.Sprintf("decoding %s: %v", e.Field, e.Err) }
func (e *FieldError) Unwrap() error { return e.Err }
