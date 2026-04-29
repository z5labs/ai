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
		name      string
		input     []byte
		wantField string
	}{
		{
			name:      "stub returns the unimplemented error chain",
			input:     []byte{0x00},
			wantField: "File",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(tc.input))
			require.Error(t, err)

			// The leaf sentinel is errUnimplemented while the package is
			// still scaffolded; replace this with the real expected sentinel
			// (e.g. io.ErrUnexpectedEOF or ErrInvalid) once readFile is
			// implemented.
			require.ErrorIs(t, err, errUnimplemented)

			var fieldErr *FieldError
			require.ErrorAs(t, err, &fieldErr)
			require.Equal(t, tc.wantField, fieldErr.Field)

			var offsetErr *OffsetError
			require.ErrorAs(t, err, &offsetErr)
		})
	}
}
