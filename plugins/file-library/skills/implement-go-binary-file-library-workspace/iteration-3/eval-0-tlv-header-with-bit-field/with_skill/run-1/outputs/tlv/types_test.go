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

func TestFlagsBitConstants(t *testing.T) {
	t.Parallel()

	require.Equal(t, Flags(0x01), FlagCompressed)
	require.Equal(t, Flags(0x02), FlagEncrypted)
	require.Equal(t, Flags(0x04), FlagSigned)
	require.Equal(t, Flags(0xF8), flagsReservedMask)
}

func TestFlagsBitFieldOps(t *testing.T) {
	t.Parallel()

	var f Flags
	f |= FlagCompressed
	f |= FlagSigned
	require.True(t, f&FlagCompressed != 0)
	require.True(t, f&FlagSigned != 0)
	require.True(t, f&FlagEncrypted == 0)
	require.True(t, f&flagsReservedMask == 0)
}

func TestFlagsString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    Flags
		want string
	}{
		{"none", 0, "NONE"},
		{"compressed", FlagCompressed, "COMPRESSED"},
		{"encrypted", FlagEncrypted, "ENCRYPTED"},
		{"signed", FlagSigned, "SIGNED"},
		{"compressed_encrypted", FlagCompressed | FlagEncrypted, "COMPRESSED|ENCRYPTED"},
		{"all_three", FlagCompressed | FlagEncrypted | FlagSigned, "COMPRESSED|ENCRYPTED|SIGNED"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, tc.v.String())
		})
	}
}

func TestMagicBytes(t *testing.T) {
	t.Parallel()

	require.Equal(t, [4]byte{0x54, 0x4C, 0x56, 0x31}, Magic)
}

func TestFileHoldsHeader(t *testing.T) {
	t.Parallel()

	f := &File{Header: Header{Magic: Magic, Version: 1}}
	require.Equal(t, uint8(1), f.Header.Version)
	require.Equal(t, Magic, f.Header.Magic)
}
