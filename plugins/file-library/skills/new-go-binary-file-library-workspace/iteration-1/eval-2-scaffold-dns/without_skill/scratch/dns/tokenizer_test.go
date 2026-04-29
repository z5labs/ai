package dns

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenizer_Next_EmptyInputReturnsEOF(t *testing.T) {
	t.Parallel()

	tk := NewTokenizer(bytes.NewReader(nil))

	tok, err := tk.Next()
	require.ErrorIs(t, err, io.EOF)
	require.Equal(t, TokenEOF, tok.Kind)
}
