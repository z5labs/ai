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
