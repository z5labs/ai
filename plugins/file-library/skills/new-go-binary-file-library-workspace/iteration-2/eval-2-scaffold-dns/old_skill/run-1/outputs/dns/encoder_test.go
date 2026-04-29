// Copyright (c) 2026 z5labs
//
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file or at https://opensource.org/licenses/MIT.

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
			name:    "stub returns unimplemented error",
			in:      &File{},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
		// wantErr captures the current stub state. Once Encode is
		// implemented, set this to nil and replace the assertion with
		// require.Equal(t, tc.in, got) on the decoded result.
		wantErr error
	}{
		{
			name:    "round-trip currently fails at unimplemented encode step",
			in:      &File{},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)

			got, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.in, got)
		})
	}
}
