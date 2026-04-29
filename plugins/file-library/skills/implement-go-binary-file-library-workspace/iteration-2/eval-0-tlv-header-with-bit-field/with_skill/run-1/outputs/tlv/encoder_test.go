package tlv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   Header
		want []byte
	}{
		{
			name: "minimal_no_flags",
			in: Header{
				Magic:    [4]byte{'T', 'L', 'V', '1'},
				Version:  1,
				Flags:    0,
				Reserved: 0,
			},
			want: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x00,
				0x00, 0x00,
			},
		},
		{
			name: "compressed_only",
			in: Header{
				Magic:    [4]byte{'T', 'L', 'V', '1'},
				Version:  1,
				Flags:    FlagCompressed,
				Reserved: 0,
			},
			want: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x01,
				0x00, 0x00,
			},
		},
		{
			name: "all_three_flags",
			in: Header{
				Magic:    [4]byte{'T', 'L', 'V', '1'},
				Version:  1,
				Flags:    FlagCompressed | FlagEncrypted | FlagSigned,
				Reserved: 0,
			},
			want: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x07,
				0x00, 0x00,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			err := Encode(&buf, &File{Header: tc.in})
			require.NoError(t, err)
			require.Equal(t, tc.want, buf.Bytes())
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
	}{
		{
			name: "no_flags",
			in: &File{Header: Header{
				Magic:   [4]byte{'T', 'L', 'V', '1'},
				Version: 1,
				Flags:   0,
			}},
		},
		{
			name: "compressed",
			in: &File{Header: Header{
				Magic:   [4]byte{'T', 'L', 'V', '1'},
				Version: 1,
				Flags:   FlagCompressed,
			}},
		},
		{
			name: "all_three",
			in: &File{Header: Header{
				Magic:   [4]byte{'T', 'L', 'V', '1'},
				Version: 1,
				Flags:   FlagCompressed | FlagEncrypted | FlagSigned,
			}},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, tc.in))

			got, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.in, got)
		})
	}
}
