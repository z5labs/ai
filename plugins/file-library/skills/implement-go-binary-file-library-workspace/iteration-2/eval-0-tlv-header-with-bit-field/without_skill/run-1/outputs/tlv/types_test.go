package tlv

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorChain(t *testing.T) {
	t.Parallel()

	err := &FieldError{Field: "File", Err: &OffsetError{Offset: 0, Err: errUnimplemented}}

	require.ErrorIs(t, err, errUnimplemented)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "File", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(0), oe.Offset)

	require.True(t, errors.Is(err, errUnimplemented))
}

func TestFlagsAccessors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		flags      Flags
		compressed bool
		encrypted  bool
		signed     bool
	}{
		{"none", 0x00, false, false, false},
		{"compressed", FlagCompressed, true, false, false},
		{"encrypted", FlagEncrypted, false, true, false},
		{"signed", FlagSigned, false, false, true},
		{"compressed+encrypted", FlagCompressed | FlagEncrypted, true, true, false},
		{"all", FlagCompressed | FlagEncrypted | FlagSigned, true, true, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.compressed, tt.flags.Compressed())
			require.Equal(t, tt.encrypted, tt.flags.Encrypted())
			require.Equal(t, tt.signed, tt.flags.Signed())
		})
	}
}
