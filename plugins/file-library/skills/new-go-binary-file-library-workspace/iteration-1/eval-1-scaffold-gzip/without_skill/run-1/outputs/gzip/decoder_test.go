package gzip_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"example.com/gzip-eval/gzip"
)

func TestDecoder_Decode_NotImplemented(t *testing.T) {
	t.Parallel()

	dec := gzip.NewDecoder(bytes.NewReader(nil))
	f, err := dec.Decode()

	require.Nil(t, f)
	require.Error(t, err)
}

func TestDecoder_DecodeHeader_NotImplemented(t *testing.T) {
	t.Parallel()

	dec := gzip.NewDecoder(bytes.NewReader(nil))
	h, err := dec.DecodeHeader()

	require.Nil(t, h)
	require.Error(t, err)
}

func TestDecoder_DecodeTrailer_NotImplemented(t *testing.T) {
	t.Parallel()

	dec := gzip.NewDecoder(bytes.NewReader(nil))
	tr, err := dec.DecodeTrailer()

	require.Nil(t, tr)
	require.Error(t, err)
}
