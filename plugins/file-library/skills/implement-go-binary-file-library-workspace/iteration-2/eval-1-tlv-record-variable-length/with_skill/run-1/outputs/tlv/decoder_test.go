package tlv

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeStringRecord(t *testing.T) {
	t.Parallel()

	// Type=STRING(0x01), Length=5 (big-endian), Value="hello"
	input := []byte{
		0x01,
		0x00, 0x05,
		0x68, 0x65, 0x6C, 0x6C, 0x6F,
	}

	f, err := Decode(bytes.NewReader(input))
	require.NoError(t, err)
	require.Len(t, f.Records, 1)
	require.Equal(t, RecordTypeSTRING, f.Records[0].Type)
	require.Equal(t, uint16(5), f.Records[0].Length)
	require.Equal(t, []byte("hello"), f.Records[0].Value)
}

func TestDecodeEmptyRecord(t *testing.T) {
	t.Parallel()

	// Type=BLOB(0x03), Length=0, no Value bytes.
	input := []byte{
		0x03,
		0x00, 0x00,
	}

	f, err := Decode(bytes.NewReader(input))
	require.NoError(t, err)
	require.Len(t, f.Records, 1)
	require.Equal(t, RecordTypeBLOB, f.Records[0].Type)
	require.Equal(t, uint16(0), f.Records[0].Length)
	require.Empty(t, f.Records[0].Value)
}

func TestDecodeIntRecord(t *testing.T) {
	t.Parallel()

	// Type=INT(0x02), Length=8, Value=42 (big-endian int64).
	input := []byte{
		0x02,
		0x00, 0x08,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A,
	}

	f, err := Decode(bytes.NewReader(input))
	require.NoError(t, err)
	require.Len(t, f.Records, 1)
	require.Equal(t, RecordTypeINT, f.Records[0].Type)
	require.Equal(t, uint16(8), f.Records[0].Length)
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0x2A}, f.Records[0].Value)
}

func TestDecodeNoRecords(t *testing.T) {
	t.Parallel()

	f, err := Decode(bytes.NewReader(nil))
	require.NoError(t, err)
	require.Empty(t, f.Records)
}

func TestDecodeMultipleRecords(t *testing.T) {
	t.Parallel()

	input := []byte{
		// Record 1: STRING "hi"
		0x01, 0x00, 0x02, 0x68, 0x69,
		// Record 2: INT 1
		0x02, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
	}

	f, err := Decode(bytes.NewReader(input))
	require.NoError(t, err)
	require.Len(t, f.Records, 2)
	require.Equal(t, RecordTypeSTRING, f.Records[0].Type)
	require.Equal(t, []byte("hi"), f.Records[0].Value)
	require.Equal(t, RecordTypeINT, f.Records[1].Type)
}

func TestDecodeTruncatedLength(t *testing.T) {
	t.Parallel()

	// Type byte present, length prefix truncated.
	input := []byte{0x01, 0x00}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Record.Length", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// Offset reflects bytes actually consumed at the failure site: 1 type byte
	// plus the 1 partial byte of the truncated big-endian uint16 length prefix.
	require.Equal(t, int64(2), oe.Offset)
}

func TestDecodeTruncatedValue(t *testing.T) {
	t.Parallel()

	// Length=5 announced but only 2 bytes follow.
	input := []byte{
		0x01,
		0x00, 0x05,
		0x68, 0x69,
	}

	_, err := Decode(bytes.NewReader(input))
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Record.Value", fe.Field)
}
