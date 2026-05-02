package kvrx

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

// TestParserRecordBool drives Parse() with real source strings for
// `record bool KEY = true|false`. Expectations are produced by Parse over a
// canonical source rather than hand-constructed, except for shape checks that
// inspect specific fields of the parsed AST.
func TestParserRecordBool(t *testing.T) {
	t.Parallel()

	t.Run("record_bool_true_shape", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader(`record bool ENABLED = true`))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		rec, ok := got.Statements[0].(*Record)
		require.True(t, ok, "first statement must be *Record")
		require.Equal(t, "bool", rec.Type)
		require.Equal(t, "ENABLED", rec.Key)
		bl, ok := rec.Value.(*BoolLiteral)
		require.True(t, ok, "value must be *BoolLiteral")
		require.Equal(t, true, bl.Value)
	})

	t.Run("record_bool_false_shape", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader(`record bool DARK = false`))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		rec, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		require.Equal(t, "bool", rec.Type)
		require.Equal(t, "DARK", rec.Key)
		bl, ok := rec.Value.(*BoolLiteral)
		require.True(t, ok)
		require.Equal(t, false, bl.Value)
	})

	t.Run("two_record_bool_statements", func(t *testing.T) {
		t.Parallel()
		want, err := Parse(strings.NewReader("record bool A = true\nrecord bool B = false"))
		require.NoError(t, err)
		got, err := Parse(strings.NewReader("record bool A = true\nrecord bool B = false"))
		require.NoError(t, err)
		require.Equal(t, want, got)
		require.Len(t, got.Statements, 2)
	})

	t.Run("missing_value_returns_unexpected_end_of_tokens", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`record bool ENABLED =`))
		var ueot *UnexpectedEndOfTokensError
		require.ErrorAs(t, err, &ueot)
	})

	t.Run("unexpected_token_at_statement_position", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader(`= bool ENABLED = true`))
		var ute *UnexpectedTokenError
		require.ErrorAs(t, err, &ute)
	})
}

// TestParserConditional drives Parse() with real source strings for
// `if (...) { ... } elif (...) { ... } else { ... }`. The parser evaluates
// the condition at parse time using previously-declared records (a tiny
// scope-walking helper); non-matching branches are preserved in the AST so
// they round-trip through the printer.
func TestParserConditional(t *testing.T) {
	t.Parallel()

	t.Run("if_only_true_constant_takes_branch", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("if (true) {\n  record bool A = true;\n}"))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		cond, ok := got.Statements[0].(*Conditional)
		require.True(t, ok, "first statement must be *Conditional")
		require.Len(t, cond.Branches, 1)
		require.Equal(t, "if", cond.Branches[0].Keyword)
		require.NotNil(t, cond.Branches[0].Condition)
		require.Len(t, cond.Branches[0].Body, 1)
		require.Equal(t, 0, cond.Active)
	})

	t.Run("if_only_false_constant_takes_no_branch", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("if (false) {\n  record bool A = true;\n}"))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		cond, ok := got.Statements[0].(*Conditional)
		require.True(t, ok)
		require.Equal(t, -1, cond.Active, "no branch should be active when condition is false")
	})

	t.Run("if_else_else_branch_takes_when_if_false", func(t *testing.T) {
		t.Parallel()
		got, err := Parse(strings.NewReader("if (false) {\n  record bool A = true;\n} else {\n  record bool B = false;\n}"))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		cond, ok := got.Statements[0].(*Conditional)
		require.True(t, ok)
		require.Len(t, cond.Branches, 2)
		require.Equal(t, "if", cond.Branches[0].Keyword)
		require.Equal(t, "else", cond.Branches[1].Keyword)
		require.Nil(t, cond.Branches[1].Condition, "else branch has no condition")
		require.Equal(t, 1, cond.Active)
	})

	t.Run("if_elif_else_elif_takes", func(t *testing.T) {
		t.Parallel()
		src := "if (false) {\n  record bool A = true;\n} elif (true) {\n  record bool B = true;\n} else {\n  record bool C = true;\n}"
		got, err := Parse(strings.NewReader(src))
		require.NoError(t, err)
		require.Len(t, got.Statements, 1)
		cond, ok := got.Statements[0].(*Conditional)
		require.True(t, ok)
		require.Len(t, cond.Branches, 3)
		require.Equal(t, "if", cond.Branches[0].Keyword)
		require.Equal(t, "elif", cond.Branches[1].Keyword)
		require.Equal(t, "else", cond.Branches[2].Keyword)
		require.Equal(t, 1, cond.Active)
	})

	t.Run("conditional_with_reference_lookup", func(t *testing.T) {
		t.Parallel()
		src := "record bool ENABLE_TLS = true\nif (&ENABLE_TLS) {\n  record bool A = true;\n}"
		got, err := Parse(strings.NewReader(src))
		require.NoError(t, err)
		require.Len(t, got.Statements, 2)
		_, ok := got.Statements[0].(*Record)
		require.True(t, ok)
		cond, ok := got.Statements[1].(*Conditional)
		require.True(t, ok)
		require.Equal(t, 0, cond.Active, "&ENABLE_TLS resolves to true so branch is taken")
	})

	t.Run("conditional_with_reference_eq_bool_literal", func(t *testing.T) {
		t.Parallel()
		src := "record bool MODE = true\nif (&MODE == true) {\n  record bool R = false;\n} else {\n  record bool R = true;\n}"
		got, err := Parse(strings.NewReader(src))
		require.NoError(t, err)
		require.Len(t, got.Statements, 2)
		cond, ok := got.Statements[1].(*Conditional)
		require.True(t, ok)
		require.Equal(t, 0, cond.Active)
	})

	t.Run("conditional_undeclared_reference_returns_error", func(t *testing.T) {
		t.Parallel()
		_, err := Parse(strings.NewReader("if (&MISSING) {\n  record bool A = true;\n}"))
		var ure *UndeclaredReferenceError
		require.ErrorAs(t, err, &ure)
		require.Equal(t, "MISSING", ure.Name)
	})
}
