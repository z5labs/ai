// Copyright (c) 2026 z5labs
//
// Licensed under the MIT License (the "License").
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT

package toml

import (
	"iter"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// collect drains an iter.Seq2[Token, error] into parallel slices. It stops
// at the first error and returns whatever tokens were produced before it.
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

	cases := []struct {
		name   string
		input  string
		want   []Token
		hasErr bool
	}{
		{
			name:  "empty input yields no tokens",
			input: "",
			want:  nil,
		},
	}

	for _, tc := range cases {
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
