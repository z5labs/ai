// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package gzip

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKindString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		kind Kind
		want string
	}{
		{name: "unknown", kind: KindUnknown, want: "Unknown"},
		{name: "example", kind: KindExample, want: "Example"},
		{name: "fallback", kind: Kind(99), want: "Kind(99)"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.kind.String())
		})
	}
}

// fixedSizeStruct is a placeholder for verifying that the package's wire
// structs are fixed-size (so binary.Size and binary.Read work directly).
// Replace this with the real header/trailer struct from SPEC.md.
type fixedSizeStruct struct {
	A uint8
	B uint16
	C uint32
}

func TestFixedSize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    any
		want int
	}{
		{name: "fixedSizeStruct", v: fixedSizeStruct{}, want: 7},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, binary.Size(tc.v))
		})
	}
}

// TestErrorChain pins down the FieldError -> OffsetError -> leaf chain shape.
// Every decoder/encoder failure path must produce this shape so callers can
// rely on errors.Is for sentinels and errors.As for the field path / byte
// offset.
func TestErrorChain(t *testing.T) {
	t.Parallel()

	err := &FieldError{
		Field: "Header",
		Err: &OffsetError{
			Offset: 4,
			Err:    errUnimplemented,
		},
	}

	t.Run("errors.Is finds the leaf sentinel", func(t *testing.T) {
		t.Parallel()
		require.ErrorIs(t, err, errUnimplemented)
	})

	t.Run("errors.As extracts the FieldError", func(t *testing.T) {
		t.Parallel()
		var fe *FieldError
		require.True(t, errors.As(err, &fe))
		require.Equal(t, "Header", fe.Field)
	})

	t.Run("errors.As extracts the OffsetError", func(t *testing.T) {
		t.Parallel()
		var oe *OffsetError
		require.True(t, errors.As(err, &oe))
		require.Equal(t, int64(4), oe.Offset)
	})
}
