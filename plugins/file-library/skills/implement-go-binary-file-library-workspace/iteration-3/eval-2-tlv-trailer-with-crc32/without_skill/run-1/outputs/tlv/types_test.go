package tlv

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorChain(t *testing.T) {
	t.Parallel()

	err := &FieldError{Field: "Trailer.CRC32", Err: &OffsetError{Offset: 12, Err: ErrChecksumMismatch}}

	require.ErrorIs(t, err, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(12), oe.Offset)

	require.True(t, errors.Is(err, ErrChecksumMismatch))
}
