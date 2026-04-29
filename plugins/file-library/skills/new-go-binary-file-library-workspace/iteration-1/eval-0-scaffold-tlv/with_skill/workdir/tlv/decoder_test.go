// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

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
			// Placeholder scenario. Once readFile is implemented, flip this
			// subtest to a happy-path assertion that compares the decoded
			// *File against the expected struct value, and add additional
			// subtests for each scenario in SPEC.md's examples.
			name:    "unimplemented stub returns wrapped errUnimplemented",
			input:   []byte{0x00},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Decode(bytes.NewReader(tc.input))
			require.ErrorIs(t, err, tc.wantErr)
			require.NotNil(t, got)
		})
	}
}
