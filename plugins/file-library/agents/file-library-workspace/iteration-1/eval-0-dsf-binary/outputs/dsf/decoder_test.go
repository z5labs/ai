package dsf

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// buildValidDSF returns a syntactically valid DSF made of: header, the given
// atoms (each as raw {ID, Size, Payload} bytes), and a correct MD5 footer.
func buildValidDSF(atoms ...Atom) []byte {
	var buf bytes.Buffer
	buf.Write(MagicCookie[:])
	_ = binary.Write(&buf, binary.LittleEndian, CurrentVersion)
	for _, a := range atoms {
		_ = binary.Write(&buf, binary.LittleEndian, a.ID)
		_ = binary.Write(&buf, binary.LittleEndian, a.Size)
		buf.Write(a.Payload)
	}
	sum := md5.Sum(buf.Bytes())
	buf.Write(sum[:])
	return buf.Bytes()
}

func TestDecodeMinimalEmpty(t *testing.T) {
	t.Parallel()

	in := buildValidDSF()
	require.Len(t, in, 12+16)

	got, err := Decode(bytes.NewReader(in))
	require.NoError(t, err)
	require.Equal(t, MagicCookie, got.Header.Cookie)
	require.Equal(t, CurrentVersion, got.Header.Version)
	require.Empty(t, got.Atoms)
	var wantFooter [16]byte
	copy(wantFooter[:], in[12:])
	require.Equal(t, wantFooter, got.Footer)
}

func TestDecodeOneEmptyAtom(t *testing.T) {
	t.Parallel()

	atom := Atom{ID: 0x44414548, Size: 8, Payload: []byte{}} // 'HEAD' in LE
	in := buildValidDSF(atom)
	got, err := Decode(bytes.NewReader(in))
	require.NoError(t, err)
	require.Len(t, got.Atoms, 1)
	require.Equal(t, uint32(0x44414548), got.Atoms[0].ID)
	require.Equal(t, uint32(8), got.Atoms[0].Size)
	require.Empty(t, got.Atoms[0].Payload)
}

func TestDecodeAtomWithPayload(t *testing.T) {
	t.Parallel()

	payload := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	atom := Atom{ID: 0x44414548, Size: 8 + uint32(len(payload)), Payload: payload}
	in := buildValidDSF(atom)
	got, err := Decode(bytes.NewReader(in))
	require.NoError(t, err)
	require.Len(t, got.Atoms, 1)
	require.Equal(t, payload, got.Atoms[0].Payload)
	require.Equal(t, uint32(12), got.Atoms[0].Size)
}

func TestDecodeWrongCookie(t *testing.T) {
	t.Parallel()

	in := buildValidDSF()
	in[0] = 'B' // corrupt cookie
	// Footer no longer matches; but we must surface the cookie error first
	// because cookie validation happens before MD5 verification on a
	// streaming decoder. (Our implementation slurps and validates header
	// before MD5 too.)
	_, err := Decode(bytes.NewReader(in))
	require.Error(t, err)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Cookie", fe.Field)

	var oe *OffsetError
	require.ErrorAs(t, err, &oe)
	require.Equal(t, int64(8), oe.Offset)

	var ce *UnexpectedCookieError
	require.ErrorAs(t, err, &ce)
	require.Equal(t, byte('B'), ce.Got[0])
}

func TestDecodeWrongVersion(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	buf.Write(MagicCookie[:])
	_ = binary.Write(&buf, binary.LittleEndian, uint32(2)) // bad version
	sum := md5.Sum(buf.Bytes())
	buf.Write(sum[:])

	_, err := Decode(bytes.NewReader(buf.Bytes()))
	require.Error(t, err)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Header.Version", fe.Field)

	var ve *UnknownVersionError
	require.ErrorAs(t, err, &ve)
	require.Equal(t, uint32(2), ve.Version)
}

func TestDecodeAtomSizeTooSmall(t *testing.T) {
	t.Parallel()

	// Build a file with one bogus atom whose Size is 4 (< 8). We have to
	// hand-roll the bytes because buildValidDSF's payload sizing depends
	// on Size being self-consistent.
	var buf bytes.Buffer
	buf.Write(MagicCookie[:])
	_ = binary.Write(&buf, binary.LittleEndian, CurrentVersion)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x44414548)) // 'HEAD'
	_ = binary.Write(&buf, binary.LittleEndian, uint32(4))          // bogus Size
	sum := md5.Sum(buf.Bytes())
	buf.Write(sum[:])

	_, err := Decode(bytes.NewReader(buf.Bytes()))
	require.Error(t, err)

	var fe *FieldError
	require.ErrorAs(t, err, &fe)
	require.Equal(t, "Atoms", fe.Field)

	var astse *AtomSizeTooSmallError
	require.ErrorAs(t, err, &astse)
	require.Equal(t, uint32(4), astse.Size)
}

func TestDecodeAtomSizeOverflow(t *testing.T) {
	t.Parallel()

	// One atom whose Size is larger than the bytes available before the
	// footer.
	var buf bytes.Buffer
	buf.Write(MagicCookie[:])
	_ = binary.Write(&buf, binary.LittleEndian, CurrentVersion)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x44414548)) // 'HEAD'
	_ = binary.Write(&buf, binary.LittleEndian, uint32(100))        // claims 100 bytes
	// only emit 4 payload bytes
	buf.Write([]byte{1, 2, 3, 4})
	sum := md5.Sum(buf.Bytes())
	buf.Write(sum[:])

	_, err := Decode(bytes.NewReader(buf.Bytes()))
	require.Error(t, err)

	var aoe *AtomSizeOverflowError
	require.ErrorAs(t, err, &aoe)
	require.Equal(t, uint32(100), aoe.Size)
}

func TestDecodeMD5Mismatch(t *testing.T) {
	t.Parallel()

	in := buildValidDSF()
	// Flip a footer byte so the recorded hash no longer matches.
	in[len(in)-1] ^= 0xFF
	_, err := Decode(bytes.NewReader(in))
	require.Error(t, err)

	var me *MD5MismatchError
	require.ErrorAs(t, err, &me)
	require.NotEqual(t, me.Got, me.Want)
}

func TestDecodeTruncated(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   []byte
	}{
		{"empty", []byte{}},
		{"one_byte", []byte{0x58}},
		{"shorter_than_header_plus_footer", append([]byte("XPLNEDSF"), 0x01, 0x00, 0x00, 0x00)},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := Decode(bytes.NewReader(tc.in))
			require.Error(t, err)
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		})
	}
}
