package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeEmptyFileWritesTrailerOnly(t *testing.T) {
	t.Parallel()

	// With no Header/Records implemented, an empty File encodes to just
	// a 4-byte trailer carrying the CRC32 of an empty body. The Trailer
	// field on input is ignored — the encoder always recomputes it.
	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, &File{}))

	want := make([]byte, 4)
	binary.BigEndian.PutUint32(want, crc32.ChecksumIEEE(nil))
	require.Equal(t, want, buf.Bytes())
}

func TestEncodeIgnoresInputTrailerCRC(t *testing.T) {
	t.Parallel()

	// Even if the caller pre-fills File.Trailer.CRC32 with a wrong
	// value, the encoder must compute the CRC32 from the bytes it
	// actually wrote, not trust the input.
	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, &File{Trailer: Trailer{CRC32: 0xDEADBEEF}}))

	want := make([]byte, 4)
	binary.BigEndian.PutUint32(want, crc32.ChecksumIEEE(nil))
	require.Equal(t, want, buf.Bytes())
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	original := &File{}

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, original))

	decoded, err := Decode(&buf)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	// The decoder fills in Trailer.CRC32 from the bytes it read, so the
	// round-trip equals the recomputed value.
	want := crc32.ChecksumIEEE(nil)
	require.Equal(t, want, decoded.Trailer.CRC32)
}
