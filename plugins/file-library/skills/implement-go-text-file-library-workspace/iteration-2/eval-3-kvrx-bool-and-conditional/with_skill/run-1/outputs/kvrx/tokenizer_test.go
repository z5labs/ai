package kvrx

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
			name:  "identifier_with_digits_and_underscore",
			input: "_temp123",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "_temp123"},
			},
		},
		{
			name:  "true_keyword_is_identifier",
			input: "true",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "true"},
			},
		},
		{
			name:  "false_keyword_is_identifier",
			input: "false",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "false"},
			},
		},
		{
			name:  "if_keyword_is_identifier",
			input: "if",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "if"},
			},
		},
		{
			name:  "elif_keyword_is_identifier",
			input: "elif",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "elif"},
			},
		},
		{
			name:  "else_keyword_is_identifier",
			input: "else",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "else"},
			},
		},
		{
			name:  "single_char_symbols",
			input: "={}();&",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 2}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 3}, Type: TokenSymbol, Value: "}"},
				{Pos: Pos{Line: 1, Column: 4}, Type: TokenSymbol, Value: "("},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenSymbol, Value: ")"},
				{Pos: Pos{Line: 1, Column: 6}, Type: TokenSymbol, Value: ";"},
				{Pos: Pos{Line: 1, Column: 7}, Type: TokenSymbol, Value: "&"},
			},
		},
		{
			name:  "two_char_eq_symbol",
			input: "==",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "=="},
			},
		},
		{
			name:  "single_string",
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
			name:  "newline_token",
			input: "\n",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenNewline, Value: "\n"},
			},
		},
		{
			name:  "record_bool_true_full_statement",
			input: `record bool ENABLED = true`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenIdentifier, Value: "bool"},
				{Pos: Pos{Line: 1, Column: 13}, Type: TokenIdentifier, Value: "ENABLED"},
				{Pos: Pos{Line: 1, Column: 21}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 23}, Type: TokenIdentifier, Value: "true"},
			},
		},
		{
			name:  "if_block_pos_across_newlines",
			input: "if (X) {\n  record bool A = false;\n}",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "if"},
				{Pos: Pos{Line: 1, Column: 4}, Type: TokenSymbol, Value: "("},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenIdentifier, Value: "X"},
				{Pos: Pos{Line: 1, Column: 6}, Type: TokenSymbol, Value: ")"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 9}, Type: TokenNewline, Value: "\n"},
				{Pos: Pos{Line: 2, Column: 3}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 2, Column: 10}, Type: TokenIdentifier, Value: "bool"},
				{Pos: Pos{Line: 2, Column: 15}, Type: TokenIdentifier, Value: "A"},
				{Pos: Pos{Line: 2, Column: 17}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 2, Column: 19}, Type: TokenIdentifier, Value: "false"},
				{Pos: Pos{Line: 2, Column: 24}, Type: TokenSymbol, Value: ";"},
				{Pos: Pos{Line: 2, Column: 25}, Type: TokenNewline, Value: "\n"},
				{Pos: Pos{Line: 3, Column: 1}, Type: TokenSymbol, Value: "}"},
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

	t.Run("unterminated_string_returns_typed_error", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader(`"unterminated`)))
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
	})

	t.Run("unexpected_character_returns_typed_error", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader("@")))
		var uce *UnexpectedCharacterError
		require.ErrorAs(t, err, &uce)
		require.Equal(t, Pos{Line: 1, Column: 1}, uce.Pos)
		require.Equal(t, '@', uce.Char)
	})
}
