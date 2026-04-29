package tlv

import (
	"errors"
	"fmt"
)

// File is the top-level decoded representation of a TLV1 file.
//
// Until Header and Records are implemented, Payload holds the raw bytes that
// precede the Trailer. The Trailer's CRC32 is computed over those bytes.
type File struct {
	// Payload is every byte of the file that precedes the Trailer.
	Payload []byte
	// Trailer carries the integrity CRC32 over Payload.
	Trailer Trailer
}

// Trailer is the fixed-size 4-byte trailer of a TLV1 file.
// CRC32 is the IEEE CRC32 of every byte preceding the trailer.
type Trailer struct {
	CRC32 uint32
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrChecksumMismatch is returned when the CRC32 stored in the trailer does not
// match the CRC32 computed over the bytes that precede it.
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
