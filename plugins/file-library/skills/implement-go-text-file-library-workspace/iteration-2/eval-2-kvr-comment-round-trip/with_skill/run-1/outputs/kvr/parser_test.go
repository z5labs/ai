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
		check func(t *testing.T, f *File)
	}{
		{
			name:  "empty_input_yields_zero_file",
			input: "",
			check: func(t *testing.T, f *File) {
				require.Equal(t, &File{}, f)
			},
		},
		{
			name:  "single_string_record",
			input: `record string GREETING = "hello"`,
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Records, 1)
				require.Len(t, f.Blocks, 0)
				rec := f.Records[0]
				require.Equal(t, "string", rec.Type)
				require.Equal(t, "GREETING", rec.Key)
				require.Equal(t, "hello", rec.Value)
				require.Empty(t, rec.LeadingComments)
			},
		},
		{
			name:  "single_number_record",
			input: `record number ANSWER = 42`,
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Records, 1)
				require.Equal(t, "number", f.Records[0].Type)
				require.Equal(t, "ANSWER", f.Records[0].Key)
				require.Equal(t, "42", f.Records[0].Value)
			},
		},
		{
			name:  "two_records",
			input: "record string A = \"a\"\nrecord number B = 1",
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Records, 2)
			},
		},
		{
			name:  "comment_above_record",
			input: "# leading comment\nrecord string GREETING = \"hi\"",
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Records, 1)
				require.Len(t, f.Blocks, 0)
				require.Equal(t, []string{"leading comment"}, f.Records[0].LeadingComments)
				require.Equal(t, "GREETING", f.Records[0].Key)
			},
		},
		{
			name:  "blank_line_then_comment_then_record",
			input: "\n# leading comment\nrecord string GREETING = \"hi\"",
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Records, 1)
				require.Equal(t, []string{"leading comment"}, f.Records[0].LeadingComments)
			},
		},
		{
			name:  "two_comments_above_same_record",
			input: "# first\n# second\nrecord string GREETING = \"hi\"",
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Records, 1)
				require.Equal(t, []string{"first", "second"}, f.Records[0].LeadingComments)
			},
		},
		{
			name:  "comment_above_block",
			input: "# block comment\nblock COLORS { record string RED = \"ff\"; }",
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Blocks, 1)
				require.Equal(t, []string{"block comment"}, f.Blocks[0].LeadingComments)
				require.Equal(t, "COLORS", f.Blocks[0].Name)
				require.Len(t, f.Blocks[0].Records, 1)
				require.Equal(t, "RED", f.Blocks[0].Records[0].Key)
			},
		},
		{
			name:  "two_comments_above_block",
			input: "# c1\n# c2\nblock B { record string X = \"x\"; }",
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Blocks, 1)
				require.Equal(t, []string{"c1", "c2"}, f.Blocks[0].LeadingComments)
			},
		},
		{
			name:  "block_with_inner_comment",
			input: "block COLORS {\n# primary red\nrecord string RED = \"ff\";\nrecord string BLUE = \"00\";\n}",
			check: func(t *testing.T, f *File) {
				require.Len(t, f.Blocks, 1)
				blk := f.Blocks[0]
				require.Len(t, blk.Records, 2)
				require.Equal(t, []string{"primary red"}, blk.Records[0].LeadingComments)
				require.Empty(t, blk.Records[1].LeadingComments)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(strings.NewReader(tc.input))
			require.NoError(t, err)
			tc.check(t, got)
		})
	}
}

func TestParserEquivalence(t *testing.T) {
	t.Parallel()

	// Per testing.md, parser tests should drive Parse() with real source
	// strings; the expected value is also produced by Parse() over a
	// canonical source.
	canonical := `record string GREETING = "hello"`
	want, err := Parse(strings.NewReader(canonical))
	require.NoError(t, err)

	got, err := Parse(strings.NewReader("record string GREETING = \"hello\"\n"))
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	t.Run("unexpected_token_in_record_value", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record string GREETING = }`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, TokenSymbol, ute.Got.Type)
	})

	t.Run("type_mismatch", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record string K = 42`))
		var tme *TypeMismatchError
		require.ErrorAs(t, err, &tme)
		require.Equal(t, "string", tme.Type)
	})

	t.Run("missing_trailing_semicolon_in_block", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`block B { record string A = "1" }`))
		require.Error(t, err)
	})
}
