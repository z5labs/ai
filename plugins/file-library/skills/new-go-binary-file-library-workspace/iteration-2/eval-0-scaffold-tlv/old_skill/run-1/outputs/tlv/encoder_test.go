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
		name    string
		input   *File
		wantErr error
	}{
		{
			// Placeholder scenario. Once writeFile is implemented, flip this
			// subtest to a happy-path assertion that compares the buffer's
			// bytes against an expected hex-byte literal, and add additional
			// subtests for each scenario in SPEC.md's examples.
			name:    "unimplemented stub returns wrapped errUnimplemented",
			input:   &File{},
			wantErr: errUnimplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.input)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
	}{
		{
			// Round-trip is the cheapest end-to-end correctness check
			// available; every encoder method should have one. Until
			// writeFile is implemented this test asserts the unimplemented
			// failure path so it can be flipped to require.NoError + a
			// require.Equal of the round-tripped struct once the encoder is
			// real.
			name: "round-trip currently fails at the unimplemented encode step",
			in:   &File{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			require.ErrorIs(t, err, errUnimplemented)

			// Once Encode succeeds, replace the assertion above with
			// require.NoError(t, err) and uncomment the lines below.
			//
			// got, err := Decode(&buf)
			// require.NoError(t, err)
			// require.Equal(t, tc.in, got)
			_ = tc.in
		})
	}
}
