package tlv

import (
	"errors"
	"fmt"
)

// File is the top-level decoded representation of a TLV1 file.
// Implementer fills this in as Header, Records, and Trailer are added.
type File struct {
	// Trailer carries the 4-byte CRC32 footer covering every byte of the
	// file preceding the trailer itself.
	Trailer Trailer
}

// Trailer is the fixed-size 4-byte footer of a TLV1 file. It carries the
// IEEE CRC32 of every byte preceding the trailer.
type Trailer struct {
	CRC32 uint32
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrChecksumMismatch is returned when the CRC32 in the trailer does not
// match the CRC32 computed over the preceding bytes during decode.
var ErrChecksumMismatch = errors.New("tlv: trailer CRC32 mismatch")

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
