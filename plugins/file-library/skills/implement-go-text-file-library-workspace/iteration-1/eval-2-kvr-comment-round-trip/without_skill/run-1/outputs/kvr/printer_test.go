package kvr

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrinter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		file *File
		want string
	}{
		{
			name: "empty_file_prints_empty_string",
			file: &File{},
			want: "",
		},
		{
			name: "single_string_record",
			file: &File{
				Statements: []Statement{
					Record{Type: "string", Key: "A", Value: "1"},
				},
			},
			want: "record string A = \"1\"\n",
		},
		{
			name: "single_number_record",
			file: &File{
				Statements: []Statement{
					Record{Type: "number", Key: "N", Value: "42"},
				},
			},
			want: "record number N = 42\n",
		},
		{
			name: "record_with_one_leading_comment",
			file: &File{
				Statements: []Statement{
					Record{LeadingComments: []string{"greet"}, Type: "string", Key: "A", Value: "1"},
				},
			},
			want: "# greet\nrecord string A = \"1\"\n",
		},
		{
			name: "record_with_two_leading_comments",
			file: &File{
				Statements: []Statement{
					Record{LeadingComments: []string{"one", "two"}, Type: "string", Key: "A", Value: "1"},
				},
			},
			want: "# one\n# two\nrecord string A = \"1\"\n",
		},
		{
			name: "block_with_leading_comment",
			file: &File{
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
			want: "# colors\nblock COLORS {\n    record string RED = \"ff0000\";\n}\n",
		},
		{
			name: "block_with_inner_record_comment",
			file: &File{
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
			want: "block COLORS {\n    # primary red\n    record string RED = \"ff0000\";\n    record string BLUE = \"0000ff\";\n}\n",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Print(&buf, tc.file)
			require.NoError(t, err)
			require.Equal(t, tc.want, buf.String())
		})
	}
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
			name:   "single_record_round_trips",
			source: `record string A = "1"`,
		},
		{
			name:   "comment_above_record_round_trips",
			source: "# greet\nrecord string A = \"1\"",
		},
		{
			name:   "blank_line_then_comment_then_record_round_trips",
			source: "\n# greet\nrecord string A = \"1\"",
		},
		{
			name:   "two_comments_above_record_round_trips",
			source: "# one\n# two\nrecord string A = \"1\"",
		},
		{
			name:   "comment_above_block_round_trips",
			source: "# colors\nblock COLORS { record string RED = \"ff0000\"; }",
		},
		{
			name:   "comment_inside_block_round_trips",
			source: "block COLORS {\n# primary red\nrecord string RED = \"ff0000\";\nrecord string BLUE = \"0000ff\";\n}",
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
			require.Equal(t, first, second)
		})
	}
}
