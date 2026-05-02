package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingWriter wraps an io.Writer and tracks the number of bytes written.
// It also feeds every successfully-written byte into a running CRC32 hash so
// the encoder can emit the trailer without buffering the whole file.
type countingWriter struct {
	w           io.Writer
	n           int64
	hash        hash.Hash32
	hashEnabled bool
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	if c.hashEnabled && n > 0 {
		_, _ = c.hash.Write(p[:n])
	}
	return n, err
}

// encoder writes TLV1 structures to an underlying io.Writer.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w: &countingWriter{
			w:           w,
			hash:        crc32.NewIEEE(),
			hashEnabled: true,
		},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every error site into the FieldError -> OffsetError -> leaf chain.
// Always called inside an `if err != nil` branch.
func (e *encoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}
}

func (e *encoder) writeAll(field string, p []byte) error {
	if _, err := e.w.Write(p); err != nil {
		return e.wrapErr(field, err)
	}
	return nil
}

func (e *encoder) writeHeader(h Header) error {
	if h.Magic == ([4]byte{}) {
		// Default to "TLV1" if caller left it zero so callers don't have to
		// remember the magic bytes for the happy path.
		h.Magic = [4]byte{'T', 'L', 'V', '1'}
	}
	if h.Version == 0 {
		h.Version = 1
	}
	var buf [8]byte
	copy(buf[0:4], h.Magic[:])
	buf[4] = h.Version
	buf[5] = h.Flags
	e.byteOrder.PutUint16(buf[6:8], h.Reserved)
	return e.writeAll("Header", buf[:])
}

func (e *encoder) writeRecord(r Record) error {
	length := r.Length
	if length == 0 && len(r.Value) > 0 {
		length = uint16(len(r.Value))
	}
	var hdr [3]byte
	hdr[0] = r.Type
	e.byteOrder.PutUint16(hdr[1:3], length)
	if err := e.writeAll("Record", hdr[:]); err != nil {
		return err
	}
	if length > 0 {
		if err := e.writeAll("Record.Value", r.Value[:length]); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) writeTrailer() error {
	// Snapshot CRC of every byte written so far, then disable hashing while
	// emitting the trailer itself.
	sum := e.w.hash.Sum32()
	e.w.hashEnabled = false
	var buf [4]byte
	e.byteOrder.PutUint32(buf[:], sum)
	return e.writeAll("Trailer", buf[:])
}

func (e *encoder) writeFile(f *File) error {
	if f == nil {
		return e.wrapErr("File", ErrInvalid)
	}
	if err := e.writeHeader(f.Header); err != nil {
		return err
	}
	for _, r := range f.Records {
		if err := e.writeRecord(r); err != nil {
			return err
		}
	}
	return e.writeTrailer()
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
