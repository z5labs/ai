package graphql

import (
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// collect drains an iter.Seq2[Token, error] into a slice of tokens plus
// the first error observed (if any). It is the standard helper for
// tokenizer tests.
func collect(seq iter.Seq2[Token, error]) ([]Token, error) {
	var tokens []Token
	var firstErr error
	for tok, err := range seq {
		if err != nil {
			firstErr = err
			break
		}
		tokens = append(tokens, tok)
	}
	return tokens, firstErr
}

func TestTokenizer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  string
		want   []Token
		hasErr bool
	}{
		{
			name:  "empty input yields no tokens and no error",
			input: "",
			want:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := collect(Tokenize(strings.NewReader(tc.input)))
			if tc.hasErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
