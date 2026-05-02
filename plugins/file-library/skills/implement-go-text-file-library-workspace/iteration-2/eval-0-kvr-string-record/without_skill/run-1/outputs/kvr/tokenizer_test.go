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
			name:  "string_record_yields_idents_symbol_and_string",
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
			name:  "leading_whitespace_shifts_first_token_column",
			input: `   record`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 4}, Type: TokenIdentifier, Value: "record"},
			},
		},
		{
			name:  "string_with_escapes_decodes_value",
			input: `"a\"b\\c\n\t"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "a\"b\\c\n\t"},
			},
		},
		{
			name:  "newline_increments_line_and_resets_column",
			input: "record\nstring",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "string"},
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
