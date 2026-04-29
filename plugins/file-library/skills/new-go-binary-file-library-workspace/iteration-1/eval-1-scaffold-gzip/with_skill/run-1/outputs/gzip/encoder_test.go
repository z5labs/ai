// Copyright (c) 2026 Z5labs and Contributors
//
// Licensed under the MIT License. See LICENSE file in the project root
// for full license information.

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
		in        *File
		wantBytes []byte
		wantErr   string
		expectErr bool
	}{
		{
			name:      "unimplemented stub returns wrapped error",
			in:        &File{},
			wantErr:   "encoding File: unimplemented",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			if tc.expectErr {
				require.Error(t, err)
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantBytes, buf.Bytes())
		})
	}
}

// TestEncodeDecodeRoundTrip demonstrates the Encode -> Decode -> require.Equal
// pattern that every encoder method should grow once the real implementation
// lands. The current expectation is that Encode fails at the unimplemented
// step, so the round-trip never reaches the comparison.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		in        *File
		expectErr bool
	}{
		{
			name:      "unimplemented stub fails before round-trip",
			in:        &File{},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			got, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.in, got)
		})
	}
}
