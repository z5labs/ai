package tlv

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeStubReturnsErrUnimplemented(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	err := Encode(&buf, &File{})
	require.ErrorIs(t, err, errUnimplemented)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "File", fe.Field)
}

func TestRoundTripFromTestdata(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		fixture string
	}{
		{name: "sample_tlv", fixture: "sample.tlv"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join("testdata", tc.fixture))
			require.NoError(t, err)

			f, err := Decode(bytes.NewReader(data))
			require.NoError(t, err)

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, f))
			require.Equal(t, data, buf.Bytes())
		})
	}
}
