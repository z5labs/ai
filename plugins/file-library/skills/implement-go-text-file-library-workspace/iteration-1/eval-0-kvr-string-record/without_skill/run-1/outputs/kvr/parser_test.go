package kvr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  *File
	}{
		{
			name:  "empty_input_yields_zero_file",
			input: "",
			want:  &File{},
		},
		{
			name:  "single_string_record",
			input: `record string GREETING = "hello"`,
			want: &File{
				Records: []Record{
					{Type: "string", Key: "GREETING", Value: "hello"},
				},
			},
		},
		{
			name:  "two_string_records_on_separate_lines",
			input: "record string A = \"1\"\nrecord string B = \"2\"",
			want: &File{
				Records: []Record{
					{Type: "string", Key: "A", Value: "1"},
					{Type: "string", Key: "B", Value: "2"},
				},
			},
		},
		{
			name:  "string_record_with_escaped_value",
			input: `record string Q = "a\"b"`,
			want: &File{
				Records: []Record{
					{Type: "string", Key: "Q", Value: `a"b`},
				},
			},
		},
		{
			name:  "string_record_with_newline_escape",
			input: `record string MSG = "a\nb"`,
			want: &File{
				Records: []Record{
					{Type: "string", Key: "MSG", Value: "a\nb"},
				},
			},
		},
		{
			name:  "extra_whitespace_between_tokens",
			input: "  record   string   K   =   \"v\"  ",
			want: &File{
				Records: []Record{
					{Type: "string", Key: "K", Value: "v"},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(strings.NewReader(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	t.Run("missing_value", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("record string K ="))
		require.Error(t, err)
		var eot *UnexpectedEndOfTokensError
		require.ErrorAs(t, err, &eot)
	})

	t.Run("non_string_top_level_token", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`"oops"`))
		require.Error(t, err)
		var ut *UnexpectedTokenError
		require.ErrorAs(t, err, &ut)
	})

	t.Run("unknown_keyword", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("notrecord string K = \"v\""))
		require.Error(t, err)
		var uk *UnexpectedKeywordError
		require.ErrorAs(t, err, &uk)
	})

	t.Run("wrong_record_type_keyword", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("record bogus K = \"v\""))
		require.Error(t, err)
		var uk *UnexpectedKeywordError
		require.ErrorAs(t, err, &uk)
	})

	t.Run("missing_equals", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record string K "v"`))
		require.Error(t, err)
		var ut *UnexpectedTokenError
		require.ErrorAs(t, err, &ut)
	})
}
