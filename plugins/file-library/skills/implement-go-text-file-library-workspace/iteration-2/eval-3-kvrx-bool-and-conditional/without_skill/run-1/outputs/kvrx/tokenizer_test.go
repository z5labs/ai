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
			name:  "record_bool_true",
			input: "record bool ENABLED = true",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenIdentifier, Value: "bool"},
				{Pos: Pos{Line: 1, Column: 13}, Type: TokenIdentifier, Value: "ENABLED"},
				{Pos: Pos{Line: 1, Column: 21}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 23}, Type: TokenIdentifier, Value: "true"},
			},
		},
		{
			name:  "record_bool_false",
			input: "record bool DARK = false",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenIdentifier, Value: "bool"},
				{Pos: Pos{Line: 1, Column: 13}, Type: TokenIdentifier, Value: "DARK"},
				{Pos: Pos{Line: 1, Column: 18}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 20}, Type: TokenIdentifier, Value: "false"},
			},
		},
		{
			name:  "if_elif_else_keywords",
			input: "if elif else",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "if"},
				{Pos: Pos{Line: 1, Column: 4}, Type: TokenIdentifier, Value: "elif"},
				{Pos: Pos{Line: 1, Column: 9}, Type: TokenIdentifier, Value: "else"},
			},
		},
		{
			name:  "double_equals_is_single_token",
			input: "&MODE == true",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "&"},
				{Pos: Pos{Line: 1, Column: 2}, Type: TokenIdentifier, Value: "MODE"},
				{Pos: Pos{Line: 1, Column: 7}, Type: TokenSymbol, Value: "=="},
				{Pos: Pos{Line: 1, Column: 10}, Type: TokenIdentifier, Value: "true"},
			},
		},
		{
			name:  "if_block_with_newlines_yields_newline_tokens",
			input: "if (true) {\nrecord bool A = true;\n}",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "if"},
				{Pos: Pos{Line: 1, Column: 4}, Type: TokenSymbol, Value: "("},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenIdentifier, Value: "true"},
				{Pos: Pos{Line: 1, Column: 9}, Type: TokenSymbol, Value: ")"},
				{Pos: Pos{Line: 1, Column: 11}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 12}, Type: TokenNewline, Value: "\n"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 2, Column: 8}, Type: TokenIdentifier, Value: "bool"},
				{Pos: Pos{Line: 2, Column: 13}, Type: TokenIdentifier, Value: "A"},
				{Pos: Pos{Line: 2, Column: 15}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 2, Column: 17}, Type: TokenIdentifier, Value: "true"},
				{Pos: Pos{Line: 2, Column: 21}, Type: TokenSymbol, Value: ";"},
				{Pos: Pos{Line: 2, Column: 22}, Type: TokenNewline, Value: "\n"},
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
