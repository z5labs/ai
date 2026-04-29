package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeTrailerHappyPath(t *testing.T) {
	t.Parallel()

	// Same minimal-header bytes as the decoder happy-path test.
	payload := []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}

	var buf bytes.Buffer
	err := Encode(&buf, &File{Payload: payload})
	require.NoError(t, err)

	expectedCRC := crc32.ChecksumIEEE(payload)
	expectedTrailer := make([]byte, 4)
	binary.BigEndian.PutUint32(expectedTrailer, expectedCRC)

	expected := append(append([]byte{}, payload...), expectedTrailer...)
	require.Equal(t, expected, buf.Bytes())
}

func TestEncodeTrailerEmptyPayload(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{})
	require.NoError(t, err)

	// CRC32 of zero input bytes is 0; trailer is four zero bytes.
	require.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, buf.Bytes())
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		payload []byte
	}{
		{"empty", nil},
		{"minimal_header", []byte{0x54, 0x4C, 0x56, 0x31, 0x01, 0x00, 0x00, 0x00}},
		{"single_byte", []byte{0xAB}},
		{"longer_payload", bytes.Repeat([]byte{0x5A}, 256)},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			original := &File{Payload: tc.payload}

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, original))

			decoded, err := Decode(&buf)
			require.NoError(t, err)
			require.NotNil(t, decoded)

			// File.Payload may be nil vs empty slice after a round-trip; compare
			// content via bytes.Equal so the test is independent of that.
			require.True(t, bytes.Equal(original.Payload, decoded.Payload),
				"payload mismatch: want %x, got %x", original.Payload, decoded.Payload)
			require.Equal(t, crc32.ChecksumIEEE(tc.payload), decoded.Trailer.CRC32)
		})
	}
}
