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

func TestRecordTypeConstants(t *testing.T) {
	t.Parallel()

	require.Equal(t, RecordType(0x01), RecordTypeSTRING)
	require.Equal(t, RecordType(0x02), RecordTypeINT)
	require.Equal(t, RecordType(0x03), RecordTypeBLOB)
	require.Equal(t, RecordType(0x04), RecordTypeNESTED)
}
