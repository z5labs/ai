// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

import (
	"encoding/binary"
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
