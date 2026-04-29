package graphql

import (
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// collect drains an iter.Seq2[Token, error] into a slice of tokens. If the
// iterator yields a non-nil error, collect returns the tokens gathered so far
// along with that error.
func collect(seq iter.Seq2[Token, error]) ([]Token, error) {
	var tokens []Token
	for tok, err := range seq {
		if err != nil {
			return tokens, err
		}
		tokens = append(tokens, tok)
	}
	return tokens, nil
}

func TestTokenizer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  string
		want   []Token
		errMsg string
	}{
		{
			name:  "empty input yields no tokens",
			input: "",
			want:  nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := collect(Tokenize(strings.NewReader(tc.input)))
			if tc.errMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
