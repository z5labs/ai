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

func TestFlagsBits(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		set  Flags
		bit  Flags
		want bool
	}{
		{"compressed_only_has_compressed", FlagCompressed, FlagCompressed, true},
		{"compressed_only_no_encrypted", FlagCompressed, FlagEncrypted, false},
		{"compressed_only_no_signed", FlagCompressed, FlagSigned, false},
		{"all_three_has_compressed", FlagCompressed | FlagEncrypted | FlagSigned, FlagCompressed, true},
		{"all_three_has_encrypted", FlagCompressed | FlagEncrypted | FlagSigned, FlagEncrypted, true},
		{"all_three_has_signed", FlagCompressed | FlagEncrypted | FlagSigned, FlagSigned, true},
		{"none_has_no_compressed", 0, FlagCompressed, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, tc.set.Has(tc.bit))
		})
	}
}

func TestFlagsMaskValues(t *testing.T) {
	t.Parallel()

	require.Equal(t, Flags(0x01), FlagCompressed)
	require.Equal(t, Flags(0x02), FlagEncrypted)
	require.Equal(t, Flags(0x04), FlagSigned)
}
