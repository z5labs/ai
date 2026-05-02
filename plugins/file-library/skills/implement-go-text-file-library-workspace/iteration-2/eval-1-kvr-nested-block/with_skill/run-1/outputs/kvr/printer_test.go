package kvr

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrinter(t *testing.T) {
	t.Parallel()

	t.Run("empty_file_prints_empty_string", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		err := Print(&buf, &File{})
		require.NoError(t, err)
		require.Equal(t, "", buf.String())
	})

	t.Run("top_level_string_record", func(t *testing.T) {
		t.Parallel()

		f, err := Parse(strings.NewReader(`record string GREETING = "hello"`))
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		require.Equal(t, "record string GREETING = \"hello\"\n", buf.String())
	})

	t.Run("empty_block", func(t *testing.T) {
		t.Parallel()

		f, err := Parse(strings.NewReader(`block COLORS { }`))
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		require.Equal(t, "block COLORS {\n}\n", buf.String())
	})

	t.Run("block_with_one_record", func(t *testing.T) {
		t.Parallel()

		f, err := Parse(strings.NewReader(`block COLORS { record string RED = "ff0000"; }`))
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		require.Equal(t,
			"block COLORS {\n    record string RED = \"ff0000\";\n}\n",
			buf.String())
	})

	t.Run("block_with_two_records", func(t *testing.T) {
		t.Parallel()

		f, err := Parse(strings.NewReader(
			`block COLORS { record string RED = "ff0000"; record string BLUE = "0000ff"; }`,
		))
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		require.Equal(t,
			"block COLORS {\n"+
				"    record string RED = \"ff0000\";\n"+
				"    record string BLUE = \"0000ff\";\n"+
				"}\n",
			buf.String())
	})
}

func TestPrinterRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		source string
	}{
		{
			name:   "empty_source_round_trips",
			source: "",
		},
		{
			name:   "single_top_level_string_record",
			source: `record string GREETING = "hello"`,
		},
		{
			name:   "empty_block",
			source: `block COLORS { }`,
		},
		{
			name:   "block_with_one_record",
			source: `block COLORS { record string RED = "ff0000"; }`,
		},
		{
			name:   "block_with_two_records",
			source: `block COLORS { record string RED = "ff0000"; record string BLUE = "0000ff"; }`,
		},
		{
			name:   "block_then_top_level_record",
			source: `block X { record string A = "1"; } record string B = "2"`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			first, err := Parse(strings.NewReader(tc.source))
			require.NoError(t, err)

			var buf bytes.Buffer
			require.NoError(t, Print(&buf, first))

			second, err := Parse(&buf)
			require.NoError(t, err)
			// Positions are not preserved across a round-trip per SPEC.md
			// Semantics: "blank lines, trailing whitespace, and the exact
			// column of a token are not preserved across a round-trip."
			// Compare structural content only.
			clearPositions(first)
			clearPositions(second)
			require.Equal(t, first, second)
		})
	}
}

// clearPositions resets every Pos in the file to the zero value so that
// round-trip tests can assert on structural equality alone (per SPEC.md
// "Whitespace fidelity").
func clearPositions(f *File) {
	for i := range f.Records {
		f.Records[i].Pos = Pos{}
	}
	for i := range f.Blocks {
		f.Blocks[i].Pos = Pos{}
		for j := range f.Blocks[i].Records {
			f.Blocks[i].Records[j].Pos = Pos{}
		}
	}
}
