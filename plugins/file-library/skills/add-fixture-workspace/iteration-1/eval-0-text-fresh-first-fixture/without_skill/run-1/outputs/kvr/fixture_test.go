package kvr

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFixtureRoundTrip reads each .kvr file under testdata/, parses it, prints
// the resulting AST, parses the printed output, and asserts the two ASTs are
// equal. This pins the Parse → Print → Parse round-trip property against
// real-world inputs (including those from production / customer reports).
func TestFixtureRoundTrip(t *testing.T) {
	t.Parallel()

	matches, err := filepath.Glob(filepath.Join("testdata", "*.kvr"))
	require.NoError(t, err)
	require.NotEmpty(t, matches, "no fixtures found under testdata/")

	for _, path := range matches {
		path := path
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			source, err := os.ReadFile(path)
			require.NoError(t, err)

			first, err := Parse(bytes.NewReader(source))
			require.NoError(t, err, "first parse of %s", path)

			var buf bytes.Buffer
			require.NoError(t, Print(&buf, first), "print of %s", path)

			second, err := Parse(&buf)
			require.NoError(t, err, "second parse of %s", path)

			require.Equal(t, first, second, "round-trip AST mismatch for %s", path)
		})
	}
}
