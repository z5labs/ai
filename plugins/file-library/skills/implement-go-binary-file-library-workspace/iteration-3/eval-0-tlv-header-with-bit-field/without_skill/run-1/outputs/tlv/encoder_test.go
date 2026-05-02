package tlv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeHeader(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   *File
		want string
	}{
		{
			name: "zero-value defaults to canonical Magic and Version=1",
			in:   &File{},
			want: "54 4C 56 31 01 00 00 00",
		},
		{
			name: "compressed flag",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: 1,
				Flags:   FlagCompressed,
			}},
			want: "54 4C 56 31 01 01 00 00",
		},
		{
			name: "all defined flags",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: 1,
				Flags:   FlagCompressed | FlagEncrypted | FlagSigned,
			}},
			want: "54 4C 56 31 01 07 00 00",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			require.NoError(t, err)
			require.Equal(t, mustDecodeHex(t, tc.want), buf.Bytes())
		})
	}
}

func TestEncodeHeaderErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		in        *File
		wantField string
		wantLeaf  error
	}{
		{
			name: "bad magic",
			in: &File{Header: Header{
				Magic:   [4]byte{'X', 'X', 'X', 'X'},
				Version: 1,
			}},
			wantField: "Header.Magic",
			wantLeaf:  ErrBadMagic,
		},
		{
			name: "unsupported version",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: 2,
			}},
			wantField: "Header.Version",
			wantLeaf:  ErrUnsupportedVersion,
		},
		{
			name: "reserved flag bits set",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: 1,
				Flags:   Flags(0x80),
			}},
			wantField: "Header.Flags",
			wantLeaf:  ErrReservedFlagBitsSet,
		},
		{
			name: "reserved non-zero",
			in: &File{Header: Header{
				Magic:    Magic,
				Version:  1,
				Reserved: 1,
			}},
			wantField: "Header.Reserved",
			wantLeaf:  ErrReservedNonZero,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Encode(&buf, tc.in)
			require.Error(t, err)

			var fe *FieldError
			require.ErrorAs(t, err, &fe)
			require.Equal(t, tc.wantField, fe.Field)

			var oe *OffsetError
			require.ErrorAs(t, err, &oe)
			_ = oe // Offset varies by error site, just confirm it's in the chain.

			require.ErrorIs(t, err, tc.wantLeaf)
		})
	}
}

func TestHeaderRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		hdr  Header
	}{
		{name: "no flags", hdr: Header{Magic: Magic, Version: 1}},
		{name: "compressed", hdr: Header{Magic: Magic, Version: 1, Flags: FlagCompressed}},
		{name: "encrypted", hdr: Header{Magic: Magic, Version: 1, Flags: FlagEncrypted}},
		{name: "signed", hdr: Header{Magic: Magic, Version: 1, Flags: FlagSigned}},
		{
			name: "all flags",
			hdr:  Header{Magic: Magic, Version: 1, Flags: FlagCompressed | FlagEncrypted | FlagSigned},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			in := &File{Header: tc.hdr}

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, in))

			out, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, in.Header, out.Header)
		})
	}
}
