package tlv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    File
		hex  string
	}{
		{
			name: "no flags",
			f: File{Header: Header{
				Magic:   Magic,
				Version: 1,
			}},
			hex: "54 4C 56 31 01 00 00 00",
		},
		{
			name: "compressed",
			f: File{Header: Header{
				Magic:   Magic,
				Version: 1,
				Flags:   FlagCompressed,
			}},
			hex: "54 4C 56 31 01 01 00 00",
		},
		{
			name: "compressed+encrypted",
			f: File{Header: Header{
				Magic:   Magic,
				Version: 1,
				Flags:   FlagCompressed | FlagEncrypted,
			}},
			hex: "54 4C 56 31 01 03 00 00",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, &tt.f)
			require.NoError(t, err)
			require.Equal(t, mustHex(t, tt.hex), buf.Bytes())
		})
	}
}

func TestEncodeHeaderFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		f         File
		wantField string
		wantLeaf  error
	}{
		{
			name:      "bad magic",
			f:         File{Header: Header{Magic: [4]byte{'X', 'X', 'X', 'X'}, Version: 1}},
			wantField: "Header.Magic",
			wantLeaf:  ErrBadMagic,
		},
		{
			name:      "wrong version",
			f:         File{Header: Header{Magic: Magic, Version: 2}},
			wantField: "Header.Version",
			wantLeaf:  ErrUnsupportedVersion,
		},
		{
			name:      "reserved flag bit",
			f:         File{Header: Header{Magic: Magic, Version: 1, Flags: 0x80}},
			wantField: "Header.Flags",
			wantLeaf:  ErrReservedFlagSet,
		},
		{
			name:      "signed flag",
			f:         File{Header: Header{Magic: Magic, Version: 1, Flags: FlagSigned}},
			wantField: "Header.Flags",
			wantLeaf:  ErrSignedFlagUnsupported,
		},
		{
			name:      "reserved non-zero",
			f:         File{Header: Header{Magic: Magic, Version: 1, Reserved: 1}},
			wantField: "Header.Reserved",
			wantLeaf:  ErrReservedNotZero,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, &tt.f)
			require.Error(t, err)

			var fe *FieldError
			require.ErrorAs(t, err, &fe)
			require.Equal(t, tt.wantField, fe.Field)

			var oe *OffsetError
			require.ErrorAs(t, err, &oe)
			require.Equal(t, int64(0), oe.Offset)

			require.ErrorIs(t, err, tt.wantLeaf)
		})
	}
}

func TestRoundTripHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   File
	}{
		{name: "no flags", in: File{Header: Header{Magic: Magic, Version: 1}}},
		{name: "compressed", in: File{Header: Header{Magic: Magic, Version: 1, Flags: FlagCompressed}}},
		{name: "encrypted", in: File{Header: Header{Magic: Magic, Version: 1, Flags: FlagEncrypted}}},
		{name: "compressed+encrypted", in: File{Header: Header{Magic: Magic, Version: 1, Flags: FlagCompressed | FlagEncrypted}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, &tt.in))

			got, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, tt.in, *got)
		})
	}
}
