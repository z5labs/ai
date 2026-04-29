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
			name: "single_number_record",
			file: &File{
				Records: []Record{
					{Type: "number", Key: "ANSWER", Value: "42"},
				},
			},
			want: "record number ANSWER = 42\n",
		},
		{
			name: "two_top_level_records",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "A", Value: "1"},
					{Type: "number", Key: "B", Value: "2"},
				},
			},
			want: "record string A = \"1\"\nrecord number B = 2\n",
		},
		{
			name: "empty_block",
			file: &File{
				Blocks: []Block{
					{Name: "COLORS"},
				},
			},
			want: "block COLORS {\n}\n",
		},
		{
			name: "block_with_one_record",
			file: &File{
				Blocks: []Block{
					{Name: "COLORS", Records: []Record{
						{Type: "string", Key: "RED", Value: "ff0000"},
					}},
				},
			},
			want: "block COLORS {\n    record string RED = \"ff0000\";\n}\n",
		},
		{
			name: "block_with_two_records",
			file: &File{
				Blocks: []Block{
					{Name: "COLORS", Records: []Record{
						{Type: "string", Key: "RED", Value: "ff0000"},
						{Type: "string", Key: "BLUE", Value: "0000ff"},
					}},
				},
			},
			want: "block COLORS {\n    record string RED = \"ff0000\";\n    record string BLUE = \"0000ff\";\n}\n",
		},
		{
			name: "string_value_with_quote_is_escaped",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "Q", Value: `a"b`},
				},
			},
			want: "record string Q = \"a\\\"b\"\n",
		},
		{
			name: "string_value_with_newline_is_escaped",
			file: &File{
				Records: []Record{
					{Type: "string", Key: "Q", Value: "x\ny"},
				},
			},
			want: "record string Q = \"x\\ny\"\n",
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
			name:   "two_top_level_records",
			source: "record string A = \"1\"\nrecord number B = 2",
		},
		{
			name:   "empty_block",
			source: `block COLORS { }`,
		},
		{
			name:   "block_with_one_record",
			source: `block COLORS { record string RED = "ff0000"; }`,
		},
		{
			name:   "block_with_two_records",
			source: `block COLORS { record string RED = "ff0000"; record string BLUE = "0000ff"; }`,
		},
		{
			name:   "top_level_record_then_block",
			source: "record string G = \"hi\"\nblock B { record number N = 1; }",
		},
		{
			name:   "string_with_escaped_quote",
			source: `record string Q = "a\"b"`,
		},
		{
			name:   "string_with_escaped_newline",
			source: `record string Q = "x\ny"`,
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
