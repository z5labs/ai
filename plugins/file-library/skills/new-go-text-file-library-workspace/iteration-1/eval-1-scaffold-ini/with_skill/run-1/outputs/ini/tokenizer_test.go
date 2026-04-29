package ini

import (
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// collect drains an iter.Seq2[Token, error] into a slice of tokens and the
// first error (if any). It stops at the first error returned by the iterator.
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
		wantOK bool
	}{
		{
			name:   "empty input yields no tokens",
			input:  "",
			want:   nil,
			wantOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := collect(Tokenize(strings.NewReader(tc.input)))
			if tc.wantOK {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.Equal(t, tc.want, got)
		})
	}
}
