// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKindString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		k    Kind
		want string
	}{
		{
			name: "unknown",
			k:    KindUnknown,
			want: "Unknown",
		},
		{
			name: "a record",
			k:    KindA,
			want: "A",
		},
		{
			name: "out of range falls back to numeric form",
			k:    Kind(255),
			want: "Kind(255)",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.k.String())
		})
	}
}

// TestFileBinarySize exercises binary.Size against the scaffold's File struct
// to demonstrate the fixed-size-check pattern. Once File is replaced with the
// real DNS message layout, the expected value should be updated to match the
// spec.
func TestFileBinarySize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    any
		want int
	}{
		{
			name: "placeholder file is two bytes",
			v:    File{},
			want: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, binary.Size(tc.v))
		})
	}
}
