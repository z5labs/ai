package dns

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser_Parse_EmptyInputReturnsZeroMessage(t *testing.T) {
	t.Parallel()

	msg, err := Decode(bytes.NewReader(nil))
	require.NoError(t, err)
	require.Equal(t, Message{}, msg)
}
