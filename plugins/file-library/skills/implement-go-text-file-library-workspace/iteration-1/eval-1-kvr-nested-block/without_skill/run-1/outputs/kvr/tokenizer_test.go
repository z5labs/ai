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
			name:  "identifier_with_digits_and_underscore",
			input: "_temp123",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "_temp123"},
			},
		},
		{
			name:  "number_token",
			input: "42",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenNumber, Value: "42"},
			},
		},
		{
			name:  "string_token_simple",
			input: `"hello"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "hello"},
			},
		},
		{
			name:  "string_token_with_escapes",
			input: `"a\"b\nc\\d\t"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "a\"b\nc\\d\t"},
			},
		},
		{
			name:  "comment_strips_leading_whitespace",
			input: "#   hello world\n",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "hello world"},
			},
		},
		{
			name:  "tight_comment",
			input: "#tight",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "tight"},
			},
		},
		{
			name:  "equal_symbol",
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
			name:  "all_block_symbols_in_sequence",
			input: "{ ; }",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenSymbol, Value: "{"},
				{Pos: Pos{Line: 1, Column: 3}, Type: TokenSymbol, Value: ";"},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenSymbol, Value: "}"},
			},
		},
		{
			name:  "record_statement",
			input: `record string GREETING = "hello"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "record"},
				{Pos: Pos{Line: 1, Column: 8}, Type: TokenIdentifier, Value: "string"},
				{Pos: Pos{Line: 1, Column: 15}, Type: TokenIdentifier, Value: "GREETING"},
				{Pos: Pos{Line: 1, Column: 24}, Type: TokenSymbol, Value: "="},
				{Pos: Pos{Line: 1, Column: 26}, Type: TokenString, Value: "hello"},
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
		_, err := collect(Tokenize(strings.NewReader(`"abc`)))
		require.Error(t, err)
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
	})

	t.Run("unterminated_string_at_newline", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader("\"abc\nrest")))
		require.Error(t, err)
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
	})

	t.Run("invalid_escape", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader(`"\q"`)))
		require.Error(t, err)
		var iee *InvalidEscapeError
		require.ErrorAs(t, err, &iee)
	})
}
