package tlv

import (
	"bytes"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeEmptyFileWritesZeroCRC(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, &File{}))
	// crc32(IEEE) of empty bytes is 0, so the trailer is 4 zero bytes.
	require.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, buf.Bytes())
}

func TestEncodeWithPrefixBytes(t *testing.T) {
	t.Parallel()

	// Drive the lower-level encoder directly so we can write some prefix
	// bytes through the counting/CRC writer before invoking writeTrailer.
	prefix := []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}

	var buf bytes.Buffer
	e := newEncoder(&buf)
	_, err := e.w.Write(prefix)
	require.NoError(t, err)

	require.NoError(t, e.writeTrailer())

	want := append([]byte(nil), prefix...)
	want = append(want,
		byte(crc32.ChecksumIEEE(prefix)>>24),
		byte(crc32.ChecksumIEEE(prefix)>>16),
		byte(crc32.ChecksumIEEE(prefix)>>8),
		byte(crc32.ChecksumIEEE(prefix)),
	)
	require.Equal(t, want, buf.Bytes())
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	original := &File{}

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, original))

	decoded, err := Decode(&buf)
	require.NoError(t, err)
	require.Equal(t, original, decoded)
}

func TestEncodeThenTamperFailsDecode(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, &File{}))

	// Flip a bit in the encoded trailer; decode must report a CRC mismatch.
	encoded := buf.Bytes()
	encoded[0] ^= 0xFF

	_, err := Decode(bytes.NewReader(encoded))
	require.ErrorIs(t, err, ErrChecksumMismatch)
}
