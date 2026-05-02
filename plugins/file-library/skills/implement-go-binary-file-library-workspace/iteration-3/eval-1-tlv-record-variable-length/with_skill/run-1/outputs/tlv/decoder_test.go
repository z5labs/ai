package tlv

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeStubReturnsErrUnimplemented(t *testing.T) {
	t.Parallel()

	_, err := Decode(bytes.NewReader([]byte{0x00}))
	require.ErrorIs(t, err, errUnimplemented)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "File", fe.Field)
}

func TestReadRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input []byte
		want  Record
	}{
		{
			name: "typical_string_record",
			input: []byte{
				// Type = STRING (0x01)
				0x01,
				// Length = 5 (big-endian uint16)
				0x00, 0x05,
				// Value = "hello"
				0x68, 0x65, 0x6C, 0x6C, 0x6F,
			},
			want: Record{
				Type:   RecordTypeSTRING,
				Length: 5,
				Value:  []byte("hello"),
			},
		},
		{
			name: "empty_record_length_zero",
			input: []byte{
				// Type = BLOB (0x03)
				0x03,
				// Length = 0
				0x00, 0x00,
				// no Value bytes
			},
			want: Record{
				Type:   RecordTypeBLOB,
				Length: 0,
				Value:  []byte{},
			},
		},
		{
			name: "int_record_length_eight",
			input: []byte{
				// Type = INT (0x02)
				0x02,
				// Length = 8
				0x00, 0x08,
				// Value = 42 (big-endian int64)
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A,
			},
			want: Record{
				Type:   RecordTypeINT,
				Length: 8,
				Value:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := newDecoder(bytes.NewReader(tc.input))
			got, err := d.readRecord()
			require.NoError(t, err)
			require.Equal(t, tc.want.Type, got.Type)
			require.Equal(t, tc.want.Length, got.Length)
			require.Equal(t, tc.want.Value, got.Value)
		})
	}
}

func TestReadRecordTruncatedType(t *testing.T) {
	t.Parallel()

	d := newDecoder(bytes.NewReader([]byte{}))
	_, err := d.readRecord()
	require.Error(t, err)
	require.ErrorIs(t, err, io.EOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Record.Type", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(0), oe.Offset)
}

func TestReadRecordTruncatedLength(t *testing.T) {
	t.Parallel()

	// Type byte present, Length truncated.
	d := newDecoder(bytes.NewReader([]byte{0x01, 0x00}))
	_, err := d.readRecord()
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Record.Length", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	// 1 byte (Type) consumed, then attempt to read 2-byte Length partially.
	require.Equal(t, int64(2), oe.Offset)
}

func TestReadRecordTruncatedValue(t *testing.T) {
	t.Parallel()

	// Type=STRING, Length=5, but only 3 value bytes.
	d := newDecoder(bytes.NewReader([]byte{0x01, 0x00, 0x05, 0x68, 0x65, 0x6C}))
	_, err := d.readRecord()
	require.Error(t, err)
	require.ErrorIs(t, err, io.ErrUnexpectedEOF)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Record.Value", fe.Field)
}

func TestReadRecordUnknownType(t *testing.T) {
	t.Parallel()

	// Type=0x99 (unknown), Length=0.
	d := newDecoder(bytes.NewReader([]byte{0x99, 0x00, 0x00}))
	_, err := d.readRecord()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnknownRecordType)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Record.Type", fe.Field)
}
