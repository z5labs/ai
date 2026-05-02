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
			name:  "string_literal",
			input: `"hello"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "hello"},
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
			name:  "braces_and_semicolon_together",
			input: "{ ; }",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 3}, Type: TokenSymbol, Value: ";"},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenSymbol, Value: "}"},
			},
		},
		{
			name:  "block_record_full",
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
