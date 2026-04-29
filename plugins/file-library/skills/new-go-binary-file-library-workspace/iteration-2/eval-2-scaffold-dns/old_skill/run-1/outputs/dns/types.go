// Copyright (c) 2026 z5labs
//
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file or at https://opensource.org/licenses/MIT.

package dns

import (
	"errors"
	"fmt"
)

// File is the top-level DNS message placeholder. Replace this struct with the
// real DNS message structure (Header, Questions, Answers, Authorities,
// Additionals) when the spec is implemented.
type File struct {
	// Placeholder is a stand-in for the real DNS message fields.
	Placeholder uint16
}

// Kind is an example enum demonstrating the typed-integer + String() pattern.
// Replace with a real DNS enum (e.g., Opcode, RCode, QType, QClass) when the
// spec is implemented.
type Kind uint8

const (
	// KindUnknown is the zero value for Kind.
	KindUnknown Kind = 0
	// KindExample is a placeholder enum value to demonstrate the pattern.
	KindExample Kind = 1
)

// String implements fmt.Stringer for Kind.
func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "Unknown"
	case KindExample:
		return "Example"
	default:
		return fmt.Sprintf("Kind(%d)", uint8(k))
	}
}

// Example bit-field mask/shift constants.
//
// DNS Header.Flags is a uint16 packing several sub-fields. The pattern below
// is a placeholder showing how to expose mask + shift constants for bit-field
// access. Replace with the real DNS flag layout (QR, Opcode, AA, TC, RD, RA,
// Z, RCode) when the spec is implemented.
const (
	// flagsExampleMask is a placeholder mask covering the high bit.
	flagsExampleMask uint16 = 0x8000
	// flagsExampleShift is the bit position of the example field.
	flagsExampleShift uint16 = 15
)

// ErrInvalid is a sentinel error returned when a value violates the DNS wire
// format invariants. Wrap it with fmt.Errorf when raising it to add structural
// context.
var ErrInvalid = errors.New("dns: invalid value")

// InvalidFieldError reports a wire-format value that is structurally illegal
// for a known field. It wraps ErrInvalid so callers can use errors.Is.
type InvalidFieldError struct {
	// Field is the dotted struct path of the offending field, e.g.
	// "Header.Flags".
	Field string
	// Got is the value that was read from the wire.
	Got uint64
}

// Error implements the error interface.
func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("dns: invalid value %d for field %s", e.Got, e.Field)
}

// Unwrap returns ErrInvalid so errors.Is(err, ErrInvalid) succeeds.
func (e *InvalidFieldError) Unwrap() error {
	return ErrInvalid
}
