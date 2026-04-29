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

// RecordType identifies the interpretation of a Record's Value payload.
type RecordType uint8

// Record.Type values per SPEC.md "Encoding Tables".
const (
	RecordTypeString RecordType = 0x01
	RecordTypeInt    RecordType = 0x02
	RecordTypeBlob   RecordType = 0x03
	RecordTypeNested RecordType = 0x04
)

// String returns the spec name of the record type, or a hex fallback for
// unknown values. Useful when reading hex-dump test failures.
func (t RecordType) String() string {
	switch t {
	case RecordTypeString:
		return "STRING"
	case RecordTypeInt:
		return "INT"
	case RecordTypeBlob:
		return "BLOB"
	case RecordTypeNested:
		return "NESTED"
	default:
		return fmt.Sprintf("RecordType(0x%02x)", uint8(t))
	}
}

// Record is a single TLV1 type-length-value record.
//
// Wire layout (big-endian):
//
//	+------+--------+-----------------+
//	| Type | Length | Value           |
//	| u8   | u16    | Length bytes    |
//	+------+--------+-----------------+
//
// A record with Length == 0 is legal and carries an empty Value.
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
