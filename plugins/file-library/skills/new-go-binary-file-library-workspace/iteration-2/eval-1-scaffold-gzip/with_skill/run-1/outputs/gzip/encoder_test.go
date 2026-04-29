// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package gzip

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     *File
		wantField string
	}{
		{
			name:      "unimplemented stub returns the full error chain",
			input:     &File{Placeholder: 0x00},
			wantField: "File",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.input)
			require.Error(t, err)

			// Leaf sentinel reachable via errors.Is.
			require.ErrorIs(t, err, errUnimplemented)

			// Field path reachable via errors.As.
			var fieldErr *FieldError
			require.ErrorAs(t, err, &fieldErr)
			require.Equal(t, tc.wantField, fieldErr.Field)

			// Byte offset reachable via errors.As.
			var offsetErr *OffsetError
			require.ErrorAs(t, err, &offsetErr)
		})
	}
}

// TestEncodeDecodeRoundTrip is the canonical end-to-end correctness check:
// Encode -> Decode -> require.Equal against the original. It is expected to
// fail at the unimplemented stub today; once both sides are real this becomes
// the cheapest sanity test in the package.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		want *File
	}{
		{
			name: "placeholder file",
			want: &File{Placeholder: 0x00},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.want)
			if err != nil {
				// Expected today: writeFile returns errUnimplemented through the
				// FieldError -> OffsetError chain. Once the encoder is real, drop
				// this branch and fall through to the Decode/Equal check below.
				require.ErrorIs(t, err, errUnimplemented)
				return
			}

			got, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
