package tlv

import (
	"encoding/binary"
	"fmt"
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

// writeRecord encodes a single TLV1 record to the underlying stream.
//
// The Length field is taken from the record as written by the caller; it
// must equal len(rec.Value). For an INT record, Length must be 8.
func (e *encoder) writeRecord(rec Record) error {
	if !isKnownRecordType(rec.Type) {
		return e.wrapErr("Record.Type",
			fmt.Errorf("%w: 0x%02X", ErrUnknownRecordType, uint8(rec.Type)))
	}

	if int(rec.Length) != len(rec.Value) {
		return e.wrapErr("Record.Length",
			fmt.Errorf("%w: Length=%d does not match len(Value)=%d",
				ErrInvalid, rec.Length, len(rec.Value)))
	}

	if rec.Type == RecordTypeINT && rec.Length != 8 {
		return e.wrapErr("Record.Length",
			fmt.Errorf("%w: INT record must have Length=8, got %d", ErrInvalid, rec.Length))
	}

	if _, err := e.w.Write([]byte{byte(rec.Type)}); err != nil {
		return e.wrapErr("Record.Type", err)
	}

	var lenBuf [2]byte
	e.byteOrder.PutUint16(lenBuf[:], rec.Length)
	if _, err := e.w.Write(lenBuf[:]); err != nil {
		return e.wrapErr("Record.Length", err)
	}

	if rec.Length > 0 {
		if _, err := e.w.Write(rec.Value); err != nil {
			return e.wrapErr("Record.Value", err)
		}
	}

	return nil
}

func (e *encoder) writeFile(f *File) error {
	return e.wrapErr("File", errUnimplemented)
}

// Encode writes f to w as a TLV1 file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
