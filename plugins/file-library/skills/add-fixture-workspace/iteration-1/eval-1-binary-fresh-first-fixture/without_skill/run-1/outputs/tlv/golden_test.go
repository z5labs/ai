package tlv

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGoldenRoundTrip pins real-world TLV1 blobs in testdata/ as regression
// fixtures: each file is decoded then re-encoded, and the resulting bytes must
// be byte-for-byte identical to the original. This is a stronger guarantee
// than AST equality and the right shape for a binary format — it catches
// drift in either direction (decoder dropping bytes, encoder reordering or
// re-flagging fields, etc.).
//
// Add new fixtures by dropping the file into testdata/ and appending its
// basename to the table below.
func TestGoldenRoundTrip(t *testing.T) {
	t.Parallel()

	fixtures := []string{
		"sample.tlv",
	}

	for _, name := range fixtures {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join("testdata", name)
			want, err := os.ReadFile(path)
			require.NoError(t, err, "read fixture %s", path)

			file, err := Decode(bytes.NewReader(want))
			require.NoError(t, err, "Decode(%s)", name)

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, file), "Encode(%s)", name)

			require.Equal(t, want, buf.Bytes(),
				"round-trip of %s did not produce byte-identical output", name)
		})
	}
}
