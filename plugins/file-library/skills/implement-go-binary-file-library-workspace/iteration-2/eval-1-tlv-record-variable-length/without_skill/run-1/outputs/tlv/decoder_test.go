package tlv

import (
	"bytes"
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

	tests := []struct {
		name string
		in   []byte
		want Record
	}{
		{
			name: "typical STRING record",
			in:   []byte{0x01, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'},
			want: Record{
				Type:   RecordTypeSTRING,
				Length: 5,
				Value:  []byte("hello"),
			},
		},
		{
			name: "empty record (Length=0)",
			in:   []byte{0x03, 0x00, 0x00},
			want: Record{
				Type:   RecordTypeBLOB,
				Length: 0,
				Value:  nil,
			},
		},
		{
			name: "INT record (Length=8)",
			in:   []byte{0x02, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
			want: Record{
				Type:   RecordTypeINT,
				Length: 8,
				Value:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2A},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := newDecoder(bytes.NewReader(tt.in))
			got, err := d.readRecord()
			require.NoError(t, err)
			require.Equal(t, tt.want.Type, got.Type)
			require.Equal(t, tt.want.Length, got.Length)
			require.Equal(t, tt.want.Value, got.Value)
		})
	}
}

func TestReadRecordUnknownType(t *testing.T) {
	t.Parallel()

	d := newDecoder(bytes.NewReader([]byte{0xFF, 0x00, 0x00}))
	_, err := d.readRecord()
	require.ErrorIs(t, err, ErrUnknownRecordType)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Record.Type", fe.Field)
}

func TestReadRecordIntWrongLength(t *testing.T) {
	t.Parallel()

	d := newDecoder(bytes.NewReader([]byte{0x02, 0x00, 0x04, 0x00, 0x00, 0x00, 0x2A}))
	_, err := d.readRecord()
	require.ErrorIs(t, err, ErrInvalid)
}
