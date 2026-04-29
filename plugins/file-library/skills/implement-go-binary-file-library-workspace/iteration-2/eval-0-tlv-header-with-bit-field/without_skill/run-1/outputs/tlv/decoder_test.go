package tlv

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// mustHex decodes a whitespace-tolerant hex string for use in test fixtures.
func mustHex(t *testing.T, s string) []byte {
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

	tests := []struct {
		name string
		hex  string
		want Header
	}{
		{
			name: "no flags",
			hex:  "54 4C 56 31 01 00 00 00",
			want: Header{
				Magic:    Magic,
				Version:  1,
				Flags:    0,
				Reserved: 0,
			},
		},
		{
			name: "compressed",
			hex:  "54 4C 56 31 01 01 00 00",
			want: Header{
				Magic:    Magic,
				Version:  1,
				Flags:    FlagCompressed,
				Reserved: 0,
			},
		},
		{
			name: "encrypted",
			hex:  "54 4C 56 31 01 02 00 00",
			want: Header{
				Magic:    Magic,
				Version:  1,
				Flags:    FlagEncrypted,
				Reserved: 0,
			},
		},
		{
			name: "compressed+encrypted",
			hex:  "54 4C 56 31 01 03 00 00",
			want: Header{
				Magic:    Magic,
				Version:  1,
				Flags:    FlagCompressed | FlagEncrypted,
				Reserved: 0,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := Decode(bytes.NewReader(mustHex(t, tt.hex)))
			require.NoError(t, err)
			require.NotNil(t, f)
			require.Equal(t, tt.want, f.Header)
		})
	}
}

func TestDecodeHeaderFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		hex       string
		wantField string
		wantLeaf  error
		wantOff   int64
	}{
		{
			name:      "bad magic",
			hex:       "00 00 00 00 01 00 00 00",
			wantField: "Header.Magic",
			wantLeaf:  ErrBadMagic,
			wantOff:   8,
		},
		{
			name:      "unsupported version",
			hex:       "54 4C 56 31 02 00 00 00",
			wantField: "Header.Version",
			wantLeaf:  ErrUnsupportedVersion,
			wantOff:   8,
		},
		{
			name:      "reserved flag bit",
			hex:       "54 4C 56 31 01 80 00 00",
			wantField: "Header.Flags",
			wantLeaf:  ErrReservedFlagSet,
			wantOff:   8,
		},
		{
			name:      "signed flag unsupported",
			hex:       "54 4C 56 31 01 04 00 00",
			wantField: "Header.Flags",
			wantLeaf:  ErrSignedFlagUnsupported,
			wantOff:   8,
		},
		{
			name:      "reserved non-zero",
			hex:       "54 4C 56 31 01 00 00 01",
			wantField: "Header.Reserved",
			wantLeaf:  ErrReservedNotZero,
			wantOff:   8,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(mustHex(t, tt.hex)))
			require.Error(t, err)

			// Verify the FieldError → OffsetError → leaf chain.
			var fe *FieldError
			require.ErrorAs(t, err, &fe)
			require.Equal(t, tt.wantField, fe.Field)

			var oe *OffsetError
			require.ErrorAs(t, err, &oe)
			require.Equal(t, tt.wantOff, oe.Offset)

			require.ErrorIs(t, err, tt.wantLeaf)
		})
	}
}

func TestDecodeShortHeader(t *testing.T) {
	t.Parallel()

	// Only 4 bytes available: io.ReadFull should report unexpected EOF, wrapped.
	_, err := Decode(bytes.NewReader(mustHex(t, "54 4C 56 31")))
	require.Error(t, err)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header", fe.Field)

	require.ErrorIs(t, err, io.ErrUnexpectedEOF)
}
