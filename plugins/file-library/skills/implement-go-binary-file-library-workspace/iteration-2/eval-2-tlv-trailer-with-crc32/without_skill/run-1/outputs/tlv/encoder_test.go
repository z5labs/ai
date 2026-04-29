package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEncodeEmptyFileWritesTrailerOnly verifies that encoding an empty file
// (no header/records yet) emits exactly the 4-byte big-endian CRC32 of the
// empty body.
func TestEncodeEmptyFileWritesTrailerOnly(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, &File{}))

	require.Len(t, buf.Bytes(), 4)
	got := binary.BigEndian.Uint32(buf.Bytes())
	require.Equal(t, crc32.ChecksumIEEE(nil), got)
}

// TestEncodeDecodeRoundTrip verifies that what Encode writes, Decode accepts.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, &File{}))

	f, err := Decode(&buf)
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Equal(t, crc32.ChecksumIEEE(nil), f.Trailer.CRC32)
}
