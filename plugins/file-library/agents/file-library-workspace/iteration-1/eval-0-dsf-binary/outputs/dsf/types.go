package dsf

import (
	"errors"
	"fmt"
)

// MagicCookie is the 8-byte ASCII tag that opens every uncompressed DSF file.
var MagicCookie = [8]byte{'X', 'P', 'L', 'N', 'E', 'D', 'S', 'F'}

// CurrentVersion is the only DSF master file format version currently defined.
const CurrentVersion uint32 = 1

// File is the top-level decoded representation of a DSF file. The atom payload
// bytes are preserved as opaque []byte slices for now; later iterations will
// decode the per-atom-ID payload into typed sub-structures.
type File struct {
	Header FileHeader
	Atoms  []Atom
	Footer [16]byte
}

// FileHeader is the fixed 12-byte header that opens every DSF file.
type FileHeader struct {
	Cookie  [8]byte
	Version uint32
}

// Atom is the generic atom envelope: a 4-byte little-endian ID, a 4-byte
// little-endian size *including* the 8-byte header, then Size-8 bytes of
// payload. The payload bytes are left opaque here; later iterations will
// decode the per-ID payload structures (PROP, DEFN, GEOD, CMDS, …) on demand.
type Atom struct {
	ID      uint32
	Size    uint32
	Payload []byte
}

// Kind is a placeholder enum demonstrating the String()-method pattern that
// every real enum the implementer adds must follow.
type Kind uint8

const (
	// KindUnknown is the zero value for Kind.
	KindUnknown Kind = 0
	// KindExample is a placeholder enum constant to demonstrate the pattern.
	KindExample Kind = 1
)

// String renders a Kind value using its declared name, falling back to a
// hex-formatted form for unknown values.
func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "Unknown"
	case KindExample:
		return "Example"
	default:
		return fmt.Sprintf("Kind(0x%02x)", uint8(k))
	}
}

// ErrInvalid is returned when the input bytes do not match the DSF format.
var ErrInvalid = errors.New("dsf: invalid file")

// errUnimplemented is retained so future iterations can still pin the error
// chain via errors.Is on stub paths.
var errUnimplemented = errors.New("dsf: unimplemented")

// UnexpectedCookieError is returned when the leading 8 bytes of the file do
// not match MagicCookie.
type UnexpectedCookieError struct {
	Got [8]byte
}

func (e *UnexpectedCookieError) Error() string {
	return fmt.Sprintf("dsf: unexpected cookie %#v, want %#v", e.Got, MagicCookie)
}

// UnknownVersionError is returned when the master file format version is not
// a recognised value.
type UnknownVersionError struct {
	Version uint32
}

func (e *UnknownVersionError) Error() string {
	return fmt.Sprintf("dsf: unknown version %d", e.Version)
}

// AtomSizeTooSmallError is returned when an atom on the wire reports a Size
// less than 8 (the smallest legal size — header bytes only).
type AtomSizeTooSmallError struct {
	Size uint32
}

func (e *AtomSizeTooSmallError) Error() string {
	return fmt.Sprintf("dsf: atom size %d is below the 8-byte minimum", e.Size)
}

// AtomSizeOverflowError is returned when an atom on the wire reports a Size
// that exceeds the bytes still available in the atom region.
type AtomSizeOverflowError struct {
	Size      uint32
	Remaining int64
}

func (e *AtomSizeOverflowError) Error() string {
	return fmt.Sprintf("dsf: atom size %d exceeds %d remaining bytes in atom region", e.Size, e.Remaining)
}

// MD5MismatchError is returned when the footer's MD5 hash does not match the
// hash recomputed across the preceding file bytes.
type MD5MismatchError struct {
	Got  [16]byte
	Want [16]byte
}

func (e *MD5MismatchError) Error() string {
	return fmt.Sprintf("dsf: md5 mismatch: got %x, want %x", e.Got, e.Want)
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
