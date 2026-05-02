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
			name: "two_string_records",
			file: &File{Records: []Record{
				{Type: "string", Key: "A", Value: "x"},
				{Type: "string", Key: "B", Value: "y"},
			}},
			want: "record string A = \"x\"\nrecord string B = \"y\"\n",
		},
		{
			name: "string_record_with_quote_and_backslash_escapes",
			file: &File{Records: []Record{
				{Type: "string", Key: "MSG", Value: `a"b\c`},
			}},
			want: "record string MSG = \"a\\\"b\\\\c\"\n",
		},
		{
			name: "string_record_with_newline_and_tab_escapes",
			file: &File{Records: []Record{
				{Type: "string", Key: "MSG", Value: "line1\n\tindented"},
			}},
			want: "record string MSG = \"line1\\n\\tindented\"\n",
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
			source: "record string A = \"x\"\nrecord string B = \"y\"",
		},
		{
			name:   "string_record_with_escaped_quote",
			source: `record string MSG = "a\"b"`,
		},
		{
			name:   "string_record_with_escaped_backslash",
			source: `record string MSG = "a\\b"`,
		},
		{
			name:   "string_record_with_escaped_newline_and_tab",
			source: `record string MSG = "line1\n\tindented"`,
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
