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

func TestDecode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     []byte
		wantErr   string
		expectErr bool
	}{
		{
			name: "unimplemented stub returns wrapped error",
			// A single placeholder byte. Replace with real spec example
			// inputs once readFile() is implemented.
			input:     []byte{0x00},
			wantErr:   "decoding File: unimplemented",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(tc.input))
			if tc.expectErr {
				require.Error(t, err)
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
