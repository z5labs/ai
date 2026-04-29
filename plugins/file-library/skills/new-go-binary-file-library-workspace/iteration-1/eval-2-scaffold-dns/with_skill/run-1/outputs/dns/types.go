// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"errors"
	"fmt"
)

// File is the placeholder root type for a DNS wire-format message.
//
// Replace this struct with the real top-level message structure (typically a
// DNS header followed by question, answer, authority, and additional
// sections) once the spec types are defined.
type File struct {
	// Placeholder is a stand-in field so binary.Size and binary.Read work
	// against File while the real layout is being filled in. Remove once
	// real fields are added.
	Placeholder uint16
}

// Kind is an example enum demonstrating the typed-integer + String() pattern
// used for DNS RR types, classes, opcodes, etc.
//
// Replace with the real DNS enums (Type, Class, Opcode, Rcode, ...) once the
// spec is consulted.
type Kind uint8

// Example Kind values. Replace with real enum members from the spec.
const (
	KindUnknown Kind = 0
	KindA       Kind = 1
)

// String implements fmt.Stringer for Kind. Real implementations should cover
// every named constant in the const block above.
func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "Unknown"
	case KindA:
		return "A"
	default:
		return fmt.Sprintf("Kind(%d)", uint8(k))
	}
}

// Example bit-field mask/shift pair.
//
// DNS packs several flags into the 16-bit Flags word of the header. The
// scaffold demonstrates the convention with a single QR (query/response) bit:
// store the underlying uint16, then mask/shift in Go.
//
//	flags := hdr.Flags
//	qr := (flags & flagsQRMask) >> flagsQRShift
//
// Replace with the real bit fields (QR, Opcode, AA, TC, RD, RA, Z, Rcode)
// from RFC 1035 section 4.1.1.
const (
	flagsQRMask  uint16 = 0x8000
	flagsQRShift uint16 = 15
)

// ErrInvalid is the sentinel returned when a wire-format value violates the
// spec in a way that doesn't deserve its own typed error. Use it via
// errors.Is.
var ErrInvalid = errors.New("dns: invalid value")

// InvalidFieldError is returned when a specific named field carries a value
// the spec disallows. Wrapping ErrInvalid keeps errors.Is(err, ErrInvalid)
// working while preserving the field-level context.
type InvalidFieldError struct {
	Field  string
	Got    uint64
	Reason string
}

// Error implements the error interface.
func (e *InvalidFieldError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("dns: invalid value %d for field %q", e.Got, e.Field)
	}
	return fmt.Sprintf("dns: invalid value %d for field %q: %s", e.Got, e.Field, e.Reason)
}

// Unwrap returns ErrInvalid so callers can match with errors.Is.
func (e *InvalidFieldError) Unwrap() error { return ErrInvalid }
