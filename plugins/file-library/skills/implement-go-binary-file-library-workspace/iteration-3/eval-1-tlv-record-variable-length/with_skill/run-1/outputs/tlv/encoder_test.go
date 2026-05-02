package tlv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeStubReturnsErrUnimplemented(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{})
	require.ErrorIs(t, err, errUnimplemented)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "File", fe.Field)
}

func TestWriteRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   Record
		want []byte
	}{
		{
			name: "typical_string_record",
			in: Record{
				Type:   RecordTypeSTRING,
				Length: 5,
				Value:  []byte("hello"),
			},
			want: []byte{
				0x01,                         // Type
				0x00, 0x05,                   // Length
				0x68, 0x65, 0x6C, 0x6C, 0x6F, // Value = "hello"
			},
		},
		{
			name: "empty_record_length_zero",
			in: Record{
				Type:   RecordTypeBLOB,
				Length: 0,
				Value:  nil,
			},
			want: []byte{
				0x03,       // Type = BLOB
				0x00, 0x00, // Length = 0
			},
		},
		{
			name: "int_record_length_eight",
			in: Record{
				Type:   RecordTypeINT,
				Length: 8,
				Value:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
			},
			want: []byte{
				0x02,       // Type = INT
				0x00, 0x08, // Length = 8
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A, // Value = 42 (BE int64)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			e := newEncoder(&buf)
			require.NoError(t, e.writeRecord(tc.in))
			require.Equal(t, tc.want, buf.Bytes())
		})
	}
}

func TestRecordRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   Record
	}{
		{
			name: "string",
			in: Record{
				Type:   RecordTypeSTRING,
				Length: 5,
				Value:  []byte("hello"),
			},
		},
		{
			name: "int",
			in: Record{
				Type:   RecordTypeINT,
				Length: 8,
				Value:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
			},
		},
		{
			name: "blob",
			in: Record{
				Type:   RecordTypeBLOB,
				Length: 4,
				Value:  []byte{0xDE, 0xAD, 0xBE, 0xEF},
			},
		},
		{
			name: "nested",
			in: Record{
				Type:   RecordTypeNESTED,
				Length: 3,
				Value:  []byte{0x01, 0x02, 0x03},
			},
		},
		{
			name: "empty",
			in: Record{
				Type:   RecordTypeBLOB,
				Length: 0,
				Value:  []byte{},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			e := newEncoder(&buf)
			require.NoError(t, e.writeRecord(tc.in))

			d := newDecoder(bytes.NewReader(buf.Bytes()))
			got, err := d.readRecord()
			require.NoError(t, err)

			require.Equal(t, tc.in.Type, got.Type)
			require.Equal(t, tc.in.Length, got.Length)
			require.Equal(t, tc.in.Value, got.Value)
		})
	}
}
