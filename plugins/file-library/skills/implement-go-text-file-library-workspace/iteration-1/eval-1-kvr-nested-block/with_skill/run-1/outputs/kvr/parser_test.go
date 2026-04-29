package kvr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Parallel()

	t.Run("empty_input_yields_zero_file", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(""))
		require.NoError(t, err)
		require.Equal(t, &File{}, got)
	})

	t.Run("single_string_record", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`record string GREETING = "hello"`))
		require.NoError(t, err)
		require.Len(t, got.Records, 1)
		require.Empty(t, got.Blocks)
		rec := got.Records[0]
		require.Equal(t, "string", rec.Type)
		require.Equal(t, "GREETING", rec.Key)
		require.Equal(t, "hello", rec.Value)
	})

	t.Run("single_number_record", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`record number ANSWER = 42`))
		require.NoError(t, err)
		require.Len(t, got.Records, 1)
		rec := got.Records[0]
		require.Equal(t, "number", rec.Type)
		require.Equal(t, "ANSWER", rec.Key)
		require.Equal(t, "42", rec.Value)
	})

	t.Run("two_top_level_records", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader("record string A = \"1\"\nrecord number B = 2"))
		require.NoError(t, err)
		require.Len(t, got.Records, 2)
		require.Equal(t, "A", got.Records[0].Key)
		require.Equal(t, "B", got.Records[1].Key)
	})

	t.Run("empty_block", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`block COLORS { }`))
		require.NoError(t, err)
		require.Empty(t, got.Records)
		require.Len(t, got.Blocks, 1)
		blk := got.Blocks[0]
		require.Equal(t, "COLORS", blk.Name)
		require.Empty(t, blk.Records)
	})

	t.Run("block_with_one_record", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`block COLORS { record string RED = "ff0000"; }`))
		require.NoError(t, err)
		require.Len(t, got.Blocks, 1)
		blk := got.Blocks[0]
		require.Equal(t, "COLORS", blk.Name)
		require.Len(t, blk.Records, 1)
		require.Equal(t, "string", blk.Records[0].Type)
		require.Equal(t, "RED", blk.Records[0].Key)
		require.Equal(t, "ff0000", blk.Records[0].Value)
	})

	t.Run("block_with_two_records", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`block COLORS { record string RED = "ff0000"; record string BLUE = "0000ff"; }`))
		require.NoError(t, err)
		require.Len(t, got.Blocks, 1)
		blk := got.Blocks[0]
		require.Equal(t, "COLORS", blk.Name)
		require.Len(t, blk.Records, 2)
		require.Equal(t, "RED", blk.Records[0].Key)
		require.Equal(t, "BLUE", blk.Records[1].Key)
	})

	t.Run("top_level_record_then_block", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader("record string G = \"hi\"\nblock B { record number N = 1; }"))
		require.NoError(t, err)
		require.Len(t, got.Records, 1)
		require.Len(t, got.Blocks, 1)
		require.Equal(t, "G", got.Records[0].Key)
		require.Equal(t, "B", got.Blocks[0].Name)
	})
}

func TestParserEquivalence(t *testing.T) {
	t.Parallel()

	// These pairs should produce equal *File values when parsed independently.
	testCases := []struct {
		name string
		a    string
		b    string
	}{
		{
			name: "whitespace_in_block_does_not_matter",
			a:    `block X { record string A = "1"; }`,
			b:    "block X {\n    record string A = \"1\";\n}",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a, err := Parse(strings.NewReader(tc.a))
			require.NoError(t, err)
			b, err := Parse(strings.NewReader(tc.b))
			require.NoError(t, err)
			require.Equal(t, a, b)
		})
	}
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	t.Run("record_missing_equals", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`record string A "1"`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
		require.Contains(t, ute.Want, TokenSymbol)
	})

	t.Run("block_missing_trailing_semicolon", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`block X { record string A = "1" }`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
	})

	t.Run("unknown_top_level_keyword", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`module X`))
		require.Error(t, err)
	})

	t.Run("bad_type_name", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`record bogus A = "1"`))
		require.Error(t, err)
	})

	t.Run("type_mismatch_string_with_number_value", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`record string A = 42`))
		var tme *TypeMismatchError
		require.ErrorAs(t, err, &tme)
	})
}
