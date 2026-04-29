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
		name      string
		input     []byte
		wantField string
	}{
		{
			name:      "stub returns unimplemented chain at File",
			input:     []byte{0x00},
			wantField: "File",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Decode(bytes.NewReader(tc.input))
			require.Nil(t, got)
			require.Error(t, err)

			require.ErrorIs(t, err, errUnimplemented)

			var fieldErr *FieldError
			require.ErrorAs(t, err, &fieldErr)
			require.Equal(t, tc.wantField, fieldErr.Field)

			var offsetErr *OffsetError
			require.ErrorAs(t, err, &offsetErr)
		})
	}
}
