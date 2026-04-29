// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

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
		wantField string
	}{
		{
			name: "unimplemented stub returns the full error chain",
			// Single zero byte — decoder is a stub, the bytes don't matter yet.
			input:     []byte{0x00},
			wantField: "File",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(tc.input))
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
