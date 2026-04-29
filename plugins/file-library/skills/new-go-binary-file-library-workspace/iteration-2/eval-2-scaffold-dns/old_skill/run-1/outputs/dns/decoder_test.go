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

func TestDecode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      []byte
		wantErr error
	}{
		{
			// Placeholder input: a single zero byte. The implementer will
			// replace this with a real DNS message hex literal once
			// readFile() is implemented, and flip wantErr to nil.
			name:    "stub returns unimplemented error",
			in:      []byte{0x00},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Decode(bytes.NewReader(tc.in))
			require.ErrorIs(t, err, tc.wantErr)
			// The stub returns a zero *File alongside the error. Once
			// readFile() is real, replace this with a require.Equal on
			// the expected decoded value.
			require.NotNil(t, got)
		})
	}
}
