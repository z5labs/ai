// Copyright (c) 2026 Z5labs and Contributors
//
// Licensed under the MIT License. See LICENSE file in the project root
// for full license information.

package gzip

import (
	"errors"
	"fmt"
)

// File is the placeholder root type for a decoded gzip file. Replace the
// placeholder field with the real header, optional fields, compressed data,
// and trailer once SPEC.md defines them.
type File struct {
	// Placeholder reserves a single byte so the struct has fixed size.
	// Real implementations should replace this with the gzip header
	// (ID1/ID2/CM/FLG/MTIME/XFL/OS), optional extra fields, compressed
	// blocks, and the CRC32/ISIZE trailer.
	Placeholder uint8
}

// Kind is an example enum demonstrating the typed-integer + String() pattern
// used for gzip enums such as the compression method (CM) or operating
// system (OS) byte. Replace with the real enums during implementation.
type Kind uint8

// Example enum values. Replace with the real gzip enum members
// (e.g. CMDeflate = 8, OSFAT = 0, OSUnix = 3).
const (
	KindUnknown Kind = 0
	KindExample Kind = 1
)

// String returns a human-readable name for the Kind. The pattern pays for
// itself the first time a test failure prints a Kind in a hex dump.
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

// Example bit field mask/shift constants. The gzip FLG byte packs FTEXT,
// FHCRC, FEXTRA, FNAME, and FCOMMENT into a single uint8; the implementer
// replaces these placeholders with the real masks during implementation.
const (
	// flagsExampleMask masks the lowest bit of an example flags byte.
	flagsExampleMask = 0x01
	// flagsExampleShift is the bit position of the example flag.
	flagsExampleShift = 0
)

// ErrInvalid is the sentinel error returned when a gzip file fails a
// stable, comparable validity check (for example, a wrong magic number).
var ErrInvalid = errors.New("gzip: invalid")

// InvalidFieldError is the structured error returned when a specific field
// holds an illegal value. Use it when context (which field, what value)
// helps the caller diagnose a malformed input.
type InvalidFieldError struct {
	Field string
	Got   uint32
	Err   error
}

// Error implements the error interface.
func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("gzip: invalid value %d for field %s: %v", e.Got, e.Field, e.Err)
}

// Unwrap returns the wrapped cause so errors.Is/errors.As work transitively.
func (e *InvalidFieldError) Unwrap() error {
	return e.Err
}
