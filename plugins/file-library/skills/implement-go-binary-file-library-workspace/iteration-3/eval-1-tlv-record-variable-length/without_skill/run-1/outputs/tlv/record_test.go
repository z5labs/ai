package tlv

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecordTypeConstants(t *testing.T) {
	t.Parallel()

	require.Equal(t, RecordType(0x01), RecordTypeSTRING)
	require.Equal(t, RecordType(0x02), RecordTypeINT)
	require.Equal(t, RecordType(0x03), RecordTypeBLOB)
	require.Equal(t, RecordType(0x04), RecordTypeNESTED)
}

func TestFileHoldsRecords(t *testing.T) {
	t.Parallel()

	f := &File{Records: []Record{{Type: RecordTypeSTRING, Length: 0, Value: nil}}}
	require.Len(t, f.Records, 1)
	require.Equal(t, RecordTypeSTRING, f.Records[0].Type)
}

func TestDecodeRecord(t *testing.T) {
	t.Parallel()

	t.Run("typical STRING record", func(t *testing.T) {
		t.Parallel()

		// Type=0x01, Length=0x0005, Value="hello"
		raw := []byte{0x01, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'}

		rec, err := DecodeRecord(bytes.NewReader(raw))
		require.NoError(t, err)
		require.Equal(t, RecordTypeSTRING, rec.Type)
		require.Equal(t, uint16(5), rec.Length)
		require.Equal(t, []byte("hello"), rec.Value)
	})

	t.Run("empty record (Length=0)", func(t *testing.T) {
		t.Parallel()

		// Type=0x03 (BLOB), Length=0x0000, no Value bytes
		raw := []byte{0x03, 0x00, 0x00}

		rec, err := DecodeRecord(bytes.NewReader(raw))
		require.NoError(t, err)
		require.Equal(t, RecordTypeBLOB, rec.Type)
		require.Equal(t, uint16(0), rec.Length)
		require.Empty(t, rec.Value)
	})

	t.Run("INT record (Length=8)", func(t *testing.T) {
		t.Parallel()

		// Type=0x02 (INT), Length=0x0008, Value = int64(42) big-endian
		var valueBuf [8]byte
		binary.BigEndian.PutUint64(valueBuf[:], 42)
		raw := append([]byte{0x02, 0x00, 0x08}, valueBuf[:]...)

		rec, err := DecodeRecord(bytes.NewReader(raw))
		require.NoError(t, err)
		require.Equal(t, RecordTypeINT, rec.Type)
		require.Equal(t, uint16(8), rec.Length)
		require.Len(t, rec.Value, 8)
		require.Equal(t, uint64(42), binary.BigEndian.Uint64(rec.Value))
	})
}

func TestEncodeRecord(t *testing.T) {
	t.Parallel()

	t.Run("typical STRING record", func(t *testing.T) {
		t.Parallel()

		rec := Record{Type: RecordTypeSTRING, Length: 5, Value: []byte("hello")}

		var buf bytes.Buffer
		require.NoError(t, EncodeRecord(&buf, rec))
		require.Equal(t,
			[]byte{0x01, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'},
			buf.Bytes(),
		)
	})

	t.Run("empty record (Length=0)", func(t *testing.T) {
		t.Parallel()

		rec := Record{Type: RecordTypeBLOB, Length: 0, Value: nil}

		var buf bytes.Buffer
		require.NoError(t, EncodeRecord(&buf, rec))
		require.Equal(t, []byte{0x03, 0x00, 0x00}, buf.Bytes())
	})

	t.Run("INT record (Length=8)", func(t *testing.T) {
		t.Parallel()

		var valueBuf [8]byte
		binary.BigEndian.PutUint64(valueBuf[:], 42)
		rec := Record{Type: RecordTypeINT, Length: 8, Value: valueBuf[:]}

		var buf bytes.Buffer
		require.NoError(t, EncodeRecord(&buf, rec))

		want := append([]byte{0x02, 0x00, 0x08}, valueBuf[:]...)
		require.Equal(t, want, buf.Bytes())
	})
}

func TestRecordRoundTrip(t *testing.T) {
	t.Parallel()

	var intValue [8]byte
	binary.BigEndian.PutUint64(intValue[:], 0xDEADBEEFCAFEBABE)

	cases := []struct {
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
				Value:  intValue[:],
			},
		},
		{
			name: "BLOB",
			rec: Record{
				Type:   RecordTypeBLOB,
				Length: 4,
				Value:  []byte{0x00, 0xFF, 0x10, 0x20},
			},
		},
		{
			name: "NESTED",
			rec: Record{
				Type:   RecordTypeNESTED,
				Length: 3,
				Value:  []byte{0x54, 0x4C, 0x56}, // partial "TLV" prefix; opaque to Record layer
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, EncodeRecord(&buf, tc.rec))

			got, err := DecodeRecord(&buf)
			require.NoError(t, err)
			require.Equal(t, tc.rec.Type, got.Type)
			require.Equal(t, tc.rec.Length, got.Length)
			require.Equal(t, tc.rec.Value, got.Value)
		})
	}
}
