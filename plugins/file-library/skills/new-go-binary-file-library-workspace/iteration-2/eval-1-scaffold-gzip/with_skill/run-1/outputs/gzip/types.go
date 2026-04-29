// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package gzip

import (
	"errors"
	"fmt"
)

// File is the placeholder root type produced by Decode and consumed by Encode.
//
// Replace the placeholder field below with the real top-level structure of the
// gzip format once SPEC.md is filled in. Field order should match the wire
// order so a reader can mentally line them up against a hex dump.
type File struct {
	// Placeholder. Replace with the real header / blocks / trailer fields.
	Placeholder uint8
}

// Kind is an example enum that demonstrates the typed-integer + String()
// pattern for any kind / compression-method / flag-byte fields the format
// defines.
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

// Example bit-field placeholder. Real formats (gzip's FLG byte is one) often
// pack multiple flags into a single byte; expose mask and shift constants and
// (optionally) accessor methods on the parent struct rather than splitting the
// byte across multiple Go fields. The wire layout has to round-trip cleanly.
//
//nolint:unused // placeholder; replace with real bit-field constants.
const (
	// flagsExampleMask is a placeholder mask; replace with the real bit layout.
	flagsExampleMask uint8 = 0x01
	// flagsExampleShift is a placeholder shift; replace with the real bit position.
	flagsExampleShift uint8 = 0
)

// ErrInvalid is the sentinel error returned when the decoder rejects an input
// for a stable, comparable reason (illegal magic, unsupported compression
// method, etc.).
//
// Use errors.Is(err, ErrInvalid) to test for this condition. Wrap with the
// FieldError / OffsetError chain (via decoder.wrapErr / encoder.wrapErr) so the
// caller also gets the failing field path and byte offset.
var ErrInvalid = errors.New("invalid gzip input")

// errUnimplemented is the placeholder error returned by the scaffolded
// readFile / writeFile stubs. Tests assert the FieldError → OffsetError → leaf
// chain via errors.Is(err, errUnimplemented). Remove it once the real decoder
// and encoder logic is in place.
var errUnimplemented = errors.New("unimplemented")

// OffsetError wraps a leaf error with the byte offset in the input/output
// stream where the read or write failed. It is the inner layer of the
// decode/encode error chain:
//
//	FieldError{Field: "Header.Length"} -> OffsetError{Offset: 4} -> <leaf>
//
// errors.As(err, &oe) extracts the byte offset; errors.Is(err, leaf) still
// works because Unwrap exposes the underlying error.
type OffsetError struct {
	// Offset is the byte position in the stream where the failure occurred.
	Offset int64
	// Err is the underlying error (a sentinel like io.ErrUnexpectedEOF, or a
	// typed format error).
	Err error
}

// Error implements the error interface.
func (e *OffsetError) Error() string {
	return fmt.Sprintf("at byte %d: %v", e.Offset, e.Err)
}

// Unwrap exposes the underlying error so errors.Is and errors.As walk the
// chain.
func (e *OffsetError) Unwrap() error {
	return e.Err
}

// FieldError wraps an error with the dotted path of the field being decoded
// or encoded when the failure occurred. It is the outer layer of the
// decode/encode error chain:
//
//	FieldError{Field: "Header.Length"} -> OffsetError{Offset: 4} -> <leaf>
//
// errors.As(err, &fe) extracts the field path; nested fields use a dotted
// path so callers can grep ("Header.Length", "Trailer.CRC32").
type FieldError struct {
	// Field is the dotted path to the field whose read or write failed.
	Field string
	// Err is the underlying error (typically *OffsetError wrapping a leaf).
	Err error
}

// Error implements the error interface.
func (e *FieldError) Error() string {
	return fmt.Sprintf("decoding %s: %v", e.Field, e.Err)
}

// Unwrap exposes the underlying error so errors.Is and errors.As walk the
// chain.
func (e *FieldError) Unwrap() error {
	return e.Err
}
