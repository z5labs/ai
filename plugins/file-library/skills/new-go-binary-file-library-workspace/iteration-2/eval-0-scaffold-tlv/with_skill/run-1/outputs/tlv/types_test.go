// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

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
		{
			name: "unknown",
			kind: KindUnknown,
			want: "Unknown",
		},
		{
			name: "example",
			kind: KindExample,
			want: "Example",
		},
		{
			name: "unrecognized falls back to numeric form",
			kind: Kind(99),
			want: "Kind(99)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.kind.String())
		})
	}
}

// fixedSizeExample is a placeholder fixed-size struct used to demonstrate the
// binary.Size() pattern. Replace it with a real fixed-size structure from the
// format once SPEC.md is filled in.
type fixedSizeExample struct {
	A uint8
	B uint16
	C uint32
}

func TestFixedSizeExampleSize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value fixedSizeExample
		want  int
	}{
		{
			name:  "zero value occupies the wire size implied by its fields",
			value: fixedSizeExample{},
			// 1 (uint8) + 2 (uint16) + 4 (uint32) = 7 bytes.
			want: 7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, binary.Size(tc.value))
		})
	}
}

// TestErrorChain pins down the FieldError -> OffsetError -> leaf chain shape.
// The decoder and encoder both build errors via wrapErr, which produces this
// exact shape; if anything in the chain breaks errors.Is or errors.As, this
// test catches it before the implementer notices.
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

	t.Run("errors.As pulls out FieldError", func(t *testing.T) {
		t.Parallel()

		var fe *FieldError
		require.ErrorAs(t, err, &fe)
		require.Equal(t, "Header", fe.Field)
	})

	t.Run("errors.As pulls out OffsetError", func(t *testing.T) {
		t.Parallel()

		var oe *OffsetError
		require.ErrorAs(t, err, &oe)
		require.Equal(t, int64(4), oe.Offset)
	})

	t.Run("the chain is walkable end to end", func(t *testing.T) {
		t.Parallel()

		// Sanity-check: unwrap once to OffsetError, twice to leaf.
		require.True(t, errors.Is(errors.Unwrap(errors.Unwrap(err)), errUnimplemented))
	})
}
