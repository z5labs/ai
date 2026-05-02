package tlvx

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeHeaderHappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   File
		want []byte
	}{
		{
			name: "minimal_no_flags",
			in: File{
				Header: Header{
					Magic:         MagicTLVX,
					Version:       1,
					Flags:         0,
					ChecksumAlg:   ChecksumAlgCRC32IEEE,
					Reserved1:     0,
					IndexCount:    0,
					ExtCount:      0,
					TrailerOffset: 16,
				},
			},
			want: []byte{
				// Magic: "TLVX"
				0x54, 0x4C, 0x56, 0x58,
				// Version=1, Flags=0, ChecksumAlg=0x01, Reserved1=0
				0x01, 0x00, 0x01, 0x00,
				// IndexCount=0
				0x00, 0x00,
				// ExtCount=0
				0x00, 0x00,
				// TrailerOffset=16
				0x00, 0x00, 0x00, 0x10,
			},
		},
		{
			name: "with_flags_compressed_indexed_sealed",
			in: File{
				Header: Header{
					Magic:         MagicTLVX,
					Version:       1,
					Flags:         FlagCompressed | FlagIndexed | FlagSealed,
					ChecksumAlg:   ChecksumAlgSHA256T32,
					Reserved1:     0,
					IndexCount:    2,
					ExtCount:      0,
					TrailerOffset: 32,
				},
			},
			want: []byte{
				0x54, 0x4C, 0x56, 0x58,
				0x01, 0x49, 0x03, 0x00,
				0x00, 0x02,
				0x00, 0x00,
				0x00, 0x00, 0x00, 0x20,
			},
		},
		{
			name: "all_seven_flags_extended",
			in: File{
				Header: Header{
					Magic: MagicTLVX,
					Version: 1,
					Flags: FlagCompressed | FlagEncrypted | FlagSigned |
						FlagIndexed | FlagExtended | FlagStrict | FlagSealed,
					ChecksumAlg:   ChecksumAlgBLAKE3T32,
					Reserved1:     0,
					IndexCount:    1,
					ExtCount:      3,
					TrailerOffset: 4096,
				},
			},
			want: []byte{
				0x54, 0x4C, 0x56, 0x58,
				0x01, 0x7F, 0x05, 0x00,
				0x00, 0x01,
				0x00, 0x03,
				0x00, 0x00, 0x10, 0x00,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, &tc.in))
			require.Equal(t, tc.want, buf.Bytes())
		})
	}
}

func TestEncodeHeaderRejectsReservedFlagBit(t *testing.T) {
	t.Parallel()

	in := File{
		Header: Header{
			Magic:       MagicTLVX,
			Version:     1,
			Flags:       0x80, // reserved bit 7 set
			ChecksumAlg: ChecksumAlgCRC32IEEE,
		},
	}

	var buf bytes.Buffer
	err := Encode(&buf, &in)
	require.Error(t, err)

	var rfb *ReservedFlagBitError
	require.ErrorAs(t, err, &rfb)
	require.Equal(t, "Header.Flags", rfb.Field)
	require.Equal(t, uint8(7), rfb.Bit)
}

func TestEncodeHeaderRejectsNonZeroReserved1(t *testing.T) {
	t.Parallel()

	in := File{
		Header: Header{
			Magic:       MagicTLVX,
			Version:     1,
			ChecksumAlg: ChecksumAlgCRC32IEEE,
			Reserved1:   0x99,
		},
	}

	var buf bytes.Buffer
	err := Encode(&buf, &in)
	require.Error(t, err)

	var rfe *ReservedFieldNonZeroError
	require.ErrorAs(t, err, &rfe)
	require.Equal(t, "Header.Reserved1", rfe.Field)
	require.Equal(t, uint8(0x99), rfe.Got)
}

func TestEncodeHeaderRejectsUnknownChecksumAlg(t *testing.T) {
	t.Parallel()

	in := File{
		Header: Header{
			Magic:       MagicTLVX,
			Version:     1,
			ChecksumAlg: 0xEE,
		},
	}

	var buf bytes.Buffer
	err := Encode(&buf, &in)
	require.Error(t, err)

	var uca *UnknownChecksumAlgError
	require.ErrorAs(t, err, &uca)
	require.Equal(t, ChecksumAlg(0xEE), uca.Alg)
}

func TestRoundTripHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   File
	}{
		{
			name: "minimal",
			in: File{Header: Header{
				Magic:         MagicTLVX,
				Version:       1,
				ChecksumAlg:   ChecksumAlgCRC32IEEE,
				TrailerOffset: 16,
			}},
		},
		{
			name: "all_flags",
			in: File{Header: Header{
				Magic: MagicTLVX,
				Version: 1,
				Flags: FlagCompressed | FlagEncrypted | FlagSigned |
					FlagIndexed | FlagExtended | FlagStrict | FlagSealed,
				ChecksumAlg:   ChecksumAlgBLAKE3T32,
				IndexCount:    1,
				ExtCount:      3,
				TrailerOffset: 4096,
			}},
		},
		{
			name: "every_checksum_alg",
			in: File{Header: Header{
				Magic:         MagicTLVX,
				Version:       1,
				ChecksumAlg:   ChecksumAlgXXH64,
				IndexCount:    65535,
				ExtCount:      65535,
				TrailerOffset: 0xFFFFFFFF,
			}},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, &tc.in))

			out, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, &tc.in, out)
		})
	}
}
