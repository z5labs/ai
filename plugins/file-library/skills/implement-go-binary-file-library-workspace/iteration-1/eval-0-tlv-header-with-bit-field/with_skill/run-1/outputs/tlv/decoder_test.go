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
			input: []byte{
				// Magic: "TLV1"
				0x54, 0x4C, 0x56, 0x31,
				// Version
				0x01,
				// Flags
				0x00,
				// Reserved
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
				0x01, // Flags = COMPRESSED
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
			name: "all_known_flags",
			input: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x07, // COMPRESSED | ENCRYPTED | SIGNED
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, err := Decode(bytes.NewReader(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.want, f.Header)
		})
	}
}

func TestDecodeHeader_TruncatedMagic(t *testing.T) {
	t.Parallel()

	// Only 2 bytes — short of the 8-byte header. binary.Read should fail with
	// io.ErrUnexpectedEOF and the error chain must report Header.
	input := []byte{0x54, 0x4C}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// 2 bytes were consumed before binary.Read returned ErrUnexpectedEOF.
	require.Equal(t, int64(2), oe.Offset)
}

func TestDecodeHeader_BadMagic(t *testing.T) {
	t.Parallel()

	input := []byte{
		// Wrong magic
		'X', 'X', 'X', 'X',
		0x01,
		0x00,
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Magic", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(8), oe.Offset)

	var uve *UnexpectedValueError
	require.ErrorAs(t, err, &uve)
	require.Equal(t, "Header.Magic", uve.Field)
}

func TestDecodeHeader_BadVersion(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x02, // unsupported
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

func TestDecodeHeader_ReservedBitsInFlags(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01,
		0x80, // upper reserved bit set
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrReservedBitsSet)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Flags", fe.Field)
}

func TestDecodeHeader_ReservedNonZero(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01,
		0x00,
		0x00, 0x01, // Reserved must be zero
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrReservedBitsSet)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Reserved", fe.Field)
}
