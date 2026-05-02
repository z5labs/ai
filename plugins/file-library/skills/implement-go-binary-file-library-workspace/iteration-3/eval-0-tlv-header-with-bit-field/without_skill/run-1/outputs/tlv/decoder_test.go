package tlv

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// mustDecodeHex strips whitespace from a hex literal and decodes it.
func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	clean := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '\n' || c == '\t' {
			continue
		}
		clean = append(clean, c)
	}
	b, err := hex.DecodeString(string(clean))
	require.NoError(t, err)
	return b
}

func TestDecodeHeader(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		hex  string
		want Header
	}{
		{
			name: "no flags",
			hex:  "54 4C 56 31 01 00 00 00",
			want: Header{Magic: Magic, Version: 1, Flags: 0, Reserved: 0},
		},
		{
			name: "compressed flag",
			hex:  "54 4C 56 31 01 01 00 00",
			want: Header{Magic: Magic, Version: 1, Flags: FlagCompressed, Reserved: 0},
		},
		{
			name: "encrypted flag",
			hex:  "54 4C 56 31 01 02 00 00",
			want: Header{Magic: Magic, Version: 1, Flags: FlagEncrypted, Reserved: 0},
		},
		{
			name: "signed flag",
			hex:  "54 4C 56 31 01 04 00 00",
			want: Header{Magic: Magic, Version: 1, Flags: FlagSigned, Reserved: 0},
		},
		{
			name: "all defined flags",
			hex:  "54 4C 56 31 01 07 00 00",
			want: Header{
				Magic:    Magic,
				Version:  1,
				Flags:    FlagCompressed | FlagEncrypted | FlagSigned,
				Reserved: 0,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := Decode(bytes.NewReader(mustDecodeHex(t, tc.hex)))
			require.NoError(t, err)
			require.NotNil(t, f)
			require.Equal(t, tc.want, f.Header)

			require.True(t, f.Header.Flags.Has(tc.want.Flags))
		})
	}
}

func TestDecodeHeaderErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		hex       string
		wantField string
		wantOff   int64
		wantLeaf  error
	}{
		{
			name:      "bad magic",
			hex:       "00 00 00 00 01 00 00 00",
			wantField: "Header.Magic",
			wantOff:   4, // failure detected after reading 4 bytes
			wantLeaf:  ErrBadMagic,
		},
		{
			name:      "unsupported version",
			hex:       "54 4C 56 31 02 00 00 00",
			wantField: "Header.Version",
			wantOff:   5,
			wantLeaf:  ErrUnsupportedVersion,
		},
		{
			name:      "reserved flag bits set",
			hex:       "54 4C 56 31 01 80 00 00",
			wantField: "Header.Flags",
			wantOff:   6,
			wantLeaf:  ErrReservedFlagBitsSet,
		},
		{
			name:      "reserved non-zero",
			hex:       "54 4C 56 31 01 00 00 01",
			wantField: "Header.Reserved",
			wantOff:   8,
			wantLeaf:  ErrReservedNonZero,
		},
		{
			name:      "short read on magic",
			hex:       "54 4C",
			wantField: "Header.Magic",
			wantOff:   2,
			wantLeaf:  io.ErrUnexpectedEOF,
		},
		{
			name:      "short read on reserved",
			hex:       "54 4C 56 31 01 00 00",
			wantField: "Header.Reserved",
			wantOff:   7,
			wantLeaf:  io.ErrUnexpectedEOF,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(mustDecodeHex(t, tc.hex)))
			require.Error(t, err)

			// FieldError outermost.
			var fe *FieldError
			require.ErrorAs(t, err, &fe)
			require.Equal(t, tc.wantField, fe.Field)

			// OffsetError next.
			var oe *OffsetError
			require.ErrorAs(t, err, &oe)
			require.Equal(t, tc.wantOff, oe.Offset)

			// Leaf at the bottom of the chain.
			require.ErrorIs(t, err, tc.wantLeaf)
		})
	}
}

func TestDecodeHeaderFailureChain_FieldErrorOffsetErrorLeaf(t *testing.T) {
	t.Parallel()

	// Unsupported version is the cleanest case to assert the full chain shape.
	bad := mustDecodeHex(t, "54 4C 56 31 02 00 00 00")
	_, err := Decode(bytes.NewReader(bad))
	require.Error(t, err)

	// Outer error must be a *FieldError.
	fe, ok := err.(*FieldError)
	require.True(t, ok, "outer error should be *FieldError, got %T", err)
	require.Equal(t, "Header.Version", fe.Field)

	// Its Unwrap must be a *OffsetError.
	oe, ok := fe.Unwrap().(*OffsetError)
	require.True(t, ok, "FieldError.Err should be *OffsetError, got %T", fe.Unwrap())
	require.Equal(t, int64(5), oe.Offset)

	// And the leaf is the sentinel.
	require.Same(t, ErrUnsupportedVersion, oe.Unwrap())
}
