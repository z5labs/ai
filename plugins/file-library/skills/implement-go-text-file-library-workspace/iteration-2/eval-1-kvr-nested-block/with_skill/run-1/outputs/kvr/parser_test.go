package kvr

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  *File
	}{
		{
			name:  "empty_input_yields_zero_file",
			input: "",
			want:  &File{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(strings.NewReader(tc.input))
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParserTopLevelRecord(t *testing.T) {
	t.Parallel()

	got, err := Parse(strings.NewReader(`record string GREETING = "hello"`))
	require.NoError(t, err)
	require.Len(t, got.Records, 1)
	require.Equal(t, "string", got.Records[0].Type)
	require.Equal(t, "GREETING", got.Records[0].Key)
	require.Equal(t, "hello", got.Records[0].Value)
	require.Empty(t, got.Blocks)
}

func TestParserBlocks(t *testing.T) {
	t.Parallel()

	t.Run("empty_block", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`block COLORS { }`))
		require.NoError(t, err)
		require.Empty(t, got.Records)
		require.Len(t, got.Blocks, 1)
		require.Equal(t, "COLORS", got.Blocks[0].Name)
		require.Empty(t, got.Blocks[0].Records)
	})

	t.Run("block_with_one_record", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(`block COLORS { record string RED = "ff0000"; }`))
		require.NoError(t, err)
		require.Empty(t, got.Records)
		require.Len(t, got.Blocks, 1)
		require.Equal(t, "COLORS", got.Blocks[0].Name)
		require.Len(t, got.Blocks[0].Records, 1)
		require.Equal(t, "string", got.Blocks[0].Records[0].Type)
		require.Equal(t, "RED", got.Blocks[0].Records[0].Key)
		require.Equal(t, "ff0000", got.Blocks[0].Records[0].Value)
	})

	t.Run("block_with_two_records", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(
			`block COLORS { record string RED = "ff0000"; record string BLUE = "0000ff"; }`,
		))
		require.NoError(t, err)
		require.Empty(t, got.Records)
		require.Len(t, got.Blocks, 1)
		require.Equal(t, "COLORS", got.Blocks[0].Name)
		require.Len(t, got.Blocks[0].Records, 2)
		require.Equal(t, "RED", got.Blocks[0].Records[0].Key)
		require.Equal(t, "ff0000", got.Blocks[0].Records[0].Value)
		require.Equal(t, "BLUE", got.Blocks[0].Records[1].Key)
		require.Equal(t, "0000ff", got.Blocks[0].Records[1].Value)
	})

	t.Run("block_followed_by_top_level_record", func(t *testing.T) {
		t.Parallel()

		got, err := Parse(strings.NewReader(
			`block X { record string A = "1"; } record string B = "2"`,
		))
		require.NoError(t, err)
		require.Len(t, got.Blocks, 1)
		require.Len(t, got.Blocks[0].Records, 1)
		require.Len(t, got.Records, 1)
		require.Equal(t, "B", got.Records[0].Key)
	})
}
