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
			name:  "comment_above_record",
			input: "# greeting\nrecord string GREETING = \"hello\"\n",
			want: &File{
				Statements: []Statement{
					Record{
						LeadingComments: []string{"greeting"},
						Type:            RecordTypeString,
						Key:             "GREETING",
						Value:           "hello",
					},
				},
			},
		},
		{
			name:  "blank_line_then_comment_then_record",
			input: "\n# greeting\nrecord string GREETING = \"hello\"\n",
			want: &File{
				Statements: []Statement{
					Record{
						LeadingComments: []string{"greeting"},
						Type:            RecordTypeString,
						Key:             "GREETING",
						Value:           "hello",
					},
				},
			},
		},
		{
			name:  "two_comments_above_same_record",
			input: "# first comment\n# second comment\nrecord string GREETING = \"hello\"\n",
			want: &File{
				Statements: []Statement{
					Record{
						LeadingComments: []string{"first comment", "second comment"},
						Type:            RecordTypeString,
						Key:             "GREETING",
						Value:           "hello",
					},
				},
			},
		},
		{
			name:  "comment_above_block",
			input: "# colors block\nblock COLORS {\n    record string RED = \"ff0000\";\n}\n",
			want: &File{
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
