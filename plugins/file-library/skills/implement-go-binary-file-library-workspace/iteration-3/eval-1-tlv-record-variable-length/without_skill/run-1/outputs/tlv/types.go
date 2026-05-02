package tlv

import (
	"errors"
	"fmt"
)

// RecordType identifies the interpretation of a Record's Value bytes.
// See SPEC.md "Encoding Tables → Record.Type values".
type RecordType uint8

const (
	// RecordTypeSTRING marks the Value as a UTF-8 string (no null terminator).
	RecordTypeSTRING RecordType = 0x01
	// RecordTypeINT marks the Value as a big-endian signed 64-bit integer.
	// Length must be 8.
	RecordTypeINT RecordType = 0x02
	// RecordTypeBLOB marks the Value as an opaque byte payload.
	RecordTypeBLOB RecordType = 0x03
	// RecordTypeNESTED marks the Value as itself a TLV1 file
	// (header + records + trailer).
	RecordTypeNESTED RecordType = 0x04
)

// Record is a single TLV1 record: a one-byte Type, a two-byte big-endian
// Length, and Length bytes of Value. A record with Length == 0 is legal and
// carries an empty Value.
type Record struct {
	Type   RecordType
	Length uint16
	Value  []byte
}

// File is the top-level decoded representation of a TLV1 file.
// Implementer fills this in as Header, Records, and Trailer are added.
type File struct {
	Records []Record
}

// ErrInvalid is returned when the input bytes do not match the TLV1 format.
var ErrInvalid = errors.New("tlv: invalid file")

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
