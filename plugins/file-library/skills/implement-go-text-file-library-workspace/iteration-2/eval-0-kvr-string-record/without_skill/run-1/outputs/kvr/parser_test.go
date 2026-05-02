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
					{Type: RecordTypeString, Key: "GREETING", Value: "hello"},
				},
			},
		},
		{
			name:  "two_string_records_separated_by_newline",
			input: "record string A = \"1\"\nrecord string B = \"2\"",
			want: &File{
				Records: []Record{
					{Type: RecordTypeString, Key: "A", Value: "1"},
					{Type: RecordTypeString, Key: "B", Value: "2"},
				},
			},
		},
		{
			name:  "string_value_decodes_escapes",
			input: `record string K = "a\"b\n"`,
			want: &File{
				Records: []Record{
					{Type: RecordTypeString, Key: "K", Value: "a\"b\n"},
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
