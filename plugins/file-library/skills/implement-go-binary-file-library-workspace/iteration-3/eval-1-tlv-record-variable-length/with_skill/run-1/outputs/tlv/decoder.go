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

func (d *decoder) readFile() (*File, error) {
	return nil, d.wrapErr("File", errUnimplemented)
}

// readRecord reads a single Type-Length-Value record from the underlying
// reader. The wire layout is:
//   - 1 byte Type
//   - 2 bytes Length (big-endian uint16)
//   - Length bytes Value
func (d *decoder) readRecord() (Record, error) {
	var rec Record

	var typeByte uint8
	if err := binary.Read(d.r, d.byteOrder, &typeByte); err != nil {
		return rec, d.wrapErr("Record.Type", err)
	}
	rt := RecordType(typeByte)
	switch rt {
	case RecordTypeSTRING, RecordTypeINT, RecordTypeBLOB, RecordTypeNESTED:
		// known
	default:
		return rec, d.wrapErr("Record.Type", ErrUnknownRecordType)
	}
	rec.Type = rt

	var length uint16
	if err := binary.Read(d.r, d.byteOrder, &length); err != nil {
		// binary.Read maps a partial read to io.ErrUnexpectedEOF for fixed-size
		// values; an EOF before any bytes maps to io.EOF.
		return rec, d.wrapErr("Record.Length", err)
	}
	rec.Length = length

	value := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(d.r, value); err != nil {
			return rec, d.wrapErr("Record.Value", err)
		}
	}
	rec.Value = value

	return rec, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
