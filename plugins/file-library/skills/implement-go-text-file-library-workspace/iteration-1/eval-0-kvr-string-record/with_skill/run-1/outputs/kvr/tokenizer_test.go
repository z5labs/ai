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
			input: "user_id123",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "user_id123"},
			},
		},
		{
			name:  "identifier_starting_with_underscore",
			input: "_temp",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "_temp"},
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
			name:  "simple_string",
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
			name:  "string_with_escaped_newline",
			input: `"line\n"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "line\n"},
			},
		},
		{
			name:  "string_with_escaped_tab",
			input: `"a\tb"`,
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenString, Value: "a\tb"},
			},
		},
		{
			name:  "record_string_assignment",
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
			name:  "leading_whitespace_then_identifier",
			input: "   foo",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 4}, Type: TokenIdentifier, Value: "foo"},
			},
		},
		{
			name:  "identifier_after_newline",
			input: "foo\nbar",
			want: []Token{
				{Pos: Pos{Line: 1, Column: 1}, Type: TokenIdentifier, Value: "foo"},
				{Pos: Pos{Line: 2, Column: 1}, Type: TokenIdentifier, Value: "bar"},
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
		_, err := collect(Tokenize(strings.NewReader("\"abc\nrest")))
		var ute *UnterminatedStringError
		require.ErrorAs(t, err, &ute)
		require.Equal(t, Pos{Line: 1, Column: 1}, ute.Pos)
	})

	t.Run("invalid_escape_sequence", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader(`"a\xb"`)))
		var ie *InvalidEscapeError
		require.ErrorAs(t, err, &ie)
	})

	t.Run("unexpected_character", func(t *testing.T) {
		t.Parallel()
		_, err := collect(Tokenize(strings.NewReader("@")))
		var uce *UnexpectedCharacterError
		require.ErrorAs(t, err, &uce)
		require.Equal(t, '@', uce.Char)
		require.Equal(t, Pos{Line: 1, Column: 1}, uce.Pos)
	})
}
