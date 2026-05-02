package kvrx

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

// TestPrinterDirectRecordBool pins the formatting choice for record-bool
// statements: `record bool KEY = true\n` (single space separation, lower-case
// `bool`, lower-case `true`/`false`, single trailing newline). The
// expectation source is built from a Parse() call so the test does not
// hand-construct AST shape — we only assert the printer's output bytes.
func TestPrinterDirectRecordBool(t *testing.T) {
	t.Parallel()

	t.Run("single_record_bool_true", func(t *testing.T) {
		t.Parallel()
		f, err := Parse(strings.NewReader(`record bool ENABLED = true`))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		require.Equal(t, "record bool ENABLED = true\n", buf.String())
	})

	t.Run("single_record_bool_false", func(t *testing.T) {
		t.Parallel()
		f, err := Parse(strings.NewReader(`record bool DARK = false`))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		require.Equal(t, "record bool DARK = false\n", buf.String())
	})

	t.Run("two_record_bools", func(t *testing.T) {
		t.Parallel()
		f, err := Parse(strings.NewReader("record bool A = true\nrecord bool B = false"))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		require.Equal(t, "record bool A = true\nrecord bool B = false\n", buf.String())
	})
}

// TestPrinterDirectConditional pins the formatting for conditional
// statements: keyword on its own line, body indented 2 spaces, statements
// terminated with `;`, closing brace on its own line, elif/else continue on
// the same line as the prior closing brace.
func TestPrinterDirectConditional(t *testing.T) {
	t.Parallel()

	t.Run("if_only", func(t *testing.T) {
		t.Parallel()
		f, err := Parse(strings.NewReader("if (true) {\n  record bool A = true;\n}"))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		want := "if (true) {\n  record bool A = true;\n}\n"
		require.Equal(t, want, buf.String())
	})

	t.Run("if_elif_else", func(t *testing.T) {
		t.Parallel()
		src := "record bool MODE = true\nif (&MODE) {\n  record bool R = true;\n} elif (false) {\n  record bool R = false;\n} else {\n  record bool R = false;\n}"
		f, err := Parse(strings.NewReader(src))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		want := "record bool MODE = true\nif (&MODE) {\n  record bool R = true;\n} elif (false) {\n  record bool R = false;\n} else {\n  record bool R = false;\n}\n"
		require.Equal(t, want, buf.String())
	})

	t.Run("if_with_eq_bool_literal", func(t *testing.T) {
		t.Parallel()
		src := "record bool MODE = true\nif (&MODE == true) {\n  record bool R = false;\n}"
		f, err := Parse(strings.NewReader(src))
		require.NoError(t, err)
		var buf bytes.Buffer
		require.NoError(t, Print(&buf, f))
		want := "record bool MODE = true\nif (&MODE == true) {\n  record bool R = false;\n}\n"
		require.Equal(t, want, buf.String())
	})
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
			name:   "single_record_bool_true",
			source: `record bool ENABLED = true`,
		},
		{
			name:   "single_record_bool_false",
			source: `record bool DARK = false`,
		},
		{
			name:   "two_record_bools",
			source: "record bool A = true\nrecord bool B = false",
		},
		{
			name:   "if_only",
			source: "if (true) {\n  record bool A = true;\n}",
		},
		{
			name:   "if_else",
			source: "if (false) {\n  record bool A = true;\n} else {\n  record bool B = false;\n}",
		},
		{
			name:   "if_elif_else",
			source: "record bool MODE = true\nif (&MODE) {\n  record bool R = true;\n} elif (false) {\n  record bool R = false;\n} else {\n  record bool R = false;\n}",
		},
		{
			name:   "if_with_eq_bool_literal",
			source: "record bool MODE = true\nif (&MODE == true) {\n  record bool R = false;\n}",
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

			second, err := Parse(strings.NewReader(buf.String()))
			require.NoError(t, err)
			require.Equal(t, first, second)
		})
	}
}
