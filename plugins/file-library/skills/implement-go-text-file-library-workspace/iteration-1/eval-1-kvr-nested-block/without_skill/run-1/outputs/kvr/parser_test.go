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
					{Type: "string", Key: "GREETING", Value: StringValue{V: "hello"}},
				},
			},
		},
		{
			name:  "single_number_record",
			input: `record number ANSWER = 42`,
			want: &File{
				Records: []Record{
					{Type: "number", Key: "ANSWER", Value: NumberValue{V: "42"}},
				},
			},
		},
		{
			name: "two_records_with_leading_comment_on_first",
			input: "# the universal greeting\n" +
				"record string GREETING = \"hello\"\n" +
				"record number ANSWER = 42\n",
			want: &File{
				Records: []Record{
					{
						LeadingComments: []string{"the universal greeting"},
						Type:            "string",
						Key:             "GREETING",
						Value:           StringValue{V: "hello"},
					},
					{Type: "number", Key: "ANSWER", Value: NumberValue{V: "42"}},
				},
			},
		},
		{
			name:  "empty_block",
			input: `block COLORS { }`,
			want: &File{
				Blocks: []Block{
					{Name: "COLORS"},
				},
			},
		},
		{
			name:  "block_with_one_record",
			input: `block COLORS { record string RED = "ff0000"; }`,
			want: &File{
				Blocks: []Block{
					{
						Name: "COLORS",
						Records: []Record{
							{Type: "string", Key: "RED", Value: StringValue{V: "ff0000"}},
						},
					},
				},
			},
		},
		{
			name: "block_with_two_records",
			input: "block COLORS {\n" +
				"    record string RED  = \"ff0000\";\n" +
				"    record string BLUE = \"0000ff\";\n" +
				"}\n",
			want: &File{
				Blocks: []Block{
					{
						Name: "COLORS",
						Records: []Record{
							{Type: "string", Key: "RED", Value: StringValue{V: "ff0000"}},
							{Type: "string", Key: "BLUE", Value: StringValue{V: "0000ff"}},
						},
					},
				},
			},
		},
		{
			name: "block_with_inner_leading_comment",
			input: "block COLORS {\n" +
				"    # primary red\n" +
				"    record string RED = \"ff0000\";\n" +
				"    record string BLUE = \"0000ff\";\n" +
				"}\n",
			want: &File{
				Blocks: []Block{
					{
						Name: "COLORS",
						Records: []Record{
							{
								LeadingComments: []string{"primary red"},
								Type:            "string",
								Key:             "RED",
								Value:           StringValue{V: "ff0000"},
							},
							{Type: "string", Key: "BLUE", Value: StringValue{V: "0000ff"}},
						},
					},
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

	t.Run("type_mismatch_string_with_number_value", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record string K = 42`))
		require.Error(t, err)
		var tme *TypeMismatchError
		require.ErrorAs(t, err, &tme)
	})

	t.Run("missing_trailing_separator_in_block", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`block X { record string A = "1" }`))
		require.Error(t, err)
	})

	t.Run("unknown_top_level_keyword", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`widget FOO = 1`))
		require.Error(t, err)
		var uke *UnexpectedKeywordError
		require.ErrorAs(t, err, &uke)
	})
}
