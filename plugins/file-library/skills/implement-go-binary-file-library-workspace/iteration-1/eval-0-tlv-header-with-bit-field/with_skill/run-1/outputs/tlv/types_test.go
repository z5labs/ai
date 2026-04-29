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

func TestHeaderSize(t *testing.T) {
	t.Parallel()

	require.Equal(t, 8, binary.Size(Header{}))
}

func TestHeaderFlagsBitField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		flag Flags
		mask byte
	}{
		{"compressed", FlagCompressed, 0x01},
		{"encrypted", FlagEncrypted, 0x02},
		{"signed", FlagSigned, 0x04},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, Flags(tc.mask), tc.flag)
		})
	}

	// Mixing flags via OR should preserve all set bits.
	combined := FlagCompressed | FlagSigned
	require.True(t, combined.Has(FlagCompressed))
	require.False(t, combined.Has(FlagEncrypted))
	require.True(t, combined.Has(FlagSigned))
}

func TestMagicConstant(t *testing.T) {
	t.Parallel()

	require.Equal(t, [4]byte{'T', 'L', 'V', '1'}, Magic)
}
