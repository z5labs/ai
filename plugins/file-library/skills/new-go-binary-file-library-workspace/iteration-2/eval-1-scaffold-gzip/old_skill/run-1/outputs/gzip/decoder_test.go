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

func TestDecode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			// Placeholder input. Flip wantErr to false and add a
			// `wantFile *File` field once readFile is implemented.
			name:    "stub returns unimplemented error",
			input:   []byte{0x00},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Decode(bytes.NewReader(tc.input))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
