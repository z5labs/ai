package tlvx

import (
	"bytes"
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

func TestFlagsHas(t *testing.T) {
	t.Parallel()

	all := FlagCompressed | FlagEncrypted | FlagSigned |
		FlagIndexed | FlagExtended | FlagStrict | FlagSealed

	// All seven defined flags set occupies bits 0..6 (0x7F).
	require.Equal(t, Flags(0x7F), all)
	require.True(t, all.Has(FlagCompressed))
	require.True(t, all.Has(FlagSealed))
	require.False(t, Flags(0).Has(FlagCompressed))
	require.False(t, FlagCompressed.Has(FlagSealed))
}

func TestHeaderRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []Header{
		{
			Magic:         Magic,
			Version:       1,
			Flags:         0,
			ChecksumAlg:   ChecksumCRC32IEEE,
			IndexCount:    0,
			ExtCount:      0,
			TrailerOffset: 16,
		},
		{
			Magic:   Magic,
			Version: 1,
			Flags: FlagCompressed | FlagEncrypted | FlagSigned |
				FlagIndexed | FlagExtended | FlagStrict | FlagSealed,
			ChecksumAlg:   ChecksumBLAKE3T32,
			IndexCount:    3,
			ExtCount:      2,
			TrailerOffset: 0x000000FF,
		},
		{
			Magic:         Magic,
			Version:       1,
			Flags:         FlagIndexed,
			ChecksumAlg:   ChecksumXXH64,
			IndexCount:    1,
			ExtCount:      0,
			TrailerOffset: 0x00000100,
		},
	}

	for i, h := range cases {
		h := h
		i := i
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, &File{Header: h}), "case %d encode", i)
			require.Equal(t, HeaderSize, buf.Len(), "case %d encoded size", i)

			f, err := Decode(&buf)
			require.NoError(t, err, "case %d decode", i)
			require.Equal(t, h, f.Header, "case %d round-trip", i)
		})
	}
}
