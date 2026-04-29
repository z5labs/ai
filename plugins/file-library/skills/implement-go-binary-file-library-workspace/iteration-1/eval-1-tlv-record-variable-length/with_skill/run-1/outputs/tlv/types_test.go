package tlv

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorChain(t *testing.T) {
	t.Parallel()

	err := &FieldError{Field: "File", Err: &OffsetError{Offset: 0, Err: errUnimplemented}}

	require.ErrorIs(t, err, errUnimplemented)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "File", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(0), oe.Offset)

	require.True(t, errors.Is(err, errUnimplemented))
}

func TestRecordTypeString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		v    RecordType
		want string
	}{
		{"string", RecordTypeString, "STRING"},
		{"int", RecordTypeInt, "INT"},
		{"blob", RecordTypeBlob, "BLOB"},
		{"nested", RecordTypeNested, "NESTED"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, tc.v.String())
		})
	}
}

func TestRecordTypeConstants(t *testing.T) {
	t.Parallel()

	require.Equal(t, RecordType(0x01), RecordTypeString)
	require.Equal(t, RecordType(0x02), RecordTypeInt)
	require.Equal(t, RecordType(0x03), RecordTypeBlob)
	require.Equal(t, RecordType(0x04), RecordTypeNested)
}

func TestFileHoldsRecords(t *testing.T) {
	t.Parallel()

	f := &File{
		Records: []Record{
			{Type: RecordTypeString, Length: 3, Value: []byte("foo")},
		},
	}
	require.Len(t, f.Records, 1)
	require.Equal(t, RecordTypeString, f.Records[0].Type)
}
