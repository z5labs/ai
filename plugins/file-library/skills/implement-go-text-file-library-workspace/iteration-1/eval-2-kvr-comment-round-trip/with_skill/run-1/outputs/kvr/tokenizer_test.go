package kvr

import (
	"errors"
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// collect drains an iter.Seq2[Token, error] into a flat slice. It returns
// every successfully yielded token plus the first non-nil error.
func collect(seq iter.Seq2[Token, error]) ([]Token, error) {
	var tokens []Token
	var err error
	for tok, e := range seq {
		if e != nil {
			err = e
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens, err
}

func TestTokenizer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  []Token
	}{
		{
			name:  "empty_input_yields_no_tokens",
			input: "",
			want:  nil,
		},
		{
			name:  "single_identifier",
			input: "hello",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "hello"},
			},
		},
		{
			name:  "identifier_with_underscore_and_digits",
			input: "user_id_123",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "user_id_123"},
			},
		},
		{
			name:  "leading_underscore_identifier",
			input: "_temp",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "_temp"},
			},
		},
		{
			name:  "symbols",
			input: "={};",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 2}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 3}, Type: TokenSymbol, Value: "}"},
				{Pos: Pos{Line: 1, Column: 4}, Type: TokenSymbol, Value: ";"},
			},
		},
		{
			name:  "string_simple",
			input: `"hello"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "hello"},
			},
		},
		{
			name:  "string_with_escape_quote",
			input: `"a\"b"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: `a"b`},
			},
		},
		{
			name:  "string_with_escape_newline",
			input: `"line1\n"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "line1\n"},
			},
		},
		{
			name:  "string_with_escape_tab",
			input: `"a\tb"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "a\tb"},
			},
		},
		{
			name:  "string_with_escape_backslash",
			input: `"a\\b"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: `a\b`},
			},
		},
		{
			name:  "number",
			input: "42",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenNumber, Value: "42"},
			},
		},
		{
			name:  "comment_with_space",
			input: "# hello world",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "hello world"},
			},
		},
		{
			name:  "comment_tight",
			input: "#tight",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "tight"},
			},
		},
		{
			name:  "comment_then_newline_then_identifier",
			input: "# a comment\nfoo",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "a comment"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "foo"},
			},
		},
		{
			name:  "comment_at_end_of_file_no_trailing_newline",
			input: "foo # tail",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "foo"},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenComment, Value: "tail"},
			},
		},
		{
			name:  "record_line",
			input: `record string GREETING = "hello"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenIdentifier, Value: "string"},
				{Pos: Pos{Line: 1, Column: 15}, Type: TokenIdentifier, Value: "GREETING"},
				{Pos: Pos{Line: 1, Column: 24}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 26}, Type: TokenString, Value: "hello"},
			},
		},
		{
			name:  "comment_above_record",
			input: "# greet\nrecord string A = \"1\"",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "greet"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 2, Column: 8}, Type: TokenIdentifier, Value: "string"},
				{Pos: Pos{Line: 2, Column: 15}, Type: TokenIdentifier, Value: "A"},
				{Pos: Pos{Line: 2, Column: 17}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 2, Column: 19}, Type: TokenString, Value: "1"},
			},
		},
		{
			name:  "newline_then_position_resets",
			input: "a\nb",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "a"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "b"},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := collect(Tokenize(strings.NewReader(tc.input)))
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestTokenizerErrors(t *testing.T) {
	t.Parallel()

	t.Run("unterminated_string_at_eof", func(t *testing.T) {
		t.Parallel()

		_, err := collect(Tokenize(strings.NewReader(`"unterminated`)))
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
	})

	t.Run("unterminated_string_with_literal_newline", func(t *testing.T) {
		t.Parallel()

		_, err := collect(Tokenize(strings.NewReader("\"line1\nrest")))
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
	})

	t.Run("invalid_escape", func(t *testing.T) {
		t.Parallel()

		_, err := collect(Tokenize(strings.NewReader(`"a\zb"`)))
		var ie *InvalidEscapeError
		require.ErrorAs(t, err, &ie)
		require.Equal(t, 'z', ie.Char)
	})

	t.Run("unexpected_character", func(t *testing.T) {
		t.Parallel()

		_, err := collect(Tokenize(strings.NewReader("@")))
		var uc *UnexpectedCharacterError
		require.ErrorAs(t, err, &uc)
		require.Equal(t, '@', uc.Char)
		require.Equal(t, Pos{Line: 1, Column: 1}, uc.Pos)
	})

	t.Run("error_chains_via_errors_is_not_required", func(t *testing.T) {
		t.Parallel()

		// just confirm error is non-nil and not nil-typed
		_, err := collect(Tokenize(strings.NewReader("?")))
		require.Error(t, err)
		require.False(t, errors.Is(err, nil))
	})
}
