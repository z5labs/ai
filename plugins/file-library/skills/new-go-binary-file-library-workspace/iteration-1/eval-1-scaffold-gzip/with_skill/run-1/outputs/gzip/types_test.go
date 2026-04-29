// Copyright (c) 2026 Z5labs and Contributors
//
// Licensed under the MIT License. See LICENSE file in the project root
// for full license information.

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
		{name: "unknown", k: KindUnknown, want: "Unknown"},
		{name: "example", k: KindExample, want: "Example"},
		{name: "fallback formats unknown numeric kind", k: Kind(255), want: "Kind(255)"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.k.String())
		})
	}
}

func TestFileBinarySize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    any
		want int
	}{
		{
			name: "placeholder File is fixed size",
			v:    File{},
			// One uint8 placeholder field. Update once the real fixed-size
			// header is defined in types.go.
			want: 1,
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
