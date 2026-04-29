package tlv

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorChain(t *testing.T) {
	t.Parallel()

	err := &FieldError{Field: "File", Err: &OffsetError{Offset: 0, Err: errUnimplemented}}

	require.ErrorIs(t, err, errUnimplemented)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "File", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(0), oe.Offset)

	require.True(t, errors.Is(err, errUnimplemented))
}

// TestChecksumMismatchSentinel verifies that ErrChecksumMismatch is exported
// and behaves as a sentinel that can be wrapped through the FieldError /
// OffsetError chain.
func TestChecksumMismatchSentinel(t *testing.T) {
	t.Parallel()

	wrapped := &FieldError{
		Field: "Trailer.CRC32",
		Err:   &OffsetError{Offset: 12, Err: ErrChecksumMismatch},
	}

	require.ErrorIs(t, wrapped, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, wrapped, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, wrapped, &oe)
	require.Equal(t, int64(12), oe.Offset)
}
