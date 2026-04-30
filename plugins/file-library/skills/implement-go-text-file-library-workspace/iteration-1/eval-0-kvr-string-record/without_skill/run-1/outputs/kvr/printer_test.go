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
				Records: []Record{
					{Type: "string", Key: "GREETING", Value: "hello"},
				},
			},
			want: "record string GREETING = \"hello\"\n",
		},
		{
			name: "two_string_records",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "A", Value: "1"},
					{Type: "string", Key: "B", Value: "2"},
				},
			},
			want: "record string A = \"1\"\nrecord string B = \"2\"\n",
		},
		{
			name: "value_with_quote_is_escaped",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "Q", Value: `a"b`},
				},
			},
			want: "record string Q = \"a\\\"b\"\n",
		},
		{
			name: "value_with_newline_is_escaped",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "M", Value: "a\nb"},
				},
			},
			want: "record string M = \"a\\nb\"\n",
		},
		{
			name: "value_with_backslash_is_escaped",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "B", Value: `a\b`},
				},
			},
			want: "record string B = \"a\\\\b\"\n",
		},
		{
			name: "value_with_tab_is_escaped",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "T", Value: "a\tb"},
				},
			},
			want: "record string T = \"a\\tb\"\n",
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
			name:   "value_with_quote_round_trips",
			source: `record string Q = "a\"b"`,
		},
		{
			name:   "value_with_newline_escape_round_trips",
			source: `record string M = "a\nb"`,
		},
		{
			name:   "value_with_backslash_round_trips",
			source: `record string B = "a\\b"`,
		},
		{
			name:   "value_with_tab_round_trips",
			source: `record string T = "a\tb"`,
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
