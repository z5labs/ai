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
		require.Equal(t, "string", got.Records[0].Type)
		require.Equal(t, "GREETING", got.Records[0].Key)
		require.Equal(t, "hello", got.Records[0].Value)
	})

	t.Run("two_string_records", func(t *testing.T) {
		t.Parallel()
		input := "record string A = \"1\"\nrecord string B = \"2\""
		got, err := Parse(strings.NewReader(input))
		require.NoError(t, err)
		require.Len(t, got.Records, 2)
		require.Equal(t, "A", got.Records[0].Key)
		require.Equal(t, "1", got.Records[0].Value)
		require.Equal(t, "B", got.Records[1].Key)
		require.Equal(t, "2", got.Records[1].Value)
	})

	t.Run("string_with_escaped_quote", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader(`record string K = "a\"b"`))
		require.NoError(t, err)
		require.Len(t, got.Records, 1)
		require.Equal(t, `a"b`, got.Records[0].Value)
	})

	t.Run("equivalent_inputs_produce_equal_files", func(t *testing.T) {
		t.Parallel()
		// Different surface whitespace, same logical AST.
		canonical, err := Parse(strings.NewReader(`record string K = "v"`))
		require.NoError(t, err)
		spaced, err := Parse(strings.NewReader("   record   string   K   =   \"v\"   "))
		require.NoError(t, err)
		require.Equal(t, canonical, spaced)
	})
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing_value_after_equals", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("record string K ="))
		var ee *UnexpectedEndOfTokensError
		require.ErrorAs(t, err, &ee)
	})

	t.Run("missing_equals", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record string K "v"`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, TokenString, ute.Got.Type)
		require.Contains(t, ute.Want, TokenSymbol)
	})

	t.Run("unknown_top_level_keyword", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`widget string K = "v"`))
		var uki *UnknownKeywordError
		require.ErrorAs(t, err, &uki)
		require.Equal(t, "widget", uki.Got)
	})

	t.Run("unknown_record_type", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record bool K = "v"`))
		var ute *UnknownTypeError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, "bool", ute.Got)
	})
}
