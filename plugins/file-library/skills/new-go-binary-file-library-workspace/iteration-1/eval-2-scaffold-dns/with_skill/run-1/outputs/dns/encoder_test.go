// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      *File
		wantErr error
	}{
		{
			name:    "stub returns unimplemented",
			in:      &File{},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// TestEncodeDecodeRoundTrip demonstrates the Encode -> Decode -> compare
// pattern. Every encoder method should grow a round-trip subtest like this
// once the real implementation lands; for now the test asserts the
// unimplemented error path so the scaffold builds and runs cleanly.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      *File
		wantErr error
	}{
		{
			name:    "round-trip stops at unimplemented encoder",
			in:      &File{},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			require.ErrorIs(t, err, tc.wantErr)

			// Once Encode is real, replace the assertion above with
			// require.NoError(t, err) and the rest of the test exercises
			// the round-trip:
			//
			//   got, err := Decode(&buf)
			//   require.NoError(t, err)
			//   require.Equal(t, tc.in, got)
			_, decErr := Decode(&buf)
			require.ErrorIs(t, decErr, tc.wantErr)
		})
	}
}
