package gzip_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"example.com/gzip-eval/gzip"
)

func TestEncoder_Encode_NotImplemented(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	enc := gzip.NewEncoder(&buf)

	err := enc.Encode(&gzip.File{})

	require.Error(t, err)
}

func TestEncoder_EncodeHeader_NotImplemented(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	enc := gzip.NewEncoder(&buf)

	err := enc.EncodeHeader(&gzip.Header{})

	require.Error(t, err)
}

func TestEncoder_EncodeTrailer_NotImplemented(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	enc := gzip.NewEncoder(&buf)

	err := enc.EncodeTrailer(&gzip.Trailer{})

	require.Error(t, err)
}
