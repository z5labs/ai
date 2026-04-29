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

// Kind is an example enum that demonstrates the typed-integer + String() pattern
// for any kind/type/tag fields the format defines.
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
const (
	// flagsExampleMask is a placeholder mask; replace with the real bit layout.
	flagsExampleMask uint8 = 0x01
	// flagsExampleShift is a placeholder shift; replace with the real bit position.
	flagsExampleShift uint8 = 0
)

// ErrInvalid is the sentinel error returned when the decoder rejects an input
// for a stable, comparable reason (illegal value, unsupported version, etc.).
//
// Use errors.Is(err, ErrInvalid) to test for this condition. Wrap with
// InvalidFieldError when the caller benefits from knowing which field failed.
var ErrInvalid = errors.New("invalid TLV input")

// InvalidFieldError carries field-level context for a violation of the format.
// It wraps ErrInvalid so callers can use errors.Is(err, ErrInvalid).
type InvalidFieldError struct {
	// Field is the dotted path to the offending field (e.g. "Header.Length").
	Field string
	// Reason is a short human-readable description of the violation.
	Reason string
}

// Error implements the error interface.
func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("invalid TLV field %s: %s", e.Field, e.Reason)
}

// Unwrap returns ErrInvalid so errors.Is(err, ErrInvalid) returns true.
func (e *InvalidFieldError) Unwrap() error {
	return ErrInvalid
}
