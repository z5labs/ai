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

	tests := []struct {
		name string
		in   Record
		want []byte
	}{
		{
			name: "typical STRING record",
			in: Record{
				Type:   RecordTypeSTRING,
				Length: 5,
				Value:  []byte("hello"),
			},
			want: []byte{0x01, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'},
		},
		{
			name: "empty record (Length=0)",
			in: Record{
				Type:   RecordTypeBLOB,
				Length: 0,
				Value:  nil,
			},
			want: []byte{0x03, 0x00, 0x00},
		},
		{
			name: "INT record (Length=8)",
			in: Record{
				Type:   RecordTypeINT,
				Length: 8,
				Value:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
			},
			want: []byte{0x02, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			e := newEncoder(&buf)
			require.NoError(t, e.writeRecord(tt.in))
			require.Equal(t, tt.want, buf.Bytes())
		})
	}
}

func TestRecordRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rec  Record
	}{
		{
			name: "STRING",
			rec: Record{
				Type:   RecordTypeSTRING,
				Length: 5,
				Value:  []byte("hello"),
			},
		},
		{
			name: "INT",
			rec: Record{
				Type:   RecordTypeINT,
				Length: 8,
				Value:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
			},
		},
		{
			name: "BLOB",
			rec: Record{
				Type:   RecordTypeBLOB,
				Length: 4,
				Value:  []byte{0xDE, 0xAD, 0xBE, 0xEF},
			},
		},
		{
			name: "NESTED",
			rec: Record{
				Type:   RecordTypeNESTED,
				Length: 3,
				Value:  []byte{0x01, 0x02, 0x03},
			},
		},
		{
			name: "empty BLOB",
			rec: Record{
				Type:   RecordTypeBLOB,
				Length: 0,
				Value:  nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, newEncoder(&buf).writeRecord(tt.rec))

			got, err := newDecoder(bytes.NewReader(buf.Bytes())).readRecord()
			require.NoError(t, err)
			require.Equal(t, tt.rec.Type, got.Type)
			require.Equal(t, tt.rec.Length, got.Length)
			if tt.rec.Length == 0 {
				require.Empty(t, got.Value)
			} else {
				require.Equal(t, tt.rec.Value, got.Value)
			}
		})
	}
}

func TestWriteRecordLengthMismatch(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := newEncoder(&buf).writeRecord(Record{
		Type:   RecordTypeSTRING,
		Length: 10,
		Value:  []byte("hi"),
	})
	require.ErrorIs(t, err, ErrInvalid)
}
