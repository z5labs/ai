package kvr

import (
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
			name:  "identifier_with_underscores_and_digits",
			input: "_temp123",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "_temp123"},
			},
		},
		{
			name:  "equals_symbol",
			input: "=",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "="},
			},
		},
		{
			name:  "open_brace_symbol",
			input: "{",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "{"},
			},
		},
		{
			name:  "close_brace_symbol",
			input: "}",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "}"},
			},
		},
		{
			name:  "semicolon_symbol",
			input: ";",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: ";"},
			},
		},
		{
			name:  "all_brace_and_semicolon_symbols_in_a_row",
			input: "{};",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 2}, Type: TokenSymbol, Value: "}"},
				{Pos: Pos{Line: 1, Column: 3}, Type: TokenSymbol, Value: ";"},
			},
		},
		{
			name:  "string_literal",
			input: `"hello"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "hello"},
			},
		},
		{
			name:  "string_with_escaped_quote",
			input: `"a\"b"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: `a"b`},
			},
		},
		{
			name:  "string_with_escape_sequences",
			input: `"line1\n\t\\"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "line1\n\t\\"},
			},
		},
		{
			name:  "number_literal",
			input: "42",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenNumber, Value: "42"},
			},
		},
		{
			name:  "comment_to_end_of_line",
			input: "# hello world",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "hello world"},
			},
		},
		{
			name:  "tight_comment_no_leading_space",
			input: "#tight",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "tight"},
			},
		},
		{
			name:  "record_keyword_then_type_then_key_then_equals_then_string",
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
			name:  "block_with_one_record",
			input: `block X { record string A = "1"; }`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "block"},
				{Pos: Pos{Line: 1, Column: 7}, Type: TokenIdentifier, Value: "X"},
				{Pos: Pos{Line: 1, Column: 9}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 11}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 18}, Type: TokenIdentifier, Value: "string"},
				{Pos: Pos{Line: 1, Column: 25}, Type: TokenIdentifier, Value: "A"},
				{Pos: Pos{Line: 1, Column: 27}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 29}, Type: TokenString, Value: "1"},
				{Pos: Pos{Line: 1, Column: 32}, Type: TokenSymbol, Value: ";"},
				{Pos: Pos{Line: 1, Column: 34}, Type: TokenSymbol, Value: "}"},
			},
		},
		{
			name:  "newline_advances_line_and_resets_column",
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

	t.Run("unterminated_string_at_newline", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader("\"abc\ndef\"")))
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
	})

	t.Run("invalid_escape_sequence", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader(`"bad\x"`)))
		var ie *InvalidEscapeError
		require.ErrorAs(t, err, &ie)
	})

	t.Run("unexpected_character", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader("@")))
		var uc *UnexpectedCharacterError
		require.ErrorAs(t, err, &uc)
		require.Equal(t, '@', uc.Char)
	})
}
