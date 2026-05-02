package tlv

import (
	"encoding/binary"
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
	return n, err
}

// decoder reads TLV1 structures from an underlying io.Reader.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{r: &countingReader{r: r}, byteOrder: binary.BigEndian}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// Always called inside an `if err != nil` branch, so it doesn't guard against
// nil errors itself.
func (d *decoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readRecord reads a single TLV1 record from the underlying reader. The byte
// layout is:
//
//	[Type:1][Length:2 BE][Value:Length]
//
// A Length of 0 is legal and yields a zero-length (but non-nil) Value slice.
func (d *decoder) readRecord() (Record, error) {
	var rec Record

	var typeBuf [1]byte
	if _, err := io.ReadFull(d.r, typeBuf[:]); err != nil {
		return rec, d.wrapErr("Record.Type", err)
	}
	rec.Type = RecordType(typeBuf[0])

	var lenBuf [2]byte
	if _, err := io.ReadFull(d.r, lenBuf[:]); err != nil {
		return rec, d.wrapErr("Record.Length", err)
	}
	rec.Length = d.byteOrder.Uint16(lenBuf[:])

	rec.Value = make([]byte, rec.Length)
	if rec.Length > 0 {
		if _, err := io.ReadFull(d.r, rec.Value); err != nil {
			return rec, d.wrapErr("Record.Value", err)
		}
	}

	return rec, nil
}

func (d *decoder) readFile() (*File, error) {
	return nil, d.wrapErr("File", errUnimplemented)
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}

// DecodeRecord reads a single TLV1 Record from r.
func DecodeRecord(r io.Reader) (Record, error) {
	return newDecoder(r).readRecord()
}
