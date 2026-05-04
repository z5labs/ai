package dsf

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"io"
)

// countingReader wraps an io.Reader and tracks the number of bytes consumed.
type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	if errors.Is(err, io.EOF) && n > 0 {
		err = io.ErrUnexpectedEOF
	}
	return n, err
}

// decoder reads DSF structures from a buffered slice of the input bytes. DSF
// requires the MD5 of every byte preceding the 16-byte footer to be re-checked
// against the footer, so the simplest correct implementation is to slurp the
// entire input first.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{r: &countingReader{r: r}, byteOrder: binary.LittleEndian}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
func (d *decoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readFile reads a complete DSF file: 12-byte header, atoms, 16-byte footer.
func (d *decoder) readFile() (*File, error) {
	// Slurp the whole input. DSF requires an MD5 over every byte preceding
	// the footer, so we need the trailing 16 bytes located before we can
	// decide where atoms stop.
	all, err := io.ReadAll(d.r)
	if err != nil {
		return nil, d.wrapErr("File", err)
	}
	if len(all) < 12+16 {
		return nil, d.wrapErr("File", io.ErrUnexpectedEOF)
	}

	hdr, err := d.readHeader(all[:12])
	if err != nil {
		return nil, err
	}

	atomsRegion := all[12 : len(all)-16]
	atoms, err := d.readAtoms(atomsRegion, 12)
	if err != nil {
		return nil, err
	}

	var footer [16]byte
	copy(footer[:], all[len(all)-16:])
	want := md5.Sum(all[:len(all)-16])
	if footer != want {
		// Position the offset at the start of the footer for clarity.
		d.r.n = int64(len(all) - 16)
		return nil, d.wrapErr("Footer.MD5", &MD5MismatchError{Got: footer, Want: want})
	}

	return &File{Header: *hdr, Atoms: atoms, Footer: footer}, nil
}

// readHeader decodes the 12-byte file header from the given byte slice. The
// byte slice must be exactly 12 bytes; the caller (readFile) is responsible
// for length-checking.
func (d *decoder) readHeader(b []byte) (*FileHeader, error) {
	var hdr FileHeader
	copy(hdr.Cookie[:], b[:8])
	if hdr.Cookie != MagicCookie {
		// Position the offset at the byte after the cookie, matching the
		// streaming consumer's view (cookie was just consumed).
		d.r.n = 8
		return nil, d.wrapErr("Header.Cookie", &UnexpectedCookieError{Got: hdr.Cookie})
	}
	hdr.Version = d.byteOrder.Uint32(b[8:12])
	if hdr.Version != CurrentVersion {
		d.r.n = 12
		return nil, d.wrapErr("Header.Version", &UnknownVersionError{Version: hdr.Version})
	}
	d.r.n = 12
	return &hdr, nil
}

// readAtoms walks a flat byte slice as a sequence of Atom records. baseOffset
// is the offset of the first byte of the slice within the original file (so
// d.r.n can be set correctly when an error occurs). The slice must contain
// only complete atoms — readFile passes only the bytes between the header and
// the footer.
func (d *decoder) readAtoms(b []byte, baseOffset int64) ([]Atom, error) {
	atoms := make([]Atom, 0)
	pos := 0
	for pos < len(b) {
		d.r.n = baseOffset + int64(pos)
		if len(b)-pos < 8 {
			return nil, d.wrapErr("Atoms", io.ErrUnexpectedEOF)
		}
		var a Atom
		a.ID = d.byteOrder.Uint32(b[pos : pos+4])
		a.Size = d.byteOrder.Uint32(b[pos+4 : pos+8])
		if a.Size < 8 {
			return nil, d.wrapErr("Atoms", &AtomSizeTooSmallError{Size: a.Size})
		}
		remaining := int64(len(b) - pos)
		if int64(a.Size) > remaining {
			return nil, d.wrapErr("Atoms", &AtomSizeOverflowError{Size: a.Size, Remaining: remaining})
		}
		// Copy the payload into a fresh slice so the caller cannot mutate
		// the caller's input by mutating Atom.Payload (and so encoder
		// round-trips do not alias bytes the user might still hold).
		payloadLen := int(a.Size) - 8
		a.Payload = make([]byte, payloadLen)
		copy(a.Payload, b[pos+8:pos+int(a.Size)])
		atoms = append(atoms, a)
		pos += int(a.Size)
	}
	d.r.n = baseOffset + int64(pos)
	return atoms, nil
}

// Decode reads a DSF file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}

