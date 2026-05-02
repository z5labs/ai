package tlv

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode_HappyPath_NoRecords(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{
		Header: Header{
			Magic:   [4]byte{'T', 'L', 'V', '1'},
			Version: 1,
		},
	})
	require.NoError(t, err)
	require.Equal(t, 12, buf.Len()) // 8-byte header + 4-byte trailer.

	// The trailer must be the CRC32 of the header.
	body := buf.Bytes()
	expected := crc32.ChecksumIEEE(body[:8])
	got := binary.BigEndian.Uint32(body[8:12])
	require.Equal(t, expected, got)
}

func TestEncode_HappyPath_OneStringRecord(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{
		Header: Header{
			Magic:   [4]byte{'T', 'L', 'V', '1'},
			Version: 1,
		},
		Records: []Record{
			{Type: 0x01, Value: []byte("hello")},
		},
	})
	require.NoError(t, err)

	body := buf.Bytes()
	require.Equal(t, 8+3+5+4, len(body))
	expected := crc32.ChecksumIEEE(body[:len(body)-4])
	got := binary.BigEndian.Uint32(body[len(body)-4:])
	require.Equal(t, expected, got)
}

func TestEncodeDecode_RoundTrip(t *testing.T) {
	t.Parallel()

	in := &File{
		Header:  Header{Magic: [4]byte{'T', 'L', 'V', '1'}, Version: 1},
		Records: []Record{{Type: 0x01, Value: []byte("hello")}, {Type: 0x03, Value: []byte{0xDE, 0xAD, 0xBE, 0xEF}}},
	}
	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, in))

	out, err := Decode(&buf)
	require.NoError(t, err)
	require.Equal(t, in.Header, out.Header)
	require.Len(t, out.Records, 2)
	require.Equal(t, in.Records[0].Type, out.Records[0].Type)
	require.Equal(t, []byte("hello"), out.Records[0].Value)
	require.Equal(t, in.Records[1].Type, out.Records[1].Type)
	require.Equal(t, []byte{0xDE, 0xAD, 0xBE, 0xEF}, out.Records[1].Value)
	require.NotZero(t, out.Trailer.CRC32)
}
