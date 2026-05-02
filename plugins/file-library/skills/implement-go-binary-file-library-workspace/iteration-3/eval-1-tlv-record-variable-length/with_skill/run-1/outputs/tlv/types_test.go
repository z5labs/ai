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
		{"string", RecordTypeSTRING, "STRING"},
		{"int", RecordTypeINT, "INT"},
		{"blob", RecordTypeBLOB, "BLOB"},
		{"nested", RecordTypeNESTED, "NESTED"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.want, tc.v.String())
		})
	}
}

func TestRecordTypeStringUnknown(t *testing.T) {
	t.Parallel()

	require.Equal(t, "RecordType(0xff)", RecordType(0xff).String())
}

func TestRecordTypeConstants(t *testing.T) {
	t.Parallel()

	require.Equal(t, RecordType(0x01), RecordTypeSTRING)
	require.Equal(t, RecordType(0x02), RecordTypeINT)
	require.Equal(t, RecordType(0x03), RecordTypeBLOB)
	require.Equal(t, RecordType(0x04), RecordTypeNESTED)
}

func TestFileHasRecordsSlice(t *testing.T) {
	t.Parallel()

	// File must hold a slice of Records.
	f := &File{Records: []Record{
		{Type: RecordTypeSTRING, Length: 5, Value: []byte("hello")},
	}}
	require.Len(t, f.Records, 1)
	require.Equal(t, RecordTypeSTRING, f.Records[0].Type)
}
