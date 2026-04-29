package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeTrailerHappyPath(t *testing.T) {
	t.Parallel()

	// Payload is the minimal TLV1 header bytes from SPEC.md's "Minimal" example:
	//   54 4C 56 31  Magic = "TLV1"
	//   01           Version = 1
	//   00           Flags = 0
	//   00 00        Reserved = 0
	payload := []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}

	// Compute the expected big-endian CRC32 (IEEE) trailer.
	crc := crc32.ChecksumIEEE(payload)
	trailer := make([]byte, 4)
	binary.BigEndian.PutUint32(trailer, crc)

	input := append(append([]byte{}, payload...), trailer...)

	f, err := Decode(bytes.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Equal(t, payload, f.Payload)
	require.Equal(t, crc, f.Trailer.CRC32)
}

func TestDecodeTrailerEmptyPayload(t *testing.T) {
	t.Parallel()

	// Edge case: zero payload bytes. CRC32 of empty input is 0.
	trailer := []byte{0x00, 0x00, 0x00, 0x00}

	f, err := Decode(bytes.NewReader(trailer))
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Empty(t, f.Payload)
	require.Equal(t, uint32(0), f.Trailer.CRC32)
}

func TestDecodeTrailerChecksumMismatch(t *testing.T) {
	t.Parallel()

	// Same payload as the happy path...
	payload := []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}
	// ... but the trailer carries a deliberately wrong CRC32.
	badTrailer := []byte{0xDE, 0xAD, 0xBE, 0xEF}

	input := append(append([]byte{}, payload...), badTrailer...)

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// Offset reflects bytes consumed: 8 bytes of payload + 4 bytes of trailer.
	require.Equal(t, int64(len(input)), oe.Offset)
}

func TestDecodeTrailerTruncated(t *testing.T) {
	t.Parallel()

	// Fewer than 4 bytes total — there's no trailer at all.
	_, err := Decode(bytes.NewReader([]byte{0x01, 0x02}))
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)
}
