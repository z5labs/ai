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
			file: &File{Statements: []Type{
				&Record{Type: "string", Key: "GREETING", Value: "hello", ValueKind: TokenString},
			}},
			want: "record string GREETING = \"hello\"\n",
		},
		{
			name: "single_number_record",
			file: &File{Statements: []Type{
				&Record{Type: "number", Key: "ANSWER", Value: "42", ValueKind: TokenNumber},
			}},
			want: "record number ANSWER = 42\n",
		},
		{
			name: "record_with_one_leading_comment",
			file: &File{Statements: []Type{
				&Record{
					LeadingComments: []string{"greet"},
					Type:            "string", Key: "A", Value: "1", ValueKind: TokenString,
				},
			}},
			want: "# greet\nrecord string A = \"1\"\n",
		},
		{
			name: "record_with_two_leading_comments",
			file: &File{Statements: []Type{
				&Record{
					LeadingComments: []string{"first", "second"},
					Type:            "string", Key: "A", Value: "1", ValueKind: TokenString,
				},
			}},
			want: "# first\n# second\nrecord string A = \"1\"\n",
		},
		{
			name: "two_records",
			file: &File{Statements: []Type{
				&Record{Type: "string", Key: "A", Value: "a", ValueKind: TokenString},
				&Record{Type: "number", Key: "B", Value: "2", ValueKind: TokenNumber},
			}},
			want: "record string A = \"a\"\nrecord number B = 2\n",
		},
		{
			name: "string_value_with_escape",
			file: &File{Statements: []Type{
				&Record{Type: "string", Key: "K", Value: `a"b`, ValueKind: TokenString},
			}},
			want: "record string K = \"a\\\"b\"\n",
		},
		{
			name: "block_with_one_record",
			file: &File{Statements: []Type{
				&Block{
					Name: "C",
					Records: []Record{
						{Type: "string", Key: "A", Value: "1", ValueKind: TokenString},
					},
				},
			}},
			want: "block C {\n    record string A = \"1\";\n}\n",
		},
		{
			name: "block_with_leading_comment",
			file: &File{Statements: []Type{
				&Block{
					LeadingComments: []string{"colors"},
					Name:            "C",
					Records: []Record{
						{Type: "string", Key: "A", Value: "1", ValueKind: TokenString},
					},
				},
			}},
			want: "# colors\nblock C {\n    record string A = \"1\";\n}\n",
		},
		{
			name: "block_with_inner_leading_comment",
			file: &File{Statements: []Type{
				&Block{
					Name: "C",
					Records: []Record{
						{
							LeadingComments: []string{"primary"},
							Type:            "string", Key: "A", Value: "1", ValueKind: TokenString,
						},
						{Type: "string", Key: "B", Value: "2", ValueKind: TokenString},
					},
				},
			}},
			want: "block C {\n    # primary\n    record string A = \"1\";\n    record string B = \"2\";\n}\n",
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
			name:   "single_string_record",
			source: `record string GREETING = "hello"`,
		},
		{
			name:   "single_number_record",
			source: `record number ANSWER = 42`,
		},
		{
			name:   "two_records",
			source: "record string A = \"1\"\nrecord number B = 2",
		},
		{
			name:   "comment_above_record",
			source: "# greet\nrecord string A = \"1\"",
		},
		{
			name:   "blank_line_then_comment_then_record",
			source: "\n\n# greet\nrecord string A = \"1\"",
		},
		{
			name:   "two_comments_above_record",
			source: "# first\n# second\nrecord string A = \"1\"",
		},
		{
			name:   "comments_attach_only_to_following",
			source: "# greet\nrecord string A = \"1\"\nrecord number B = 2",
		},
		{
			name:   "block_with_records",
			source: "block C { record string A = \"1\"; record string B = \"2\"; }",
		},
		{
			name:   "comment_above_block",
			source: "# colors\nblock C { record string A = \"1\"; }",
		},
		{
			name:   "comment_inside_block",
			source: "block C {\n# primary\nrecord string A = \"1\";\nrecord string B = \"2\";\n}",
		},
		{
			name:   "string_with_escape_quote",
			source: `record string K = "a\"b"`,
		},
		{
			name:   "string_with_escape_newline",
			source: `record string K = "line1\n"`,
		},
		{
			name:   "spec_example_round_trip",
			source: "# greet\nrecord string A = \"1\"\n\n# answer\nrecord number B = 2",
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
