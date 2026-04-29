// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

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
			name:      "stub returns unimplemented chain at File",
			file:      &File{},
			wantField: "File",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.file)
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

// TestEncodeDecodeRoundTrip demonstrates the round-trip pattern that every
// encoder method should have once the real types and read/write logic
// land. While the stubs return errUnimplemented, this test asserts the
// expected failure shape so the implementer notices when the pipeline
// starts producing real bytes (and can then flip these expectations to
// require.NoError + require.Equal).
func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		file *File
	}{
		{name: "placeholder file", file: &File{Placeholder: 0}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			encErr := Encode(&buf, tc.file)
			require.ErrorIs(t, encErr, errUnimplemented)

			got, decErr := Decode(&buf)
			require.Nil(t, got)
			require.ErrorIs(t, decErr, errUnimplemented)

			// Once both sides are real, replace the assertions above
			// with:
			//   require.NoError(t, encErr)
			//   require.NoError(t, decErr)
			//   require.Equal(t, tc.file, got)
		})
	}
}
