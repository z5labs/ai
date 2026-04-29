package tlv

import (
	"encoding/binary"
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

// encoder writes TLV1 structures to an underlying io.Writer.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{w: &countingWriter{w: w}, byteOrder: binary.BigEndian}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
func (e *encoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}
}

// writeFile writes every Record in f to the underlying writer.
//
// Header and Trailer are not yet implemented; this version of the encoder
// emits the records back-to-back with no padding.
func (e *encoder) writeFile(f *File) error {
	for i := range f.Records {
		if err := e.writeRecord(&f.Records[i]); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) writeRecord(r *Record) error {
	if _, err := e.w.Write([]byte{byte(r.Type)}); err != nil {
		return e.wrapErr("Record.Type", err)
	}

	var lenBuf [2]byte
	e.byteOrder.PutUint16(lenBuf[:], r.Length)
	if _, err := e.w.Write(lenBuf[:]); err != nil {
		return e.wrapErr("Record.Length", err)
	}

	if r.Length > 0 {
		if _, err := e.w.Write(r.Value[:r.Length]); err != nil {
			return e.wrapErr("Record.Value", err)
		}
	}
	return nil
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
