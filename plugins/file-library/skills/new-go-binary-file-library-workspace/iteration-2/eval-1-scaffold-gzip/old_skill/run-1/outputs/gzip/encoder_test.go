// Copyright (c) 2026 Z5labs and Contributors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package gzip

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
		wantErr bool
	}{
		{
			// Placeholder input. Flip wantErr to false and add a
			// `wantBytes []byte` field once writeFile is implemented.
			name:    "stub returns unimplemented error",
			in:      &File{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      *File
		wantErr bool
	}{
		{
			// Placeholder. Once Encode and Decode are implemented,
			// flip wantErr to false; the assertion below already
			// asserts the round-trip equality.
			name:    "stub fails at unimplemented encode step",
			in:      &File{},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			if tc.wantErr {
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
