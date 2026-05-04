package dsf

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
		v    Kind
		want string
	}{
		{"unknown", KindUnknown, "Unknown"},
		{"example", KindExample, "Example"},
		{"unrecognised", Kind(0xff), "Kind(0xff)"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, tc.v.String())
		})
	}
}

func TestFileHeaderSize(t *testing.T) {
	t.Parallel()
	require.Equal(t, 12, binary.Size(FileHeader{}))
}

func TestMagicCookieValue(t *testing.T) {
	t.Parallel()
	require.Equal(t, [8]byte{'X', 'P', 'L', 'N', 'E', 'D', 'S', 'F'}, MagicCookie)
}

func TestErrorChain(t *testing.T) {
	t.Parallel()

	err := &FieldError{Field: "Header", Err: &OffsetError{Offset: 4, Err: errUnimplemented}}

	require.ErrorIs(t, err, errUnimplemented)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(4), oe.Offset)

	require.True(t, errors.Is(err, errUnimplemented))
}
