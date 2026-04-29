// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"errors"
	"fmt"
)

// File is the placeholder root type produced by Decode and consumed by
// Encode. It will be replaced with the real DNS message structure once the
// types are filled in from RFC 1035.
type File struct {
	// Placeholder field. Replace with the real top-level fields (e.g. a
	// Header plus question / answer / authority / additional sections).
	Placeholder uint32
}

// Kind is an example enum demonstrating the typed-integer + const + String()
// pattern used throughout this package. Replace with a real DNS enum (e.g.
// Type, Class, Opcode, Rcode) once the spec types are scaffolded.
type Kind uint8

// Example Kind values. These exist solely to demonstrate the enum pattern
// and have no DNS semantics.
const (
	KindUnknown Kind = 0
	KindA       Kind = 1
	KindNS      Kind = 2
)

// String returns a human-readable name for the Kind. Unknown values render
// as "Kind(<n>)" so test failures and hex dumps stay legible.
func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "Unknown"
	case KindA:
		return "A"
	case KindNS:
		return "NS"
	default:
		return fmt.Sprintf("Kind(%d)", uint8(k))
	}
}

// Example bit field mask/shift constants. DNS packs the QR / Opcode / AA /
// TC / RD / RA / Z / Rcode fields into a single 16-bit Flags word in the
// header; the real implementation will replace these placeholders with the
// real masks and shifts. They are defined here (and intentionally unused
// until then) to demonstrate the convention.
const (
	flagsPlaceholderMask  uint16 = 0x8000 //nolint:unused // placeholder
	flagsPlaceholderShift uint8  = 15     //nolint:unused // placeholder
)

// ErrInvalid is the sentinel returned when bytes were read successfully but
// their value violates the DNS wire format (e.g. an out-of-range enum, a
// length prefix that exceeds the remaining message). Use it with errors.Is.
var ErrInvalid = errors.New("invalid")

// errUnimplemented is the placeholder leaf error returned by the decoder
// and encoder stubs. The error chain (FieldError -> OffsetError ->
// errUnimplemented) is real even before the read/write logic is, so tests
// can pin down errors.Is / errors.As behaviour from day one.
var errUnimplemented = errors.New("unimplemented")

// OffsetError records the byte offset at which a decode or encode failure
// occurred. It always wraps another error and is itself wrapped by
// FieldError. Construct it only via decoder.wrapErr / encoder.wrapErr so
// the offset stays consistent.
type OffsetError struct {
	Offset int64
	Err    error
}

// Error formats as "at byte N: <err>".
func (e *OffsetError) Error() string {
	return fmt.Sprintf("at byte %d: %s", e.Offset, e.Err)
}

// Unwrap returns the wrapped leaf error so errors.Is / errors.As can walk
// the chain.
func (e *OffsetError) Unwrap() error {
	return e.Err
}

// FieldError records the dotted path of the field that failed to decode or
// encode (e.g. "Header.Length"). It is the outermost wrapper in the chain
// and always wraps an OffsetError, which in turn wraps the leaf error.
type FieldError struct {
	Field string
	Err   error
}

// Error formats as "decoding <Field>: <err>".
func (e *FieldError) Error() string {
	return fmt.Sprintf("decoding %s: %s", e.Field, e.Err)
}

// Unwrap returns the wrapped OffsetError so errors.Is / errors.As can walk
// the chain.
func (e *FieldError) Unwrap() error {
	return e.Err
}
