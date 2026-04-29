// Copyright (c) 2026 Z5labs and Contributors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gzip

import (
	"errors"
	"fmt"
)

// File is the placeholder root type for a parsed gzip file. The real
// implementation should replace Placeholder with the actual top-level
// fields described in the gzip SPEC.md (e.g. Header, Members, Trailer).
type File struct {
	// Placeholder is a stand-in field so the struct compiles before
	// the real wire fields are added.
	Placeholder uint8
}

// Kind is an example enum type illustrating the typed-integer + const
// block + String() pattern. Replace with a real gzip enum (for example
// the compression method or operating-system identifier) when filling
// in the spec.
type Kind uint8

const (
	// KindUnknown is the zero value placeholder.
	KindUnknown Kind = 0
	// KindExample is a placeholder enum member.
	KindExample Kind = 1
)

// String implements fmt.Stringer for Kind.
func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "unknown"
	case KindExample:
		return "example"
	default:
		return fmt.Sprintf("Kind(%d)", uint8(k))
	}
}

// Bit field placeholder constants. Replace these with the real gzip
// FLG byte masks (FTEXT, FHCRC, FEXTRA, FNAME, FCOMMENT) when the
// types are filled in.
const (
	// flagsExampleMask masks the example flag bit.
	flagsExampleMask uint8 = 0x01
	// flagsExampleShift is the bit position of the example flag.
	flagsExampleShift uint8 = 0
)

// ErrInvalid is the sentinel error returned when a value violates the
// gzip wire format in a way that does not need extra context.
var ErrInvalid = errors.New("gzip: invalid value")

// InvalidFieldError is returned when a specific field carries an
// illegal value. It demonstrates the struct-error pattern with
// Error() and Unwrap().
type InvalidFieldError struct {
	Field string
	Got   uint64
	Err   error
}

// Error implements the error interface.
func (e *InvalidFieldError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("gzip: invalid field %s = %d: %v", e.Field, e.Got, e.Err)
	}
	return fmt.Sprintf("gzip: invalid field %s = %d", e.Field, e.Got)
}

// Unwrap returns the wrapped error so callers can use errors.Is/As.
func (e *InvalidFieldError) Unwrap() error {
	return e.Err
}
