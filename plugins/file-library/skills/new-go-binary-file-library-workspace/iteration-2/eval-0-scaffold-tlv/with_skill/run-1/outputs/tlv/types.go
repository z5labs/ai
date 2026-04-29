// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

import (
	"errors"
	"fmt"
)

// File is the placeholder root type produced by Decode and consumed by Encode.
//
// Replace the placeholder field below with the real top-level structure of the
// TLV format once SPEC.md is filled in. Field order should match the wire
// order so a reader can mentally line them up against a hex dump.
type File struct {
	// Placeholder. Replace with the real records / sections / header fields.
	Placeholder uint8
}

// Kind is an example enum that demonstrates the typed-integer + String()
// pattern for any kind/opcode/tag fields the format defines.
//
// Replace the constants below with the real wire values from SPEC.md.
type Kind uint8

const (
	// KindUnknown is the zero value and represents an unset or unrecognized kind.
	KindUnknown Kind = 0
	// KindExample is a placeholder enum value to demonstrate the pattern.
	KindExample Kind = 1
)

// String returns a human-readable name for the Kind. This pays for itself the
// first time a test failure prints a Kind value or a hex dump references one.
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

// Example bit-field placeholder. Real formats often pack multiple flags into a
// single byte; expose mask and shift constants and (optionally) accessor
// methods on the parent struct rather than splitting the byte across multiple
// Go fields. The wire layout has to round-trip cleanly.
//
// These are unused until the implementer wires them up against real bit
// fields in SPEC.md.
const (
	// flagsExampleMask is a placeholder mask; replace with the real bit layout.
	flagsExampleMask uint8 = 0x01
	// flagsExampleShift is a placeholder shift; replace with the real bit position.
	flagsExampleShift uint8 = 0
)

// ErrInvalid is the sentinel error returned when the decoder rejects an input
// for a stable, comparable reason (illegal value, unsupported version, etc.).
//
// Use errors.Is(err, ErrInvalid) to test for this condition.
var ErrInvalid = errors.New("invalid TLV input")

// errUnimplemented is the sentinel returned by the decoder/encoder stubs while
// the format details are still placeholders. Tests assert the full error chain
// (FieldError -> OffsetError -> errUnimplemented) survives errors.Is, which
// pins the chain shape down before the implementer fills in real reads/writes.
var errUnimplemented = errors.New("unimplemented")

// OffsetError carries the byte offset at which a decode/encode failure
// happened. It always sits inside a FieldError; callers can pull the offset
// out with errors.As(err, &oe).
type OffsetError struct {
	// Offset is the byte offset within the input/output stream at which the
	// failure was detected.
	Offset int64
	// Err is the underlying error.
	Err error
}

// Error implements the error interface.
func (e *OffsetError) Error() string {
	return fmt.Sprintf("at byte %d: %s", e.Offset, e.Err)
}

// Unwrap returns the wrapped error so errors.Is/errors.As can walk the chain.
func (e *OffsetError) Unwrap() error {
	return e.Err
}

// FieldError carries the dotted field path at which a decode/encode failure
// happened (e.g. "Header.Length"). It is always the outermost wrapper in the
// error chain; callers can pull the field path out with errors.As(err, &fe).
type FieldError struct {
	// Field is the dotted path to the offending field (e.g. "Header.Length").
	Field string
	// Err is the underlying error (typically an *OffsetError).
	Err error
}

// Error implements the error interface.
func (e *FieldError) Error() string {
	return fmt.Sprintf("decoding %s: %s", e.Field, e.Err)
}

// Unwrap returns the wrapped error so errors.Is/errors.As can walk the chain.
func (e *FieldError) Unwrap() error {
	return e.Err
}
