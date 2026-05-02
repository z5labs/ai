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
			name:  "identifier_with_underscore_and_digits",
			input: "_temp123",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "_temp123"},
			},
		},
		{
			name:  "two_identifiers_separated_by_space",
			input: "foo bar",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "foo"},
				{Pos: Pos{Line: 1, Column: 5}, Type: TokenIdentifier, Value: "bar"},
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
			name:  "string_simple",
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
			name:  "number",
			input: "42",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenNumber, Value: "42"},
			},
		},
		{
			name:  "comment_at_start_of_line",
			input: "# hello world",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "hello world"},
			},
		},
		{
			name:  "comment_tight_no_space_after_hash",
			input: "#tight",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "tight"},
			},
		},
		{
			name:  "comment_then_newline_then_identifier",
			input: "# leading\nrecord",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "leading"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "record"},
			},
		},
		{
			name:  "comment_terminated_by_eof",
			input: "# trailing",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "trailing"},
			},
		},
		{
			name:  "two_comments_consecutive",
			input: "# one\n# two",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenComment, Value: "one"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenComment, Value: "two"},
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
		_, err := collect(Tokenize(strings.NewReader("\"a\nb\"")))
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
	})

	t.Run("invalid_escape", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader(`"a\xb"`)))
		var ie *InvalidEscapeError
		require.ErrorAs(t, err, &ie)
	})

	t.Run("unexpected_character", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader("?")))
		var uce *UnexpectedCharacterError
		require.ErrorAs(t, err, &uce)
		require.Equal(t, Pos{Line: 1, Column: 1}, uce.Pos)
		require.Equal(t, '?', uce.Char)
	})
}
