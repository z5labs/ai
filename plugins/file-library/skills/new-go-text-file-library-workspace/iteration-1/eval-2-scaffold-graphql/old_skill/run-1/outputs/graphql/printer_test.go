package graphql

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrinter(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
		want string
	}{
		{
			name: "empty file prints empty string",
			in:   &File{},
			want: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := Print(&buf, tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestPrinterRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "empty input round-trips",
			input: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := Parse(strings.NewReader(tc.input))
			require.NoError(t, err)

			var buf bytes.Buffer
			err = Print(&buf, parsed)
			require.NoError(t, err)

			reparsed, err := Parse(strings.NewReader(buf.String()))
			require.NoError(t, err)
			require.Equal(t, parsed, reparsed)
		})
	}
}
