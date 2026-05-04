package dsf

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeMinimalEmpty(t *testing.T) {
	t.Parallel()

	in := &File{
		Header: FileHeader{Cookie: MagicCookie, Version: CurrentVersion},
	}

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, in))

	require.Equal(t, 12+16, buf.Len())
	require.Equal(t, MagicCookie[:], buf.Bytes()[:8])
	require.Equal(t, CurrentVersion, binary.LittleEndian.Uint32(buf.Bytes()[8:12]))

	// Footer must equal the MD5 of all preceding bytes.
	want := md5.Sum(buf.Bytes()[:12])
	require.Equal(t, want[:], buf.Bytes()[12:])
}

func TestEncodeOneAtom(t *testing.T) {
	t.Parallel()

	payload := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	in := &File{
		Header: FileHeader{Cookie: MagicCookie, Version: CurrentVersion},
		Atoms: []Atom{
			{ID: 0x44414548, Size: 8 + uint32(len(payload)), Payload: payload},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, Encode(&buf, in))

	require.Equal(t, 12+12+16, buf.Len())
	want := md5.Sum(buf.Bytes()[:12+12])
	require.Equal(t, want[:], buf.Bytes()[12+12:])
}

func TestEncodeRejectsBadVersion(t *testing.T) {
	t.Parallel()

	in := &File{Header: FileHeader{Cookie: MagicCookie, Version: 99}}
	var buf bytes.Buffer
	err := Encode(&buf, in)
	require.Error(t, err)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Version", fe.Field)

	var ve *UnknownVersionError
	require.ErrorAs(t, err, &ve)
	require.Equal(t, uint32(99), ve.Version)
}

func TestEncodeRejectsBadCookie(t *testing.T) {
	t.Parallel()

	in := &File{Header: FileHeader{Cookie: [8]byte{'B', 'A', 'D', 'M', 'A', 'G', 'I', 'C'}, Version: 1}}
	var buf bytes.Buffer
	err := Encode(&buf, in)
	require.Error(t, err)

	var ce *UnexpectedCookieError
	require.ErrorAs(t, err, &ce)
	require.Equal(t, byte('B'), ce.Got[0])
}

func TestEncodeRejectsAtomSizePayloadMismatch(t *testing.T) {
	t.Parallel()

	in := &File{
		Header: FileHeader{Cookie: MagicCookie, Version: CurrentVersion},
		Atoms: []Atom{
			// Size claims 100, payload is only 4 bytes.
			{ID: 0x44414548, Size: 100, Payload: []byte{1, 2, 3, 4}},
		},
	}
	var buf bytes.Buffer
	err := Encode(&buf, in)
	require.Error(t, err)

	var ao *AtomSizeOverflowError
	require.ErrorAs(t, err, &ao)
	require.Equal(t, uint32(100), ao.Size)
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *File
	}{
		{
			name: "empty",
			in:   &File{Header: FileHeader{Cookie: MagicCookie, Version: CurrentVersion}},
		},
		{
			name: "one_empty_atom",
			in: &File{
				Header: FileHeader{Cookie: MagicCookie, Version: CurrentVersion},
				Atoms: []Atom{
					{ID: 0x44414548, Size: 8, Payload: []byte{}},
				},
			},
		},
		{
			name: "one_atom_with_payload",
			in: &File{
				Header: FileHeader{Cookie: MagicCookie, Version: CurrentVersion},
				Atoms: []Atom{
					{ID: 0x44414548, Size: 12, Payload: []byte{0x01, 0x02, 0x03, 0x04}},
				},
			},
		},
		{
			name: "multiple_atoms",
			in: &File{
				Header: FileHeader{Cookie: MagicCookie, Version: CurrentVersion},
				Atoms: []Atom{
					{ID: 0x44414548, Size: 10, Payload: []byte{0xAA, 0xBB}},
					{ID: 0x47454F44, Size: 11, Payload: []byte{0xCC, 0xDD, 0xEE}},
					{ID: 0x434D4453, Size: 8, Payload: []byte{}},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			require.NoError(t, Encode(&buf, tc.in))
			data := buf.Bytes()

			out, err := Decode(bytes.NewReader(data))
			require.NoError(t, err)

			// Footer in `tc.in` is zero; on round trip the decoded File
			// carries the freshly computed footer. Compare everything else.
			require.Equal(t, tc.in.Header, out.Header)
			require.Equal(t, len(tc.in.Atoms), len(out.Atoms))
			for i := range tc.in.Atoms {
				require.Equal(t, tc.in.Atoms[i].ID, out.Atoms[i].ID)
				require.Equal(t, tc.in.Atoms[i].Size, out.Atoms[i].Size)
				require.Equal(t, tc.in.Atoms[i].Payload, out.Atoms[i].Payload)
			}

			// Re-encoding the decoded file must produce byte-identical output.
			var buf2 bytes.Buffer
			require.NoError(t, Encode(&buf2, out))
			require.Equal(t, data, buf2.Bytes())
		})
	}
}

func TestRoundTripFromTestdata(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		fixture string
	}{
		{name: "dsf_real_world_dsf", fixture: "dsf-real-world.dsf"},
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
