// Copyright (c) 2026 Z5labs and Contributors
//
// SPDX-License-Identifier: MIT

package dns

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKindString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		kind Kind
		want string
	}{
		{name: "unknown", kind: KindUnknown, want: "Unknown"},
		{name: "a", kind: KindA, want: "A"},
		{name: "ns", kind: KindNS, want: "NS"},
		{name: "out of range", kind: Kind(255), want: "Kind(255)"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.kind.String())
		})
	}
}

// fixedSizeProbe is a placeholder fixed-size struct used to exercise the
// binary.Size pattern. Replace with the real DNS Header (12 bytes per
// RFC 1035 section 4.1.1) once the types are filled in.
type fixedSizeProbe struct {
	A uint16
	B uint16
	C uint32
}

func TestFixedSizeProbeSize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    any
		want int
	}{
		{name: "fixedSizeProbe is 8 bytes", v: fixedSizeProbe{}, want: 8},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, binary.Size(tc.v))
		})
	}
}

func TestErrorChain(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "field wraps offset wraps leaf",
			err: &FieldError{
				Field: "Header",
				Err: &OffsetError{
					Offset: 4,
					Err:    errUnimplemented,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorIs(t, tc.err, errUnimplemented)

			var fe *FieldError
			require.True(t, errors.As(tc.err, &fe))
			require.Equal(t, "Header", fe.Field)

			var oe *OffsetError
			require.True(t, errors.As(tc.err, &oe))
			require.Equal(t, int64(4), oe.Offset)
		})
	}
}
