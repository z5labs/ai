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
			name:  "empty_block",
			input: "block X { }",
			want: &File{
				Blocks: []Block{
					{Name: "X"},
				},
			},
		},
		{
			name:  "block_with_one_record",
			input: `block X { record string A = "1"; }`,
			want: &File{
				Blocks: []Block{
					{
						Name: "X",
						Records: []Record{
							{Type: RecordTypeString, Key: "A", Value: "1"},
						},
					},
				},
			},
		},
		{
			name:  "block_with_two_records",
			input: `block X { record string A = "1"; record string B = "2"; }`,
			want: &File{
				Blocks: []Block{
					{
						Name: "X",
						Records: []Record{
							{Type: RecordTypeString, Key: "A", Value: "1"},
							{Type: RecordTypeString, Key: "B", Value: "2"},
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
