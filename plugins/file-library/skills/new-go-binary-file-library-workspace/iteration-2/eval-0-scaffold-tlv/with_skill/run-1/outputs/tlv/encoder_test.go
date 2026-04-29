// Copyright (c) 2026 Z5labs and Contributors
// SPDX-License-Identifier: MIT

package tlv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		file      *File
		wantField string
	}{
		{
			name:      "stub returns the unimplemented error chain",
			file:      &File{Placeholder: 0x00},
			wantField: "File",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.file)
			require.Error(t, err)

			// The leaf sentinel is errUnimplemented while the package is
			// still scaffolded; replace this with the real expected sentinel
			// once writeFile is implemented.
			require.ErrorIs(t, err, errUnimplemented)

			var fieldErr *FieldError
			require.ErrorAs(t, err, &fieldErr)
			require.Equal(t, tc.wantField, fieldErr.Field)

			var offsetErr *OffsetError
			require.ErrorAs(t, err, &offsetErr)
		})
	}
}

// TestEncodeDecodeRoundTrip demonstrates the Encode -> Decode -> require.Equal
// pattern that every real encoder method should have a counterpart for. While
// the stubs return errUnimplemented, the round trip can't actually succeed;
// once both sides are real, flip the require.Error to require.NoError and
// the require.Equal will catch byte-order or length-prefix mismatches.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		file *File
	}{
		{
			name: "placeholder file round trips through Encode and Decode",
			file: &File{Placeholder: 0x42},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			encErr := Encode(&buf, tc.file)
			// Stub fails here; replace with require.NoError once writeFile
			// is implemented.
			require.ErrorIs(t, encErr, errUnimplemented)

			got, decErr := Decode(&buf)
			// Stub fails here too; replace with require.NoError + the equal
			// check below once readFile is implemented.
			require.ErrorIs(t, decErr, errUnimplemented)
			require.Nil(t, got)

			// Once both sides are real, this is the line that proves the
			// round trip:
			//
			//   require.Equal(t, tc.file, got)
		})
	}
}
