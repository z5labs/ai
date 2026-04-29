package tlv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
		want []byte
	}{
		{
			name: "typical_string_record",
			in: &File{Records: []Record{
				{Type: RecordTypeString, Length: 5, Value: []byte("hello")},
			}},
			want: []byte{0x01, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f},
		},
		{
			name: "empty_record_length_zero",
			in: &File{Records: []Record{
				{Type: RecordTypeBlob, Length: 0, Value: nil},
			}},
			want: []byte{0x03, 0x00, 0x00},
		},
		{
			name: "int_record_length_eight",
			in: &File{Records: []Record{
				{Type: RecordTypeInt, Length: 8, Value: []byte{0, 0, 0, 0, 0, 0, 0, 0x2a}},
			}},
			want: []byte{0x02, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, tc.in))
			require.Equal(t, tc.want, buf.Bytes())
		})
	}
}

func TestRecordRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
	}{
		{
			name: "string",
			in: &File{Records: []Record{
				{Type: RecordTypeString, Length: 5, Value: []byte("hello")},
			}},
		},
		{
			name: "int",
			in: &File{Records: []Record{
				{Type: RecordTypeInt, Length: 8, Value: []byte{0, 0, 0, 0, 0, 0, 0, 0x2a}},
			}},
		},
		{
			name: "blob",
			in: &File{Records: []Record{
				{Type: RecordTypeBlob, Length: 4, Value: []byte{0xde, 0xad, 0xbe, 0xef}},
			}},
		},
		{
			name: "nested",
			in: &File{Records: []Record{
				// NESTED Value is itself a TLV1 file's bytes; treat as opaque here.
				{Type: RecordTypeNested, Length: 3, Value: []byte{0x01, 0x02, 0x03}},
			}},
		},
		{
			name: "multiple_mixed",
			in: &File{Records: []Record{
				{Type: RecordTypeString, Length: 3, Value: []byte("foo")},
				{Type: RecordTypeBlob, Length: 0, Value: []byte{}},
				{Type: RecordTypeInt, Length: 8, Value: []byte{0, 0, 0, 0, 0, 0, 0, 0x2a}},
			}},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, tc.in))

			decoded, err := Decode(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.in, decoded)
		})
	}
}
