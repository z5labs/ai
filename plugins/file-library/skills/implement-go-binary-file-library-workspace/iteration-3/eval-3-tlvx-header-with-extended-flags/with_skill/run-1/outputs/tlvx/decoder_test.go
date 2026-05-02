package tlvx

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeHeaderHappyPath(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input []byte
		want  Header
	}{
		{
			name: "minimal_no_flags",
			input: []byte{
				// Magic: "TLVX"
				0x54, 0x4C, 0x56, 0x58,
				// Version=1, Flags=0x00, ChecksumAlg=0x01 (CRC32_IEEE), Reserved1=0
				0x01, 0x00, 0x01, 0x00,
				// IndexCount=0
				0x00, 0x00,
				// ExtCount=0
				0x00, 0x00,
				// TrailerOffset=16 (header-only, no body)
				0x00, 0x00, 0x00, 0x10,
			},
			want: Header{
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
		{
			name: "with_flags_compressed_indexed_sealed",
			input: []byte{
				// Magic: "TLVX"
				0x54, 0x4C, 0x56, 0x58,
				// Version=1, Flags=0x49 (COMPRESSED|INDEXED|SEALED), ChecksumAlg=0x03 (SHA256_T32), Reserved1=0
				0x01, 0x49, 0x03, 0x00,
				// IndexCount=2
				0x00, 0x02,
				// ExtCount=0
				0x00, 0x00,
				// TrailerOffset=0x00000020 (32)
				0x00, 0x00, 0x00, 0x20,
			},
			want: Header{
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
		{
			name: "all_seven_flags_extended",
			input: []byte{
				// Magic: "TLVX"
				0x54, 0x4C, 0x56, 0x58,
				// Version=1, Flags=0x7F (all 7 defined flags), ChecksumAlg=0x05 (BLAKE3_T32), Reserved1=0
				0x01, 0x7F, 0x05, 0x00,
				// IndexCount=1
				0x00, 0x01,
				// ExtCount=3
				0x00, 0x03,
				// TrailerOffset=0x00001000 (4096)
				0x00, 0x00, 0x10, 0x00,
			},
			want: Header{
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
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := Decode(bytes.NewReader(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.want, f.Header)
		})
	}
}

func TestDecodeHeaderMagicMismatch(t *testing.T) {
	t.Parallel()

	// Magic: "TLV1" instead of "TLVX" — the rest of the header is well-formed
	// but should never be inspected: the decoder must fail at the magic check.
	input := []byte{
		0x54, 0x4C, 0x56, 0x31, // Magic = "TLV1" (mismatch)
		0x01, 0x00, 0x01, 0x00, // Version, Flags, ChecksumAlg, Reserved1
		0x00, 0x00, 0x00, 0x00, // IndexCount, ExtCount
		0x00, 0x00, 0x00, 0x10, // TrailerOffset
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)

	// FieldError → OffsetError → UnexpectedMagicError leaf chain.
	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Magic", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// Magic is the first 4 bytes; the read consumes 4 bytes before the check fails,
	// so the offset of the failure is 4.
	require.Equal(t, int64(4), oe.Offset)

	var ume *UnexpectedMagicError
	require.ErrorAs(t, err, &ume)
	require.Equal(t, [4]byte{'T', 'L', 'V', '1'}, ume.Got)
}

func TestDecodeHeaderTruncated(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     []byte
		wantField string
	}{
		{
			name:      "truncated_in_magic",
			input:     []byte{0x54, 0x4C}, // 2 bytes — short of the 4-byte magic
			wantField: "Header.Magic",
		},
		{
			name: "truncated_after_magic",
			// Magic complete, then only 4 bytes of the remaining 12 (e.g., the
			// rest of the body never arrives). Surfaces under "Header".
			input: []byte{
				0x54, 0x4C, 0x56, 0x58, // Magic = "TLVX"
				0x01, 0x00, 0x01, 0x00, // partial body (Version, Flags, ChecksumAlg, Reserved1)
			},
			wantField: "Header",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(tc.input))
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)

			var fe *FieldError
			require.ErrorAs(t, err, &fe)
			require.Equal(t, tc.wantField, fe.Field)
		})
	}
}

func TestDecodeHeaderUnknownVersion(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x58, // Magic = "TLVX"
		0x02,                   // Version=2 (unsupported)
		0x00, 0x01, 0x00,       // Flags, ChecksumAlg, Reserved1
		0x00, 0x00, 0x00, 0x00, // IndexCount, ExtCount
		0x00, 0x00, 0x00, 0x10, // TrailerOffset
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)

	var uve *UnknownVersionError
	require.ErrorAs(t, err, &uve)
	require.Equal(t, uint8(2), uve.Version)
}

func TestDecodeHeaderReserved1NonZero(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x58, // Magic = "TLVX"
		0x01,                   // Version=1
		0x00, 0x01, 0xAB,       // Flags, ChecksumAlg, Reserved1=0xAB (non-zero!)
		0x00, 0x00, 0x00, 0x00, // IndexCount, ExtCount
		0x00, 0x00, 0x00, 0x10, // TrailerOffset
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)

	var rfe *ReservedFieldNonZeroError
	require.ErrorAs(t, err, &rfe)
	require.Equal(t, "Header.Reserved1", rfe.Field)
	require.Equal(t, uint8(0xAB), rfe.Got)
}

func TestDecodeHeaderReservedFlagBitSet(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x58, // Magic = "TLVX"
		0x01,                   // Version=1
		0x80, 0x01, 0x00,       // Flags=0x80 (reserved bit 7 set), ChecksumAlg, Reserved1
		0x00, 0x00, 0x00, 0x00, // IndexCount, ExtCount
		0x00, 0x00, 0x00, 0x10, // TrailerOffset
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)

	var rfb *ReservedFlagBitError
	require.ErrorAs(t, err, &rfb)
	require.Equal(t, "Header.Flags", rfb.Field)
	require.Equal(t, uint8(7), rfb.Bit)
}

func TestDecodeHeaderUnknownChecksumAlg(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x58, // Magic = "TLVX"
		0x01,                   // Version=1
		0x00, 0xEE, 0x00,       // Flags, ChecksumAlg=0xEE (unknown), Reserved1
		0x00, 0x00, 0x00, 0x00, // IndexCount, ExtCount
		0x00, 0x00, 0x00, 0x10, // TrailerOffset
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)

	var uca *UnknownChecksumAlgError
	require.ErrorAs(t, err, &uca)
	require.Equal(t, ChecksumAlg(0xEE), uca.Alg)
}
