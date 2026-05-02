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
			input: "user_id_42",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "user_id_42"},
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
			name:  "simple_string_literal",
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
			name:  "string_with_escaped_backslash",
			input: `"a\\b"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: `a\b`},
			},
		},
		{
			name:  "string_with_escaped_newline_and_tab",
			input: `"a\nb\tc"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "a\nb\tc"},
			},
		},
		{
			name:  "record_string_declaration",
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
			name:  "two_records_separated_by_newline",
			input: "record string A = \"x\"\nrecord string B = \"y\"",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenIdentifier, Value: "string"},
				{Pos: Pos{Line: 1, Column: 15}, Type: TokenIdentifier, Value: "A"},
				{Pos: Pos{Line: 1, Column: 17}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 19}, Type: TokenString, Value: "x"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 2, Column: 8}, Type: TokenIdentifier, Value: "string"},
				{Pos: Pos{Line: 2, Column: 15}, Type: TokenIdentifier, Value: "B"},
				{Pos: Pos{Line: 2, Column: 17}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 2, Column: 19}, Type: TokenString, Value: "y"},
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

func TestTokenizerUnterminatedString(t *testing.T) {
	t.Parallel()

	_, err := collect(Tokenize(strings.NewReader(`"unterminated`)))
	var ute *UnterminatedStringError
	require.ErrorAs(t, err, &ute)
	require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
}

func TestTokenizerUnexpectedCharacter(t *testing.T) {
	t.Parallel()

	_, err := collect(Tokenize(strings.NewReader("@")))
	var uce *UnexpectedCharacterError
	require.True(t, errors.As(err, &uce))
	require.Equal(t, Pos{Line: 1, Column: 1}, uce.Pos)
	require.Equal(t, '@', uce.Char)
}
