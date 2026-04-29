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

// writeFile writes every record in f to the underlying writer in order.
// (Header/Trailer wiring is deferred per the current implementation scope.)
func (e *encoder) writeFile(f *File) error {
	for i := range f.Records {
		if err := e.writeRecord(&f.Records[i]); err != nil {
			return err
		}
	}
	return nil
}

// writeRecord writes a single Record: 1-byte Type, 2-byte big-endian Length, Length bytes of Value.
func (e *encoder) writeRecord(rec *Record) error {
	if err := binary.Write(e.w, e.byteOrder, uint8(rec.Type)); err != nil {
		return e.wrapErr("Record.Type", err)
	}
	if err := binary.Write(e.w, e.byteOrder, rec.Length); err != nil {
		return e.wrapErr("Record.Length", err)
	}
	if rec.Length > 0 {
		if _, err := e.w.Write(rec.Value[:rec.Length]); err != nil {
			return e.wrapErr("Record.Value", err)
		}
	}
	return nil
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
