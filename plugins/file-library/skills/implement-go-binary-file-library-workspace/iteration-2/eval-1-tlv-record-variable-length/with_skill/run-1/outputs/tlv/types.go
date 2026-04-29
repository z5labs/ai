package tlv

import (
	"errors"
	"fmt"
)

// File is the top-level decoded representation of a TLV1 file.
// Implementer fills this in as Header, Records, and Trailer are added.
type File struct {
	Records []Record
}

// RecordType identifies the kind of payload carried by a Record. The values
// match the encoding table in SPEC.md (Encoding Tables → Record.Type values).
type RecordType uint8

const (
	// RecordTypeSTRING is a UTF-8 string payload (no null terminator).
	RecordTypeSTRING RecordType = 0x01
	// RecordTypeINT is a big-endian signed 64-bit integer (Length must be 8).
	RecordTypeINT RecordType = 0x02
	// RecordTypeBLOB is an opaque byte payload.
	RecordTypeBLOB RecordType = 0x03
	// RecordTypeNESTED is a nested TLV1 file (header + records + trailer).
	RecordTypeNESTED RecordType = 0x04
)

// String returns the human-readable name of a RecordType. Unknown values are
// rendered as their hexadecimal byte so test failures are easy to read.
func (t RecordType) String() string {
	switch t {
	case RecordTypeSTRING:
		return "STRING"
	case RecordTypeINT:
		return "INT"
	case RecordTypeBLOB:
		return "BLOB"
	case RecordTypeNESTED:
		return "NESTED"
	default:
		return fmt.Sprintf("RecordType(0x%02x)", uint8(t))
	}
}

// Record is a variable-length type-length-value record. Type and Length are
// fixed-width on the wire (1 + 2 bytes); Value is exactly Length bytes long.
type Record struct {
	Type   RecordType
	Length uint16
	Value  []byte
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
