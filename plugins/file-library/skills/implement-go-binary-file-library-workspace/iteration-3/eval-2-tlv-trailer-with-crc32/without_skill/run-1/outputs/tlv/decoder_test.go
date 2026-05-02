package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

// minimalValidFile builds a TLV1 byte stream with a header, the supplied
// records, and a correctly-computed CRC32 trailer. It returns the raw bytes
// and the CRC value that was written.
func minimalValidFile(t *testing.T, records []Record) ([]byte, uint32) {
	t.Helper()
	var body bytes.Buffer
	// Header.
	body.Write([]byte{'T', 'L', 'V', '1', 0x01, 0x00, 0x00, 0x00})
	// Records.
	for _, r := range records {
		var hdr [3]byte
		hdr[0] = r.Type
		length := r.Length
		if length == 0 && len(r.Value) > 0 {
			length = uint16(len(r.Value))
		}
		binary.BigEndian.PutUint16(hdr[1:3], length)
		body.Write(hdr[:])
		body.Write(r.Value[:length])
	}
	sum := crc32.ChecksumIEEE(body.Bytes())
	var trailer [4]byte
	binary.BigEndian.PutUint32(trailer[:], sum)
	body.Write(trailer[:])
	return body.Bytes(), sum
}

func TestDecode_HappyPath_NoRecords(t *testing.T) {
	t.Parallel()

	raw, sum := minimalValidFile(t, nil)
	f, err := Decode(bytes.NewReader(raw))
	require.NoError(t, err)
	require.NotNil(t, f)
	require.Equal(t, [4]byte{'T', 'L', 'V', '1'}, f.Header.Magic)
	require.Equal(t, uint8(1), f.Header.Version)
	require.Equal(t, uint8(0), f.Header.Flags)
	require.Equal(t, uint16(0), f.Header.Reserved)
	require.Empty(t, f.Records)
	require.Equal(t, sum, f.Trailer.CRC32)
}

func TestDecode_HappyPath_OneStringRecord(t *testing.T) {
	t.Parallel()

	rec := Record{Type: 0x01, Value: []byte("hello")}
	raw, sum := minimalValidFile(t, []Record{rec})
	f, err := Decode(bytes.NewReader(raw))
	require.NoError(t, err)
	require.Len(t, f.Records, 1)
	require.Equal(t, uint8(0x01), f.Records[0].Type)
	require.Equal(t, uint16(5), f.Records[0].Length)
	require.Equal(t, []byte("hello"), f.Records[0].Value)
	require.Equal(t, sum, f.Trailer.CRC32)
}

func TestDecode_ChecksumMismatch(t *testing.T) {
	t.Parallel()

	raw, _ := minimalValidFile(t, []Record{{Type: 0x01, Value: []byte("hello")}})
	// Corrupt the trailer (last 4 bytes) so it no longer matches the CRC of
	// the preceding bytes.
	raw[len(raw)-1] ^= 0xFF

	_, err := Decode(bytes.NewReader(raw))
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChecksumMismatch)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Trailer.CRC32", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// The trailer occupies the last 4 bytes; offset after reading the trailer
	// is the full file length.
	require.Equal(t, int64(len(raw)), oe.Offset)
}

func TestDecode_ChecksumMismatch_CorruptBody(t *testing.T) {
	t.Parallel()

	raw, _ := minimalValidFile(t, []Record{{Type: 0x01, Value: []byte("hello")}})
	// Corrupt a byte in the record value so the running CRC diverges from
	// the (still-valid-shape) trailer.
	// Record header starts after the 8-byte file header: type(1)+len(2)+value(5).
	// First value byte is at offset 8+3 = 11.
	raw[11] ^= 0x01

	_, err := Decode(bytes.NewReader(raw))
	require.ErrorIs(t, err, ErrChecksumMismatch)
}
