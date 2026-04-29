package tlv

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input []byte
		want  []Record
	}{
		{
			name: "typical_string_record",
			input: []byte{
				// Type=STRING(0x01), Length=5 (big-endian), Value="hello"
				0x01, 0x00, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f,
			},
			want: []Record{
				{Type: RecordTypeString, Length: 5, Value: []byte("hello")},
			},
		},
		{
			name: "empty_record_length_zero",
			input: []byte{
				// Type=BLOB(0x03), Length=0, no Value
				0x03, 0x00, 0x00,
			},
			want: []Record{
				{Type: RecordTypeBlob, Length: 0, Value: []byte{}},
			},
		},
		{
			name: "int_record_length_eight",
			input: []byte{
				// Type=INT(0x02), Length=8, Value=42 (int64 big-endian)
				0x02, 0x00, 0x08,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a,
			},
			want: []Record{
				{Type: RecordTypeInt, Length: 8, Value: []byte{0, 0, 0, 0, 0, 0, 0, 0x2a}},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f, err := Decode(bytes.NewReader(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.want, f.Records)
		})
	}
}

func TestDecodeRecordTruncated(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     []byte
		wantField string
	}{
		{
			// Length says 5 but only 2 value bytes are present.
			name:      "truncated_value",
			input:     []byte{0x01, 0x00, 0x05, 0x68, 0x65},
			wantField: "Record.Value",
		},
		{
			// Header byte arrives but length prefix is cut short.
			name:      "truncated_length",
			input:     []byte{0x01, 0x00},
			wantField: "Record.Length",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Decode(bytes.NewReader(tc.input))
			require.Error(t, err)
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)

			var fe *FieldError
			require.ErrorAs(t, err, &fe)
			require.Equal(t, tc.wantField, fe.Field)

			var oe *OffsetError
			require.ErrorAs(t, err, &oe)
			require.Greater(t, oe.Offset, int64(0))
		})
	}
}
