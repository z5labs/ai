package gzip_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"example.com/gzip-eval/gzip"
)

func TestMagicBytes(t *testing.T) {
	t.Parallel()

	require.Equal(t, byte(0x1f), gzip.Magic1)
	require.Equal(t, byte(0x8b), gzip.Magic2)
}

func TestCompressionDeflate(t *testing.T) {
	t.Parallel()

	require.Equal(t, byte(8), gzip.CompressionDeflate)
}

func TestFlagBitsAreDistinct(t *testing.T) {
	t.Parallel()

	flags := []byte{
		gzip.FlagText,
		gzip.FlagHCRC,
		gzip.FlagExtra,
		gzip.FlagName,
		gzip.FlagComment,
	}

	seen := map[byte]bool{}
	for _, f := range flags {
		require.False(t, seen[f], "duplicate flag bit %#x", f)
		seen[f] = true
	}
}
