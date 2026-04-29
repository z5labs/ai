package tlv

import (
	"encoding/binary"
	"fmt"
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
// All decode-time errors must go through this helper so the offset stays accurate.
func (d *decoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readRecord decodes a single TLV1 record from the underlying stream.
//
// The returned record has its Type, Length, and Value populated; the Value
// slice is freshly allocated and owned by the caller.
func (d *decoder) readRecord() (Record, error) {
	var rec Record

	var typeBuf [1]byte
	if _, err := io.ReadFull(d.r, typeBuf[:]); err != nil {
		return Record{}, d.wrapErr("Record.Type", err)
	}
	rec.Type = RecordType(typeBuf[0])

	if !isKnownRecordType(rec.Type) {
		return Record{}, d.wrapErr("Record.Type",
			fmt.Errorf("%w: 0x%02X", ErrUnknownRecordType, uint8(rec.Type)))
	}

	var lenBuf [2]byte
	if _, err := io.ReadFull(d.r, lenBuf[:]); err != nil {
		return Record{}, d.wrapErr("Record.Length", err)
	}
	rec.Length = d.byteOrder.Uint16(lenBuf[:])

	if rec.Type == RecordTypeINT && rec.Length != 8 {
		return Record{}, d.wrapErr("Record.Length",
			fmt.Errorf("%w: INT record must have Length=8, got %d", ErrInvalid, rec.Length))
	}

	if rec.Length > 0 {
		rec.Value = make([]byte, rec.Length)
		if _, err := io.ReadFull(d.r, rec.Value); err != nil {
			return Record{}, d.wrapErr("Record.Value", err)
		}
	}

	return rec, nil
}

func isKnownRecordType(t RecordType) bool {
	switch t {
	case RecordTypeSTRING, RecordTypeINT, RecordTypeBLOB, RecordTypeNESTED:
		return true
	default:
		return false
	}
}

func (d *decoder) readFile() (*File, error) {
	return nil, d.wrapErr("File", errUnimplemented)
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
