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
			file: &File{Records: []Record{
				{Type: "string", Key: "GREETING", Value: "hello"},
			}},
			want: "record string GREETING = \"hello\"\n",
		},
		{
			name: "single_number_record",
			file: &File{Records: []Record{
				{Type: "number", Key: "ANSWER", Value: "42"},
			}},
			want: "record number ANSWER = 42\n",
		},
		{
			name: "record_with_leading_comment",
			file: &File{Records: []Record{
				{LeadingComments: []string{"leading"}, Type: "string", Key: "G", Value: "hi"},
			}},
			want: "# leading\nrecord string G = \"hi\"\n",
		},
		{
			name: "record_with_two_leading_comments",
			file: &File{Records: []Record{
				{LeadingComments: []string{"first", "second"}, Type: "string", Key: "G", Value: "hi"},
			}},
			want: "# first\n# second\nrecord string G = \"hi\"\n",
		},
		{
			name: "string_with_escapes",
			file: &File{Records: []Record{
				{Type: "string", Key: "K", Value: "a\"b\n\t\\"},
			}},
			want: "record string K = \"a\\\"b\\n\\t\\\\\"\n",
		},
		{
			name: "block_with_one_record",
			file: &File{Blocks: []Block{
				{Name: "COLORS", Records: []Record{{Type: "string", Key: "RED", Value: "ff"}}},
			}},
			want: "block COLORS {\nrecord string RED = \"ff\";\n}\n",
		},
		{
			name: "block_with_leading_comment",
			file: &File{Blocks: []Block{
				{LeadingComments: []string{"a block"}, Name: "B", Records: []Record{
					{Type: "string", Key: "X", Value: "x"},
				}},
			}},
			want: "# a block\nblock B {\nrecord string X = \"x\";\n}\n",
		},
		{
			name: "block_with_inner_comment_on_record",
			file: &File{Blocks: []Block{
				{Name: "COLORS", Records: []Record{
					{LeadingComments: []string{"primary red"}, Type: "string", Key: "RED", Value: "ff"},
					{Type: "string", Key: "BLUE", Value: "00"},
				}},
			}},
			want: "block COLORS {\n# primary red\nrecord string RED = \"ff\";\nrecord string BLUE = \"00\";\n}\n",
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
			name:   "two_records",
			source: "record string A = \"a\"\nrecord number B = 1",
		},
		{
			name:   "comment_above_record",
			source: "# leading comment\nrecord string GREETING = \"hi\"",
		},
		{
			name:   "blank_line_then_comment_then_record",
			source: "\n# leading comment\nrecord string GREETING = \"hi\"",
		},
		{
			name:   "two_comments_above_same_record",
			source: "# first\n# second\nrecord string GREETING = \"hi\"",
		},
		{
			name:   "comment_above_block",
			source: "# block comment\nblock COLORS { record string RED = \"ff\"; }",
		},
		{
			name:   "block_with_inner_comment",
			source: "block COLORS {\n# primary red\nrecord string RED = \"ff\";\nrecord string BLUE = \"00\";\n}",
		},
		{
			name:   "record_string_with_escapes",
			source: `record string K = "a\"b\n\t\\"`,
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
