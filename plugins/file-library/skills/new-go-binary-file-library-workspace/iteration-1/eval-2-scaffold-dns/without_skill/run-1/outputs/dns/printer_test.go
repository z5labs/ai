package dns

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrinter_Print_ZeroMessageWritesNothing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, Message{})
	require.NoError(t, err)
	require.Equal(t, 0, buf.Len())
}
