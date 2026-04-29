package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDecodeTrailerHappyPath verifies that the decoder accepts a payload whose
// last four bytes are the IEEE CRC32 of the preceding bytes and surfaces that
// CRC on the returned File.
func TestDecodeTrailerHappyPath(t *testing.T) {
	t.Parallel()

	// Use an arbitrary "body" — the trailer logic shouldn't care about its
	// shape (header/records are not implemented yet).
	body := []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}
	want := crc32.ChecksumIEEE(body)

	var buf bytes.Buffer
	buf.Write(body)
	binary.Write(&buf, binary.BigEndian, want)

	f, err := Decode(&buf)
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Equal(t, want, f.Trailer.CRC32)
}

// TestDecodeTrailerEmptyBody verifies that a file with no preceding bytes
// (trailer-only, CRC32 of empty input) decodes successfully.
func TestDecodeTrailerEmptyBody(t *testing.T) {
	t.Parallel()

	want := crc32.ChecksumIEEE(nil)

	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, want)

	f, err := Decode(&buf)
	require.NoError(t, err)
	require.Equal(t, want, f.Trailer.CRC32)
}

// TestDecodeTrailerChecksumMismatch verifies that a tampered trailer surfaces
// ErrChecksumMismatch, wrapped through wrapErr (FieldError → OffsetError →
// leaf).
func TestDecodeTrailerChecksumMismatch(t *testing.T) {
	t.Parallel()

	body := []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}
	good := crc32.ChecksumIEEE(body)
	bad := good ^ 0xFFFFFFFF // any value distinct from good

	var buf bytes.Buffer
	buf.Write(body)
	binary.Write(&buf, binary.BigEndian, bad)

	_, err := Decode(&buf)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// The full payload has been read by the time we detect the mismatch.
	require.Equal(t, int64(len(body)+4), oe.Offset)
}

// TestDecodeTrailerTooShort verifies that an input shorter than the 4-byte
// trailer surfaces a typed error.
func TestDecodeTrailerTooShort(t *testing.T) {
	t.Parallel()

	_, err := Decode(bytes.NewReader([]byte{0x00}))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalid)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer", fe.Field)
}
