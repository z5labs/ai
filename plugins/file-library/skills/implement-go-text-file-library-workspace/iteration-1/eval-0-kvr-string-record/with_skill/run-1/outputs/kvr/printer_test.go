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
			file: &File{Records: []Record{{Type: "string", Key: "GREETING", Value: "hello"}}},
			want: `record string GREETING = "hello"` + "\n",
		},
		{
			name: "two_string_records",
			file: &File{Records: []Record{
				{Type: "string", Key: "A", Value: "1"},
				{Type: "string", Key: "B", Value: "2"},
			}},
			want: "record string A = \"1\"\nrecord string B = \"2\"\n",
		},
		{
			name: "string_value_with_quote_is_escaped",
			file: &File{Records: []Record{{Type: "string", Key: "K", Value: `a"b`}}},
			want: `record string K = "a\"b"` + "\n",
		},
		{
			name: "string_value_with_backslash_is_escaped",
			file: &File{Records: []Record{{Type: "string", Key: "K", Value: `a\b`}}},
			want: `record string K = "a\\b"` + "\n",
		},
		{
			name: "string_value_with_newline_is_escaped",
			file: &File{Records: []Record{{Type: "string", Key: "K", Value: "a\nb"}}},
			want: `record string K = "a\nb"` + "\n",
		},
		{
			name: "string_value_with_tab_is_escaped",
			file: &File{Records: []Record{{Type: "string", Key: "K", Value: "a\tb"}}},
			want: `record string K = "a\tb"` + "\n",
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
			name:   "two_string_records",
			source: "record string A = \"1\"\nrecord string B = \"2\"",
		},
		{
			name:   "value_with_escaped_quote",
			source: `record string K = "a\"b"`,
		},
		{
			name:   "value_with_escaped_backslash",
			source: `record string K = "a\\b"`,
		},
		{
			name:   "value_with_escaped_newline",
			source: `record string K = "line\n"`,
		},
		{
			name:   "value_with_escaped_tab",
			source: `record string K = "a\tb"`,
		},
		{
			name:   "empty_string_value",
			source: `record string K = ""`,
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
