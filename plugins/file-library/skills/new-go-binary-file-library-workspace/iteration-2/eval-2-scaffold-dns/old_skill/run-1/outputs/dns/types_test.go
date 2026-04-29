// Copyright (c) 2026 z5labs
//
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file or at https://opensource.org/licenses/MIT.

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
		in   Kind
		want string
	}{
		{
			name: "zero value renders as Unknown",
			in:   KindUnknown,
			want: "Unknown",
		},
		{
			name: "example value renders by name",
			in:   KindExample,
			want: "Example",
		},
		{
			name: "unknown value renders with numeric form",
			in:   Kind(99),
			want: "Kind(99)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.in.String())
		})
	}
}

// fixedSizeFile is a placeholder fixed-size struct used to demonstrate the
// binary.Size() check pattern. Replace with the real fixed-size DNS structure
// (e.g., Header is 12 bytes) once the spec is implemented.
type fixedSizeFile struct {
	A uint16
	B uint16
	C uint32
}

func TestFileBinarySize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   any
		want int
	}{
		{
			name: "fixed-size placeholder struct is 8 bytes",
			in:   fixedSizeFile{},
			want: 8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, binary.Size(tc.in))
		})
	}
}
