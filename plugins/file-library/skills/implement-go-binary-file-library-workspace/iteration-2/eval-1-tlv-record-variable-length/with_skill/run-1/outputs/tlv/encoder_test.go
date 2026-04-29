package tlv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeStringRecord(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{Records: []Record{{
		Type:   RecordTypeSTRING,
		Length: 5,
		Value:  []byte("hello"),
	}}})
	require.NoError(t, err)

	want := []byte{
		0x01,
		0x00, 0x05,
		0x68, 0x65, 0x6C, 0x6C, 0x6F,
	}
	require.Equal(t, want, buf.Bytes())
}

func TestEncodeEmptyRecord(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{Records: []Record{{
		Type:   RecordTypeBLOB,
		Length: 0,
		Value:  nil,
	}}})
	require.NoError(t, err)

	require.Equal(t, []byte{0x03, 0x00, 0x00}, buf.Bytes())
}

func TestEncodeIntRecord(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{Records: []Record{{
		Type:   RecordTypeINT,
		Length: 8,
		Value:  []byte{0, 0, 0, 0, 0, 0, 0, 0x2A},
	}}})
	require.NoError(t, err)

	want := []byte{
		0x02,
		0x00, 0x08,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A,
	}
	require.Equal(t, want, buf.Bytes())
}

func TestEncodeNoRecords(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{})
	require.NoError(t, err)
	require.Empty(t, buf.Bytes())
}

func TestRoundTripPerRecordType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		rec  Record
	}{
		{
			name: "string",
			rec:  Record{Type: RecordTypeSTRING, Length: 5, Value: []byte("hello")},
		},
		{
			name: "int",
			rec:  Record{Type: RecordTypeINT, Length: 8, Value: []byte{0, 0, 0, 0, 0, 0, 0, 0x2A}},
		},
		{
			name: "blob",
			rec:  Record{Type: RecordTypeBLOB, Length: 4, Value: []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		},
		{
			name: "nested",
			rec:  Record{Type: RecordTypeNESTED, Length: 3, Value: []byte{0x01, 0x00, 0x00}},
		},
		{
			name: "empty",
			rec:  Record{Type: RecordTypeBLOB, Length: 0, Value: nil},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			original := &File{Records: []Record{tc.rec}}

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, original))

			decoded, err := Decode(&buf)
			require.NoError(t, err)
			require.Len(t, decoded.Records, 1)
			require.Equal(t, tc.rec.Type, decoded.Records[0].Type)
			require.Equal(t, tc.rec.Length, decoded.Records[0].Length)
			if tc.rec.Length == 0 {
				require.Empty(t, decoded.Records[0].Value)
			} else {
				require.Equal(t, tc.rec.Value, decoded.Records[0].Value)
			}
		})
	}
}
