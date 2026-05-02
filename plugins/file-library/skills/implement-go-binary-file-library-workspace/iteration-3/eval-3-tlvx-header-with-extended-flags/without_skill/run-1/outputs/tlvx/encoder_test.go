package tlvx

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeHeader_HexLiterals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		header  Header
		wantHex string
	}{
		{
			name: "minimal header",
			header: Header{
				Magic:         Magic,
				Version:       1,
				Flags:         0,
				ChecksumAlg:   ChecksumCRC32IEEE,
				IndexCount:    0,
				ExtCount:      0,
				TrailerOffset: 16,
			},
			wantHex: "54 4C 56 58 01 00 01 00 00 00 00 00 00 00 00 10",
		},
		{
			name: "all seven defined flags set, BLAKE3_T32",
			header: Header{
				Magic:   Magic,
				Version: 1,
				Flags: FlagCompressed | FlagEncrypted | FlagSigned |
					FlagIndexed | FlagExtended | FlagStrict | FlagSealed,
				ChecksumAlg:   ChecksumBLAKE3T32,
				IndexCount:    3,
				ExtCount:      2,
				TrailerOffset: 0x000000FF,
			},
			wantHex: "54 4C 56 58 01 7F 05 00 00 03 00 02 00 00 00 FF",
		},
		{
			name: "XXH64 + only INDEXED",
			header: Header{
				Magic:         Magic,
				Version:       1,
				Flags:         FlagIndexed,
				ChecksumAlg:   ChecksumXXH64,
				IndexCount:    1,
				ExtCount:      0,
				TrailerOffset: 0x00000100,
			},
			wantHex: "54 4C 56 58 01 08 04 00 00 01 00 00 00 00 01 00",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, &File{Header: tc.header})
			require.NoError(t, err)
			require.Equal(t, hexBytes(t, tc.wantHex), buf.Bytes())
			require.Equal(t, HeaderSize, buf.Len())
		})
	}
}

func TestEncodeHeader_ZeroMagicSubstitutesCanonical(t *testing.T) {
	t.Parallel()

	// Caller leaves Magic and Version as zero values; encoder substitutes
	// the canonical "TLVX"/version-1 bytes per spec.
	var buf bytes.Buffer
	err := Encode(&buf, &File{Header: Header{ChecksumAlg: ChecksumCRC32IEEE}})
	require.NoError(t, err)
	got := buf.Bytes()
	require.Equal(t, byte('T'), got[0])
	require.Equal(t, byte('L'), got[1])
	require.Equal(t, byte('V'), got[2])
	require.Equal(t, byte('X'), got[3])
	require.Equal(t, byte(1), got[4])
}
