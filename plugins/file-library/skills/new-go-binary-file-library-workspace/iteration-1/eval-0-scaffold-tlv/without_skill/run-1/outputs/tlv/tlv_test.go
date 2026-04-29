package tlv_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"example.com/tlv-eval/tlv"
)

func TestEncoder_Encode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		record tlv.Record
		want   []byte
	}{
		{
			name:   "empty value",
			record: tlv.Record{Type: 0x0001},
			want:   []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:   "single byte value",
			record: tlv.Record{Type: 0x00FF, Value: []byte{0xAB}},
			want:   []byte{0x00, 0xFF, 0x00, 0x00, 0x00, 0x01, 0xAB},
		},
		{
			name:   "multi byte value",
			record: tlv.Record{Type: 0x1234, Value: []byte("hello")},
			want: []byte{
				0x12, 0x34,
				0x00, 0x00, 0x00, 0x05,
				'h', 'e', 'l', 'l', 'o',
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			enc := tlv.NewEncoder(&buf)

			err := enc.Encode(tt.record)
			require.NoError(t, err)
			require.Equal(t, tt.want, buf.Bytes())
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []byte
		want tlv.Record
	}{
		{
			name: "empty value",
			in:   []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			want: tlv.Record{Type: 0x0001},
		},
		{
			name: "single byte value",
			in:   []byte{0x00, 0xFF, 0x00, 0x00, 0x00, 0x01, 0xAB},
			want: tlv.Record{Type: 0x00FF, Value: []byte{0xAB}},
		},
		{
			name: "multi byte value",
			in: []byte{
				0x12, 0x34,
				0x00, 0x00, 0x00, 0x05,
				'h', 'e', 'l', 'l', 'o',
			},
			want: tlv.Record{Type: 0x1234, Value: []byte("hello")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dec := tlv.NewDecoder(bytes.NewReader(tt.in))

			got, err := dec.Decode()
			require.NoError(t, err)
			require.Equal(t, tt.want.Type, got.Type)
			require.Equal(t, tt.want.Value, got.Value)
		})
	}
}

func TestDecoder_Decode_EOF(t *testing.T) {
	t.Parallel()

	dec := tlv.NewDecoder(bytes.NewReader(nil))
	_, err := dec.Decode()
	require.ErrorIs(t, err, io.EOF)
}

func TestDecoder_Decode_ShortHeader(t *testing.T) {
	t.Parallel()

	dec := tlv.NewDecoder(bytes.NewReader([]byte{0x00, 0x01}))
	_, err := dec.Decode()
	require.Error(t, err)
	require.ErrorIs(t, err, tlv.ErrShortRead)
}

func TestDecoder_Decode_ShortValue(t *testing.T) {
	t.Parallel()

	// header announces 4 bytes of value but only 2 are present
	in := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0xAA, 0xBB}
	dec := tlv.NewDecoder(bytes.NewReader(in))
	_, err := dec.Decode()
	require.Error(t, err)
	require.ErrorIs(t, err, tlv.ErrShortRead)
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	records := []tlv.Record{
		{Type: 1, Value: []byte("alpha")},
		{Type: 2, Value: nil},
		{Type: 3, Value: bytes.Repeat([]byte{0xCD}, 64)},
	}

	var buf bytes.Buffer
	enc := tlv.NewEncoder(&buf)
	for _, r := range records {
		require.NoError(t, enc.Encode(r))
	}

	dec := tlv.NewDecoder(&buf)
	for _, want := range records {
		got, err := dec.Decode()
		require.NoError(t, err)
		require.Equal(t, want.Type, got.Type)
		if len(want.Value) == 0 {
			require.Empty(t, got.Value)
		} else {
			require.Equal(t, want.Value, got.Value)
		}
	}

	_, err := dec.Decode()
	require.True(t, errors.Is(err, io.EOF))
}

func TestRecord_Len(t *testing.T) {
	t.Parallel()

	require.Equal(t, 0, tlv.Record{}.Len())
	require.Equal(t, 3, tlv.Record{Value: []byte{1, 2, 3}}.Len())
}
