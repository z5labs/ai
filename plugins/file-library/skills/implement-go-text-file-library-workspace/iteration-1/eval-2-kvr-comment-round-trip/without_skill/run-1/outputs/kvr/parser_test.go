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
			input: `record string A = "1"`,
			want: &File{
				Statements: []Statement{
					Record{Type: "string", Key: "A", Value: "1"},
				},
			},
		},
		{
			name:  "single_number_record",
			input: `record number N = 42`,
			want: &File{
				Statements: []Statement{
					Record{Type: "number", Key: "N", Value: "42"},
				},
			},
		},
		{
			name:  "two_records_back_to_back",
			input: "record string A = \"1\"\nrecord number B = 2",
			want: &File{
				Statements: []Statement{
					Record{Type: "string", Key: "A", Value: "1"},
					Record{Type: "number", Key: "B", Value: "2"},
				},
			},
		},
		{
			name:  "comment_above_record",
			input: "# greet\nrecord string A = \"1\"",
			want: &File{
				Statements: []Statement{
					Record{LeadingComments: []string{"greet"}, Type: "string", Key: "A", Value: "1"},
				},
			},
		},
		{
			name:  "blank_line_then_comment_then_record",
			input: "\n# greet\nrecord string A = \"1\"",
			want: &File{
				Statements: []Statement{
					Record{LeadingComments: []string{"greet"}, Type: "string", Key: "A", Value: "1"},
				},
			},
		},
		{
			name:  "two_comments_above_same_record",
			input: "# one\n# two\nrecord string A = \"1\"",
			want: &File{
				Statements: []Statement{
					Record{LeadingComments: []string{"one", "two"}, Type: "string", Key: "A", Value: "1"},
				},
			},
		},
		{
			name:  "comment_above_block",
			input: "# colors\nblock COLORS { record string RED = \"ff0000\"; }",
			want: &File{
				Statements: []Statement{
					Block{
						LeadingComments: []string{"colors"},
						Name:            "COLORS",
						Records: []Record{
							{Type: "string", Key: "RED", Value: "ff0000"},
						},
					},
				},
			},
		},
		{
			name:  "comment_inside_block_attaches_to_inner_record",
			input: "block COLORS {\n# primary red\nrecord string RED = \"ff0000\";\nrecord string BLUE = \"0000ff\";\n}",
			want: &File{
				Statements: []Statement{
					Block{
						Name: "COLORS",
						Records: []Record{
							{LeadingComments: []string{"primary red"}, Type: "string", Key: "RED", Value: "ff0000"},
							{Type: "string", Key: "BLUE", Value: "0000ff"},
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
		_, err := Parse(strings.NewReader("record string K = 42"))
		require.Error(t, err)
		var tm *TypeMismatchError
		require.ErrorAs(t, err, &tm)
	})

	t.Run("block_missing_trailing_semicolon", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("block X { record string A = \"1\" }"))
		require.Error(t, err)
		var ut *UnexpectedTokenError
		require.ErrorAs(t, err, &ut)
	})

	t.Run("unknown_top_level_keyword", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("weird thing"))
		require.Error(t, err)
	})
}
