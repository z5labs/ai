package tlv

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input []byte
		want  Header
	}{
		{
			name: "minimal_no_flags",
			// Magic "TLV1", Version=1, Flags=0, Reserved=0
			input: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x00,
				0x00, 0x00,
			},
			want: Header{
				Magic:    [4]byte{'T', 'L', 'V', '1'},
				Version:  1,
				Flags:    0,
				Reserved: 0,
			},
		},
		{
			name: "compressed_flag",
			input: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x01,
				0x00, 0x00,
			},
			want: Header{
				Magic:    [4]byte{'T', 'L', 'V', '1'},
				Version:  1,
				Flags:    FlagCompressed,
				Reserved: 0,
			},
		},
		{
			name: "all_three_flags",
			input: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x07, // 0x01 | 0x02 | 0x04
				0x00, 0x00,
			},
			want: Header{
				Magic:    [4]byte{'T', 'L', 'V', '1'},
				Version:  1,
				Flags:    FlagCompressed | FlagEncrypted | FlagSigned,
				Reserved: 0,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, err := Decode(bytes.NewReader(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.want, f.Header)
		})
	}
}

func TestDecodeHeaderTruncated(t *testing.T) {
	t.Parallel()

	// Only 5 bytes — truncated mid-header. Magic + Version read OK, then
	// Flags/Reserved short-read.
	input := []byte{0x54, 0x4C, 0x56, 0x31, 0x01}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// 5 bytes were consumed before the read failed.
	require.Equal(t, int64(5), oe.Offset)
}

func TestDecodeHeaderInvalidMagic(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x00, 0x00, 0x00, 0x00, // wrong magic
		0x01,
		0x00,
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidMagic)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Magic", fe.Field)
}

func TestDecodeHeaderUnsupportedVersion(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x02, // version 2 is not supported
		0x00,
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnsupportedVersion)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Version", fe.Field)
}

func TestDecodeHeaderReservedFlagsBitsSet(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01,
		0x08, // bit 3 set; reserved
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrReservedBitsSet)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Flags", fe.Field)
}

func TestDecodeHeaderReservedFieldNonZero(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01,
		0x00,
		0x00, 0x01, // Reserved = 1 (must be 0)
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrReservedBitsSet)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Reserved", fe.Field)
}
