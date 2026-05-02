package tlvx

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// hexBytes decodes a hex string (whitespace allowed) into a byte slice. It
// fatals the test on bad input — a fixture-author bug.
func hexBytes(t *testing.T, s string) []byte {
	t.Helper()
	clean := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\n' || c == '\t' {
			continue
		}
		clean = append(clean, c)
	}
	out, err := hex.DecodeString(string(clean))
	require.NoError(t, err)
	return out
}

func TestDecodeHeader_HexLiterals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		hex  string
		want Header
	}{
		{
			name: "minimal: version 1, no flags, CRC32_IEEE, trailer at offset 16",
			hex:  "54 4C 56 58 01 00 01 00 00 00 00 00 00 00 00 10",
			want: Header{
				Magic:         Magic,
				Version:       1,
				Flags:         0,
				ChecksumAlg:   ChecksumCRC32IEEE,
				Reserved1:     0,
				IndexCount:    0,
				ExtCount:      0,
				TrailerOffset: 16,
			},
		},
		{
			name: "all seven defined flags set, BLAKE3_T32, indexed and extended",
			hex:  "54 4C 56 58 01 7F 05 00 00 03 00 02 00 00 00 FF",
			want: Header{
				Magic:   Magic,
				Version: 1,
				Flags: FlagCompressed | FlagEncrypted | FlagSigned |
					FlagIndexed | FlagExtended | FlagStrict | FlagSealed,
				ChecksumAlg:   ChecksumBLAKE3T32,
				Reserved1:     0,
				IndexCount:    3,
				ExtCount:      2,
				TrailerOffset: 0x000000FF,
			},
		},
		{
			name: "XXH64 checksum, only INDEXED flag",
			hex:  "54 4C 56 58 01 08 04 00 00 01 00 00 00 00 01 00",
			want: Header{
				Magic:         Magic,
				Version:       1,
				Flags:         FlagIndexed,
				ChecksumAlg:   ChecksumXXH64,
				Reserved1:     0,
				IndexCount:    1,
				ExtCount:      0,
				TrailerOffset: 0x00000100,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := Decode(bytes.NewReader(hexBytes(t, tc.hex)))
			require.NoError(t, err)
			require.NotNil(t, f)
			require.Equal(t, tc.want, f.Header)
		})
	}
}

func TestDecodeHeader_FlagHelpers(t *testing.T) {
	t.Parallel()

	f, err := Decode(bytes.NewReader(hexBytes(t,
		"54 4C 56 58 01 7F 05 00 00 03 00 02 00 00 00 FF")))
	require.NoError(t, err)

	require.True(t, f.Header.Flags.Has(FlagCompressed))
	require.True(t, f.Header.Flags.Has(FlagEncrypted))
	require.True(t, f.Header.Flags.Has(FlagSigned))
	require.True(t, f.Header.Flags.Has(FlagIndexed))
	require.True(t, f.Header.Flags.Has(FlagExtended))
	require.True(t, f.Header.Flags.Has(FlagStrict))
	require.True(t, f.Header.Flags.Has(FlagSealed))
}

func TestDecodeHeader_MagicMismatch_FieldErrorOffsetErrorLeaf(t *testing.T) {
	t.Parallel()

	// "XXXX" instead of "TLVX" — first four bytes fail the magic check.
	bad := hexBytes(t, "58 58 58 58 01 00 01 00 00 00 00 00 00 00 00 10")

	_, err := Decode(bytes.NewReader(bad))
	require.Error(t, err)

	// Leaf assertion: ErrMagicMismatch sits at the bottom of the chain.
	require.ErrorIs(t, err, ErrMagicMismatch)

	// Outer wrapper: FieldError naming the field that failed.
	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Magic", fe.Field)

	// Middle wrapper: OffsetError sitting between FieldError and the leaf.
	// fe.Err must be exactly an *OffsetError, not the leaf, to prove the
	// chain order is FieldError → OffsetError → leaf.
	oe, ok := fe.Err.(*OffsetError)
	require.True(t, ok, "FieldError.Err must be *OffsetError, got %T", fe.Err)
	// OffsetError.Offset is the start of the failing field; Magic begins
	// at byte 0.
	require.Equal(t, int64(0), oe.Offset)

	// And the leaf error sits inside the OffsetError.
	require.ErrorIs(t, oe.Err, ErrMagicMismatch)
}

func TestDecodeHeader_TruncatedInput_WrappedError(t *testing.T) {
	t.Parallel()

	// Only three bytes — magic read fails partway.
	_, err := Decode(bytes.NewReader([]byte{'T', 'L', 'V'}))
	require.Error(t, err)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Magic", fe.Field)

	// Underlying I/O error from io.ReadFull on short input.
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}
