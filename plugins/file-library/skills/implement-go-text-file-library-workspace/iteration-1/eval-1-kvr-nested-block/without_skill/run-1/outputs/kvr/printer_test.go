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
					{Type: "string", Key: "GREETING", Value: StringValue{V: "hello"}},
				},
			},
			want: "record string GREETING = \"hello\"\n",
		},
		{
			name: "single_number_record",
			file: &File{
				Records: []Record{
					{Type: "number", Key: "ANSWER", Value: NumberValue{V: "42"}},
				},
			},
			want: "record number ANSWER = 42\n",
		},
		{
			name: "record_with_leading_comment",
			file: &File{
				Records: []Record{
					{
						LeadingComments: []string{"greet"},
						Type:            "string",
						Key:             "A",
						Value:           StringValue{V: "1"},
					},
				},
			},
			want: "# greet\nrecord string A = \"1\"\n",
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
					{
						Name: "COLORS",
						Records: []Record{
							{Type: "string", Key: "RED", Value: StringValue{V: "ff0000"}},
						},
					},
				},
			},
			want: "block COLORS {\n\trecord string RED = \"ff0000\";\n}\n",
		},
		{
			name: "block_with_two_records",
			file: &File{
				Blocks: []Block{
					{
						Name: "COLORS",
						Records: []Record{
							{Type: "string", Key: "RED", Value: StringValue{V: "ff0000"}},
							{Type: "string", Key: "BLUE", Value: StringValue{V: "0000ff"}},
						},
					},
				},
			},
			want: "block COLORS {\n\trecord string RED = \"ff0000\";\n\trecord string BLUE = \"0000ff\";\n}\n",
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
			name: "two_records_with_leading_comment",
			source: "# the universal greeting\n" +
				"record string GREETING = \"hello\"\n" +
				"record number ANSWER = 42\n",
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
			name: "block_with_two_records",
			source: "block COLORS {\n" +
				"    record string RED  = \"ff0000\";\n" +
				"    record string BLUE = \"0000ff\";\n" +
				"}\n",
		},
		{
			name: "block_with_inner_leading_comment",
			source: "block COLORS {\n" +
				"    # primary red\n" +
				"    record string RED = \"ff0000\";\n" +
				"    record string BLUE = \"0000ff\";\n" +
				"}\n",
		},
		{
			name: "records_then_block",
			source: "record string A = \"1\"\n" +
				"block X { record string B = \"2\"; }\n",
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
