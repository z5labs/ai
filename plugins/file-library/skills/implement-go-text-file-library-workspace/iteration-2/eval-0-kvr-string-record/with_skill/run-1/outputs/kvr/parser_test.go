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

	t.Run("single_string_record_parses", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`record string GREETING = "hello"`))
		require.NoError(t, err)
		require.Len(t, got.Records, 1)
		require.Equal(t, "string", got.Records[0].Type)
		require.Equal(t, "GREETING", got.Records[0].Key)
		require.Equal(t, "hello", got.Records[0].Value)
	})

	t.Run("two_string_records_parse", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader("record string A = \"x\"\nrecord string B = \"y\""))
		require.NoError(t, err)
		require.Len(t, got.Records, 2)
		require.Equal(t, "string", got.Records[0].Type)
		require.Equal(t, "A", got.Records[0].Key)
		require.Equal(t, "x", got.Records[0].Value)
		require.Equal(t, "string", got.Records[1].Type)
		require.Equal(t, "B", got.Records[1].Key)
		require.Equal(t, "y", got.Records[1].Value)
	})

	t.Run("string_record_with_escapes_parses", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`record string MSG = "a\"b\nc"`))
		require.NoError(t, err)
		require.Len(t, got.Records, 1)
		require.Equal(t, "MSG", got.Records[0].Key)
		require.Equal(t, "a\"b\nc", got.Records[0].Value)
	})

	t.Run("equivalent_sources_yield_equal_files", func(t *testing.T) {
		t.Parallel()

		canonical, err := Parse(strings.NewReader(`record string GREETING = "hello"`))
		require.NoError(t, err)

		spaced, err := Parse(strings.NewReader("record   string\tGREETING\n=\n\"hello\""))
		require.NoError(t, err)

		require.Equal(t, canonical, spaced)
	})
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	t.Run("unexpected_token_when_value_is_brace", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`record string GREETING = }`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, TokenSymbol, ute.Got.Type)
		require.Contains(t, ute.Want, TokenString)
	})

	t.Run("unexpected_end_of_tokens_after_equals", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`record string GREETING =`))
		var eot *UnexpectedEndOfTokensError
		require.ErrorAs(t, err, &eot)
	})

	t.Run("unexpected_keyword_when_not_record", func(t *testing.T) {
		t.Parallel()

		_, err := Parse(strings.NewReader(`block FOO`))
		var uke *UnexpectedKeywordError
		require.ErrorAs(t, err, &uke)
		require.Equal(t, "block", uke.Got.Value)
		require.Contains(t, uke.Want, "record")
	})
}
