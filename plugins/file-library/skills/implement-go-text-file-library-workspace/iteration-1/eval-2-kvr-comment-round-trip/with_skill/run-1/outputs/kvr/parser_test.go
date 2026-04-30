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

func TestParserShape(t *testing.T) {
	t.Parallel()

	t.Run("single_string_record", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader(`record string GREETING = "hello"`))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		rec, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		require.Equal(t, "string", rec.Type)
		require.Equal(t, "GREETING", rec.Key)
		require.Equal(t, "hello", rec.Value)
		require.Equal(t, TokenString, rec.ValueKind)
		require.Empty(t, rec.LeadingComments)
	})

	t.Run("single_number_record", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader(`record number ANSWER = 42`))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		rec, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		require.Equal(t, "number", rec.Type)
		require.Equal(t, "ANSWER", rec.Key)
		require.Equal(t, "42", rec.Value)
		require.Equal(t, TokenNumber, rec.ValueKind)
	})

	t.Run("comment_above_record_attaches_as_leading", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("# greet\nrecord string A = \"1\""))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		rec, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		require.Equal(t, []string{"greet"}, rec.LeadingComments)
		require.Equal(t, "A", rec.Key)
	})

	t.Run("blank_line_then_comment_then_record", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("\n\n# greet\nrecord string A = \"1\""))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		rec, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		require.Equal(t, []string{"greet"}, rec.LeadingComments)
	})

	t.Run("two_comments_above_record_attach_in_order", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("# first\n# second\nrecord string A = \"1\""))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		rec, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		require.Equal(t, []string{"first", "second"}, rec.LeadingComments)
	})

	t.Run("two_records_separated_by_newline", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("record string A = \"1\"\nrecord number B = 2"))
		require.NoError(t, err)
		require.Len(t, got.Statements, 2)
		_, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		_, ok = got.Statements[1].(*Record)
		require.True(t, ok)
	})

	t.Run("comments_attach_only_to_following_record", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("# greet\nrecord string A = \"1\"\nrecord number B = 2"))
		require.NoError(t, err)
		require.Len(t, got.Statements, 2)
		rec1 := got.Statements[0].(*Record)
		rec2 := got.Statements[1].(*Record)
		require.Equal(t, []string{"greet"}, rec1.LeadingComments)
		require.Empty(t, rec2.LeadingComments)
	})

	t.Run("block_with_one_record", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader(`block COLORS { record string RED = "ff0000"; }`))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		blk, ok := got.Statements[0].(*Block)
		require.True(t, ok)
		require.Equal(t, "COLORS", blk.Name)
		require.Len(t, blk.Records, 1)
		require.Equal(t, "RED", blk.Records[0].Key)
		require.Equal(t, "ff0000", blk.Records[0].Value)
	})

	t.Run("comment_above_block_attaches_as_leading", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("# colors\nblock C { record string A = \"1\"; }"))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		blk, ok := got.Statements[0].(*Block)
		require.True(t, ok)
		require.Equal(t, []string{"colors"}, blk.LeadingComments)
	})

	t.Run("comment_inside_block_attaches_to_inner_record", func(t *testing.T) {
		t.Parallel()
		src := "block C {\n# primary\nrecord string A = \"1\";\nrecord string B = \"2\";\n}"
		got, err := Parse(strings.NewReader(src))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		blk := got.Statements[0].(*Block)
		require.Len(t, blk.Records, 2)
		require.Equal(t, []string{"primary"}, blk.Records[0].LeadingComments)
		require.Empty(t, blk.Records[1].LeadingComments)
	})
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	t.Run("unexpected_token_at_top_level", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("="))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
	})

	t.Run("record_missing_equals", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record string A "1"`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
	})

	t.Run("type_mismatch_string_with_number", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record string A = 42`))
		var tme *TypeMismatchError
		require.ErrorAs(t, err, &tme)
	})

	t.Run("type_mismatch_number_with_string", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record number A = "x"`))
		var tme *TypeMismatchError
		require.ErrorAs(t, err, &tme)
	})

	t.Run("unknown_type_name", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record bool A = "x"`))
		require.Error(t, err)
	})

	t.Run("block_missing_trailing_semicolon", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`block X { record string A = "1" }`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
	})
}
