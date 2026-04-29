// Copyright (c) 2026 Z5labs and Contributors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gzip

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
			want: "unknown",
		},
		{
			name: "example",
			k:    KindExample,
			want: "example",
		},
		{
			name: "fallback formats unknown values",
			k:    Kind(99),
			want: "Kind(99)",
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

// fixedSizePlaceholder is a stand-in fixed-size struct so the
// binary.Size pattern is demonstrated. Replace with a real
// fixed-size gzip header struct once the spec is filled in.
type fixedSizePlaceholder struct {
	A uint8
	B uint16
	C uint32
}

func TestFixedSizePlaceholder_BinarySize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    any
		want int
	}{
		{
			name: "fixedSizePlaceholder is 7 bytes",
			v:    fixedSizePlaceholder{},
			want: 7,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := binary.Size(tc.v)
			require.Equal(t, tc.want, got)
		})
	}
}
