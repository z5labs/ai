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
			input: "foo",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "foo"},
			},
		},
		{
			name:  "single_number",
			input: "42",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenNumber, Value: "42"},
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
			name:  "string_with_escapes",
			input: `"a\"b\n\t\\"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "a\"b\n\t\\"},
			},
		},
		{
			name:  "symbols",
			input: "= { } ;",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 3}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenSymbol, Value: "}"},
				{Pos: Pos{Line: 1, Column: 7}, Type: TokenSymbol, Value: ";"},
			},
		},
		{
			name:  "comment_with_space_after_hash",
			input: "# hello world",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "hello world"},
			},
		},
		{
			name:  "comment_tight_after_hash",
			input: "#tight",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "tight"},
			},
		},
		{
			name:  "comment_then_record_on_next_line",
			input: "# leading\nrecord",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "leading"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "record"},
			},
		},
		{
			name:  "two_comments_in_a_row",
			input: "# one\n# two\n",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "one"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenComment, Value: "two"},
			},
		},
		{
			name:  "record_declaration",
			input: `record string A = "1"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenIdentifier, Value: "string"},
				{Pos: Pos{Line: 1, Column: 15}, Type: TokenIdentifier, Value: "A"},
				{Pos: Pos{Line: 1, Column: 17}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 19}, Type: TokenString, Value: "1"},
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

	testCases := []struct {
		name    string
		input   string
		assertE func(t *testing.T, err error)
	}{
		{
			name:  "unterminated_string_at_eof",
			input: `"oops`,
			assertE: func(t *testing.T, err error) {
				var ute *UnterminatedStringError
				require.ErrorAs(t, err, &ute)
			},
		},
		{
			name:  "unterminated_string_at_newline",
			input: "\"oops\nrecord",
			assertE: func(t *testing.T, err error) {
				var ute *UnterminatedStringError
				require.ErrorAs(t, err, &ute)
			},
		},
		{
			name:  "unexpected_character",
			input: "@",
			assertE: func(t *testing.T, err error) {
				var uc *UnexpectedCharacterError
				require.ErrorAs(t, err, &uc)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := collect(Tokenize(strings.NewReader(tc.input)))
			require.Error(t, err)
			tc.assertE(t, err)
		})
	}
}
