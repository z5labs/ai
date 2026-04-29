package graphql

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestParser exercises the public Parse() surface. Per the package
// CLAUDE.md, parser tests must drive Parse() with real source strings and
// never construct AST nodes by hand.
func TestParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		input  string
		want   *File
		hasErr bool
	}{
		{
			name:  "empty input parses to an empty File",
			input: "",
			want:  &File{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(strings.NewReader(tc.input))
			if tc.hasErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
