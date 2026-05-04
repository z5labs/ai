package dsf

import (
	"crypto/md5"
	"encoding/binary"
	"hash"
	"io"
)

// countingWriter wraps an io.Writer and tracks the number of bytes written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

// encoder writes DSF structures to an underlying io.Writer. The MD5 footer is
// computed by tee-writing every byte through an md5 hasher and then writing
// the hash sum at the end.
type encoder struct {
	w         *countingWriter
	hasher    hash.Hash
	teed      io.Writer // multi-writer over (w, hasher)
	byteOrder binary.ByteOrder
}

func newEncoder(w io.Writer) *encoder {
	cw := &countingWriter{w: w}
	h := md5.New()
	return &encoder{
		w:         cw,
		hasher:    h,
		teed:      io.MultiWriter(cw, h),
		byteOrder: binary.LittleEndian,
	}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
func (e *encoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}
}

// writeFile writes a complete DSF file: header, atoms, MD5 footer.
func (e *encoder) writeFile(f *File) error {
	if err := e.writeHeader(&f.Header); err != nil {
		return err
	}
	for i := range f.Atoms {
		if err := e.writeAtom(&f.Atoms[i]); err != nil {
			return err
		}
	}
	return e.writeFooter()
}

func (e *encoder) writeHeader(h *FileHeader) error {
	cookie := h.Cookie
	if cookie == ([8]byte{}) {
		// Default a zero cookie to MagicCookie so a caller can construct
		// File{} values ergonomically without having to copy the magic by hand.
		cookie = MagicCookie
	}
	if cookie != MagicCookie {
		return e.wrapErr("Header.Cookie", &UnexpectedCookieError{Got: cookie})
	}
	version := h.Version
	if version == 0 {
		version = CurrentVersion
	}
	if version != CurrentVersion {
		return e.wrapErr("Header.Version", &UnknownVersionError{Version: version})
	}
	if _, err := e.teed.Write(cookie[:]); err != nil {
		return e.wrapErr("Header.Cookie", err)
	}
	var vbuf [4]byte
	e.byteOrder.PutUint32(vbuf[:], version)
	if _, err := e.teed.Write(vbuf[:]); err != nil {
		return e.wrapErr("Header.Version", err)
	}
	return nil
}

func (e *encoder) writeAtom(a *Atom) error {
	expected := uint32(len(a.Payload)) + 8
	size := a.Size
	if size == 0 {
		size = expected
	}
	if size < 8 {
		return e.wrapErr("Atoms", &AtomSizeTooSmallError{Size: size})
	}
	if size != expected {
		// Caller-supplied Size disagrees with payload length: refuse rather
		// than silently masking a programming error.
		return e.wrapErr("Atoms", &AtomSizeOverflowError{Size: size, Remaining: int64(expected)})
	}
	var hdr [8]byte
	e.byteOrder.PutUint32(hdr[0:4], a.ID)
	e.byteOrder.PutUint32(hdr[4:8], size)
	if _, err := e.teed.Write(hdr[:]); err != nil {
		return e.wrapErr("Atoms", err)
	}
	if len(a.Payload) > 0 {
		if _, err := e.teed.Write(a.Payload); err != nil {
			return e.wrapErr("Atoms", err)
		}
	}
	return nil
}

func (e *encoder) writeFooter() error {
	var sum [16]byte
	copy(sum[:], e.hasher.Sum(nil))
	// Footer goes through the underlying writer only — it is not part of
	// the MD5 itself.
	if _, err := e.w.Write(sum[:]); err != nil {
		return e.wrapErr("Footer.MD5", err)
	}
	return nil
}

// Encode writes f to w as a DSF file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
