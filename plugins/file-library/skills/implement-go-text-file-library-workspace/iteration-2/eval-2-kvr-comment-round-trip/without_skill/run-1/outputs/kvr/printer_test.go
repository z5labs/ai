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
			name: "comment_above_record",
			file: &File{
				Statements: []Statement{
					Record{
						LeadingComments: []string{"greeting"},
						Type:            RecordTypeString,
						Key:             "GREETING",
						Value:           "hello",
					},
				},
			},
			want: "# greeting\nrecord string GREETING = \"hello\"\n",
		},
		{
			name: "two_comments_above_same_record",
			file: &File{
				Statements: []Statement{
					Record{
						LeadingComments: []string{"first comment", "second comment"},
						Type:            RecordTypeString,
						Key:             "GREETING",
						Value:           "hello",
					},
				},
			},
			want: "# first comment\n# second comment\nrecord string GREETING = \"hello\"\n",
		},
		{
			name: "comment_above_block",
			file: &File{
				Statements: []Statement{
					Block{
						LeadingComments: []string{"colors block"},
						Name:            "COLORS",
						Records: []Record{
							{
								Type:  RecordTypeString,
								Key:   "RED",
								Value: "ff0000",
							},
						},
					},
				},
			},
			want: "# colors block\nblock COLORS {\n    record string RED = \"ff0000\";\n}\n",
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
			name:   "comment_above_record_round_trips",
			source: "# greeting\nrecord string GREETING = \"hello\"\n",
		},
		{
			name:   "blank_line_then_comment_then_record_round_trips",
			source: "\n# greeting\nrecord string GREETING = \"hello\"\n",
		},
		{
			name:   "two_comments_above_same_record_round_trip",
			source: "# first comment\n# second comment\nrecord string GREETING = \"hello\"\n",
		},
		{
			name:   "comment_above_block_round_trips",
			source: "# colors block\nblock COLORS {\n    record string RED = \"ff0000\";\n}\n",
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
