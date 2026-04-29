package tlv

import (
	"errors"
	"fmt"
)

// File is the top-level decoded representation of a TLV1 file.
// Implementer fills this in as Header, Records, and Trailer are added.
type File struct {
	// Trailer carries the 4-byte CRC32 (IEEE) covering every byte preceding it.
	Trailer Trailer
}

// Trailer is the fixed-size 4-byte file trailer. The CRC32 is computed (IEEE
// polynomial, big-endian on the wire) over every byte from offset 0 of the
// file up to (but not including) the trailer itself.
type Trailer struct {
	CRC32 uint32
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrChecksumMismatch is returned by Decode when the CRC32 carried in the
// trailer does not match the running CRC32 computed over the preceding bytes.
var ErrChecksumMismatch = errors.New("tlv: checksum mismatch")

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
