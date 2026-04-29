// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

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
		input   []byte
		wantErr error
	}{
		{
			name:    "stub returns unimplemented",
			input:   []byte{0x00},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(tc.input))
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
