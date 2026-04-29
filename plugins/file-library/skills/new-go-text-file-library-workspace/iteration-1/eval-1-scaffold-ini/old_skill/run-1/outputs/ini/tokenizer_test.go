package ini

import (
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// collect drains an iter.Seq2[Token, error] into a token slice. If the
// iterator yields an error, collect returns it and the tokens accumulated
// up to that point.
func collect(seq iter.Seq2[Token, error]) ([]Token, error) {
	var tokens []Token
	var firstErr error
	seq(func(tok Token, err error) bool {
		if err != nil {
			firstErr = err
			return false
		}
		tokens = append(tokens, tok)
		return true
	})
	return tokens, firstErr
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
