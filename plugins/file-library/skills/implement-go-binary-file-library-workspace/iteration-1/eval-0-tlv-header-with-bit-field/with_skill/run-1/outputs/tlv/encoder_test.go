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
		in   *File
		want []byte
	}{
		{
			name: "minimal_no_flags",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: Version1,
			}},
			want: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x00,
				0x00, 0x00,
			},
		},
		{
			name: "compressed_flag",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: Version1,
				Flags:   FlagCompressed,
			}},
			want: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x01,
				0x00, 0x00,
			},
		},
		{
			name: "all_known_flags",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: Version1,
				Flags:   FlagCompressed | FlagEncrypted | FlagSigned,
			}},
			want: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x07,
				0x00, 0x00,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, tc.in))
			require.Equal(t, tc.want, buf.Bytes())
		})
	}
}

func TestHeaderRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
	}{
		{
			name: "no_flags",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: Version1,
			}},
		},
		{
			name: "compressed_only",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: Version1,
				Flags:   FlagCompressed,
			}},
		},
		{
			name: "all_known_flags",
			in: &File{Header: Header{
				Magic:   Magic,
				Version: Version1,
				Flags:   FlagCompressed | FlagEncrypted | FlagSigned,
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, tc.in))

			decoded, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.in, decoded)
		})
	}
}

func TestEncodeHeader_ReservedBitsRejected(t *testing.T) {
	t.Parallel()

	// Encoder must refuse to emit bytes that the decoder would later reject.
	// This is the "fail fast" half of the round-trip contract.
	in := &File{Header: Header{
		Magic:   Magic,
		Version: Version1,
		Flags:   FlagsReservedMask, // upper bits set
	}}

	var buf bytes.Buffer
	err := Encode(&buf, in)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrReservedBitsSet)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Flags", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// Validation runs before any bytes are written.
	require.Equal(t, int64(0), oe.Offset)
}
