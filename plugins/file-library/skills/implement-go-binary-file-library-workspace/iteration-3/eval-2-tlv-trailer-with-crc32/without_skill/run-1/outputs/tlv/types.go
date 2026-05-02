package tlv

import (
	"errors"
	"fmt"
)

// File is the top-level decoded representation of a TLV1 file.
type File struct {
	Header  Header
	Records []Record
	Trailer Trailer
}

// Header is the fixed 8-byte file header.
type Header struct {
	Magic    [4]byte
	Version  uint8
	Flags    uint8
	Reserved uint16
}

// Record is a single TLV record.
type Record struct {
	Type   uint8
	Length uint16
	Value  []byte
}

// Trailer is the fixed 4-byte file trailer holding the CRC32 (IEEE) of every
// byte preceding the trailer.
type Trailer struct {
	CRC32 uint32
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

// ErrChecksumMismatch is returned when the CRC32 in the trailer does not
// match the CRC32 computed over the bytes preceding the trailer.
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
