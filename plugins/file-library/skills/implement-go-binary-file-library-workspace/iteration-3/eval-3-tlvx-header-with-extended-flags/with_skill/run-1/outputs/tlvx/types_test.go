package tlvx

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
	require.Equal(t, 16, binary.Size(Header{}))
}

func TestFlagsBitFieldMasks(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		mask Flags
		want byte
	}{
		{"compressed", FlagCompressed, 0x01},
		{"encrypted", FlagEncrypted, 0x02},
		{"signed", FlagSigned, 0x04},
		{"indexed", FlagIndexed, 0x08},
		{"extended", FlagExtended, 0x10},
		{"strict", FlagStrict, 0x20},
		{"sealed", FlagSealed, 0x40},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, byte(tc.want), byte(tc.mask))
		})
	}
}

func TestFlagsCombineAndQuery(t *testing.T) {
	t.Parallel()

	f := Flags(0)
	f |= FlagCompressed
	f |= FlagIndexed
	f |= FlagSealed

	require.True(t, f&FlagCompressed != 0)
	require.True(t, f&FlagIndexed != 0)
	require.True(t, f&FlagSealed != 0)
	require.True(t, f&FlagEncrypted == 0)
	require.True(t, f&FlagSigned == 0)
	require.True(t, f&FlagExtended == 0)
	require.True(t, f&FlagStrict == 0)
}

func TestChecksumAlgString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    ChecksumAlg
		want string
	}{
		{"crc32_ieee", ChecksumAlgCRC32IEEE, "CRC32_IEEE"},
		{"crc64_ecma", ChecksumAlgCRC64ECMA, "CRC64_ECMA"},
		{"sha256_t32", ChecksumAlgSHA256T32, "SHA256_T32"},
		{"xxh64", ChecksumAlgXXH64, "XXH64"},
		{"blake3_t32", ChecksumAlgBLAKE3T32, "BLAKE3_T32"},
		{"unknown", ChecksumAlg(0xff), "ChecksumAlg(0xff)"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, tc.v.String())
		})
	}
}

func TestUnexpectedMagicError(t *testing.T) {
	t.Parallel()

	got := [4]byte{'T', 'L', 'V', '1'}
	err := &UnexpectedMagicError{Got: got}
	require.Contains(t, err.Error(), "magic")
}
