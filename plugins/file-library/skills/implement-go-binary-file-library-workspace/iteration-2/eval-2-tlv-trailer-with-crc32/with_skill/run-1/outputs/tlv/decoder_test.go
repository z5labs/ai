package tlv

import (
	"bytes"
	"hash/crc32"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeEmptyFileHappyPath(t *testing.T) {
	t.Parallel()

	// No header, no records, just a trailer over zero preceding bytes.
	// crc32(IEEE) of the empty byte string is 0.
	var buf [4]byte
	// Trailer = big-endian 0x00000000.
	f, err := Decode(bytes.NewReader(buf[:]))
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Equal(t, uint32(0), f.Trailer.CRC32)
}

func TestDecodeWithPrefixBytesHappyPath(t *testing.T) {
	t.Parallel()

	// Although File is currently just a Trailer, the CRC must cover any
	// bytes the decoder consumes before the trailer. We simulate this by
	// driving the lower-level decoder with a small prefix and a matching
	// CRC32 in the trailer.
	prefix := []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}
	expected := crc32.ChecksumIEEE(prefix)

	d := newDecoder(bytes.NewReader(append(append([]byte(nil), prefix...), bytesBE32(expected)...)))

	// Pretend the implementer "consumed" the prefix; CRC must update from countingReader reads.
	consumed := make([]byte, len(prefix))
	_, err := io.ReadFull(d.r, consumed)
	require.NoError(t, err)
	require.Equal(t, prefix, consumed)

	tr, err := d.readTrailer()
	require.NoError(t, err)
	require.Equal(t, expected, tr.CRC32)
}

func TestDecodeChecksumMismatch(t *testing.T) {
	t.Parallel()

	// Trailer claims CRC = 0xDEADBEEF but the preceding bytes (none) hash to 0.
	input := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(4), oe.Offset)
}

func TestDecodeTruncatedTrailer(t *testing.T) {
	t.Parallel()

	// Only 2 of the 4 trailer bytes -> ErrUnexpectedEOF wrapped in the chain.
	_, err := Decode(bytes.NewReader([]byte{0x00, 0x00}))
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)
}

// bytesBE32 returns the big-endian 4-byte encoding of v.
func bytesBE32(v uint32) []byte {
	return []byte{
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
}
