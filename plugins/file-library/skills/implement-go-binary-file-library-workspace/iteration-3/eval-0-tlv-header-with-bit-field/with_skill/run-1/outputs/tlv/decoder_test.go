package tlv

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   []byte
		want Header
	}{
		{
			name: "minimal_no_flags",
			// Magic="TLV1", Version=1, Flags=0, Reserved=0
			in: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x00,
				0x00, 0x00,
			},
			want: Header{
				Magic:    [4]byte{0x54, 0x4C, 0x56, 0x31},
				Version:  1,
				Flags:    0,
				Reserved: 0,
			},
		},
		{
			name: "compressed_flag_set",
			// Magic="TLV1", Version=1, Flags=COMPRESSED (0x01), Reserved=0
			in: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x01,
				0x00, 0x00,
			},
			want: Header{
				Magic:    [4]byte{0x54, 0x4C, 0x56, 0x31},
				Version:  1,
				Flags:    FlagCompressed,
				Reserved: 0,
			},
		},
		{
			name: "compressed_and_encrypted",
			// Magic="TLV1", Version=1, Flags=COMPRESSED|ENCRYPTED (0x03), Reserved=0
			in: []byte{
				0x54, 0x4C, 0x56, 0x31,
				0x01,
				0x03,
				0x00, 0x00,
			},
			want: Header{
				Magic:    [4]byte{0x54, 0x4C, 0x56, 0x31},
				Version:  1,
				Flags:    FlagCompressed | FlagEncrypted,
				Reserved: 0,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := Decode(bytes.NewReader(tc.in))
			require.NoError(t, err)
			require.Equal(t, tc.want, f.Header)
		})
	}
}

func TestDecodeHeader_TruncatedReturnsErrorChain(t *testing.T) {
	t.Parallel()

	// Only 2 bytes — too short to read the 8-byte header. The underlying
	// binary.Read will surface io.ErrUnexpectedEOF.
	input := []byte{0x54, 0x4C}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)

	// Leaf sentinel.
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	// FieldError wrapping the field path.
	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header", fe.Field)

	// OffsetError sandwiched between FieldError and the leaf.
	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.GreaterOrEqual(t, oe.Offset, int64(0))

	// FieldError → OffsetError → leaf chain shape.
	require.IsType(t, &FieldError{}, err)
	require.IsType(t, &OffsetError{}, fe.Err)
	require.True(t, errors.Is(oe.Err, io.ErrUnexpectedEOF))
}

func TestDecodeHeader_InvalidMagic(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x00, 0x00, 0x00, 0x00, // wrong magic
		0x01,
		0x00,
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.ErrorIs(t, err, ErrInvalidMagic)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Magic", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(8), oe.Offset)
}

func TestDecodeHeader_UnsupportedVersion(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x02, // version 2 — unsupported
		0x00,
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.ErrorIs(t, err, ErrUnsupportedVersion)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Version", fe.Field)
}

func TestDecodeHeader_ReservedFlagBitsSet(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01,
		0x08, // bit 3 set — reserved
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.ErrorIs(t, err, ErrReservedFlagsSet)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Flags", fe.Field)
}

func TestDecodeHeader_SignedFlagRejected(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01,
		0x04, // SIGNED
		0x00, 0x00,
	}

	_, err := Decode(bytes.NewReader(input))
	require.ErrorIs(t, err, ErrSignedFlagUnsupported)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Flags", fe.Field)
}

func TestDecodeHeader_ReservedFieldNonZero(t *testing.T) {
	t.Parallel()

	input := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01,
		0x00,
		0x00, 0x01, // Reserved=1
	}

	_, err := Decode(bytes.NewReader(input))
	require.ErrorIs(t, err, ErrReservedNonZero)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Reserved", fe.Field)
}
