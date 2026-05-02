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
			name: "single_string_record_prints_canonical_form",
			file: &File{
				Records: []Record{
					{Type: RecordTypeString, Key: "GREETING", Value: "hello"},
				},
			},
			want: "record string GREETING = \"hello\"\n",
		},
		{
			name: "two_string_records_each_on_own_line",
			file: &File{
				Records: []Record{
					{Type: RecordTypeString, Key: "A", Value: "1"},
					{Type: RecordTypeString, Key: "B", Value: "2"},
				},
			},
			want: "record string A = \"1\"\nrecord string B = \"2\"\n",
		},
		{
			name: "string_value_escapes_special_characters",
			file: &File{
				Records: []Record{
					{Type: RecordTypeString, Key: "K", Value: "a\"b\\c\nd\te"},
				},
			},
			want: `record string K = "a\"b\\c\nd\te"` + "\n",
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
			name:   "single_string_record_round_trips",
			source: `record string GREETING = "hello"`,
		},
		{
			name:   "two_string_records_round_trip",
			source: "record string A = \"1\"\nrecord string B = \"2\"",
		},
		{
			name:   "string_value_with_escapes_round_trips",
			source: `record string K = "a\"b\\c\n\t"`,
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
