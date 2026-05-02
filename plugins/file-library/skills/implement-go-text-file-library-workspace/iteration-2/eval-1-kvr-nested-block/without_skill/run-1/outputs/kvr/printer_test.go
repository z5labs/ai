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
			name: "empty_block_prints_open_close_braces",
			file: &File{
				Blocks: []Block{{Name: "X"}},
			},
			want: "block X {\n}\n",
		},
		{
			name: "block_with_one_record",
			file: &File{
				Blocks: []Block{{
					Name: "X",
					Records: []Record{
						{Type: RecordTypeString, Key: "A", Value: "1"},
					},
				}},
			},
			want: "block X {\n    record string A = \"1\";\n}\n",
		},
		{
			name: "block_with_two_records",
			file: &File{
				Blocks: []Block{{
					Name: "X",
					Records: []Record{
						{Type: RecordTypeString, Key: "A", Value: "1"},
						{Type: RecordTypeString, Key: "B", Value: "2"},
					},
				}},
			},
			want: "block X {\n    record string A = \"1\";\n    record string B = \"2\";\n}\n",
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
			name:   "empty_block_round_trips",
			source: "block X { }",
		},
		{
			name:   "block_with_one_record_round_trips",
			source: `block X { record string A = "1"; }`,
		},
		{
			name:   "block_with_two_records_round_trips",
			source: `block X { record string A = "1"; record string B = "2"; }`,
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
