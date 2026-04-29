package tlv

import (
	"encoding/binary"
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

func TestTrailerSize(t *testing.T) {
	t.Parallel()

	// Trailer is fixed at 4 bytes per the SPEC.md "Trailer" field definition.
	require.Equal(t, 4, binary.Size(Trailer{}))
}

func TestErrChecksumMismatchChain(t *testing.T) {
	t.Parallel()

	// A typed error wrapping ErrChecksumMismatch via wrapErr-style chain.
	err := &FieldError{Field: "Trailer.CRC32", Err: &OffsetError{Offset: 8, Err: ErrChecksumMismatch}}

	require.ErrorIs(t, err, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(8), oe.Offset)
}
