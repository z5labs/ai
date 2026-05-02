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

// Record is a single Type-Length-Value record from a TLV1 file. Type is one of
// the RecordType constants; Length is the byte length of Value (0 ≤ Length ≤
// 65535); Value is the raw payload, interpretation of which depends on Type.
type Record struct {
	Type   RecordType
	Length uint16
	Value  []byte
}

// RecordType identifies the kind of payload carried by a Record.Value.
type RecordType uint8

// RecordType values per the TLV1 spec's Encoding Tables.
const (
	RecordTypeSTRING RecordType = 0x01
	RecordTypeINT    RecordType = 0x02
	RecordTypeBLOB   RecordType = 0x03
	RecordTypeNESTED RecordType = 0x04
)

// String returns the human-readable name of the RecordType.
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

// ErrUnknownRecordType is returned when a Record.Type byte does not match any
// known RecordType value. The decoder wraps this in a FieldError → OffsetError
// chain so callers can choose to skip or fail per the spec.
var ErrUnknownRecordType = errors.New("tlv: unknown record type")

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
