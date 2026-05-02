package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// trailerBytes returns 4 big-endian bytes encoding crc.
func trailerBytes(crc uint32) []byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], crc)
	return b[:]
}

func TestDecodeTrailerHappyPath(t *testing.T) {
	t.Parallel()

	// Body is the bytes preceding the trailer; per the spec the CRC32
	// covers offsets 0..len(body)-1.
	body := []byte{
		0x54, 0x4C, 0x56, 0x31, // Magic = "TLV1"
		0x01,       // Version = 1
		0x00,       // Flags = 0
		0x00, 0x00, // Reserved = 0
	}
	wantCRC := crc32.ChecksumIEEE(body)

	input := append([]byte(nil), body...)
	input = append(input, trailerBytes(wantCRC)...)

	f, err := Decode(bytes.NewReader(input))
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Equal(t, wantCRC, f.Trailer.CRC32)
}

func TestDecodeTrailerEmptyBody(t *testing.T) {
	t.Parallel()

	// A file with zero records is legal; a file with no header would
	// also have an empty body. The spec doesn't bar the trailer-only
	// case, and we want to verify the running-CRC mechanism works at
	// length 0 (CRC32 of empty input = 0).
	wantCRC := crc32.ChecksumIEEE(nil)

	input := trailerBytes(wantCRC)

	f, err := Decode(bytes.NewReader(input))
	require.NoError(t, err)
	require.Equal(t, wantCRC, f.Trailer.CRC32)
}

func TestDecodeTrailerCRCMismatch(t *testing.T) {
	t.Parallel()

	body := []byte{
		0x54, 0x4C, 0x56, 0x31,
		0x01, 0x00, 0x00, 0x00,
	}
	// Corrupt the trailer by flipping a bit in the CRC.
	wrongCRC := crc32.ChecksumIEEE(body) ^ 0xFFFFFFFF

	input := append([]byte(nil), body...)
	input = append(input, trailerBytes(wrongCRC)...)

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// Offset is the count of bytes consumed before the failure: full file length.
	require.Equal(t, int64(len(input)), oe.Offset)
}

func TestDecodeTrailerTruncated(t *testing.T) {
	t.Parallel()

	// Fewer than 4 bytes total — the trailer can't be read.
	input := []byte{0x00, 0x00}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)
}
