package tlv

import (
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

// readFile reads zero or more Records from d.r until EOF.
//
// Header and Trailer are not yet implemented; this version of the decoder
// treats the entire stream as a sequence of Records.
func (d *decoder) readFile() (*File, error) {
	f := &File{}
	for {
		// Peek for a single Type byte. EOF here is the clean termination signal.
		var typeByte [1]byte
		_, err := io.ReadFull(d.r, typeByte[:])
		if errors.Is(err, io.EOF) {
			return f, nil
		}
		if err != nil {
			return nil, d.wrapErr("Record.Type", err)
		}

		rec, err := d.readRecordRest(RecordType(typeByte[0]))
		if err != nil {
			return nil, err
		}
		f.Records = append(f.Records, rec)
	}
}

// readRecordRest reads Length and Value once Type has already been consumed.
// Splitting the read keeps EOF-at-record-boundary distinguishable from a
// truncation mid-record.
func (d *decoder) readRecordRest(t RecordType) (Record, error) {
	var lenBuf [2]byte
	if _, err := io.ReadFull(d.r, lenBuf[:]); err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return Record{}, d.wrapErr("Record.Length", err)
	}
	length := d.byteOrder.Uint16(lenBuf[:])

	value := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(d.r, value); err != nil {
			if errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return Record{}, d.wrapErr("Record.Value", err)
		}
	}

	return Record{Type: t, Length: length, Value: value}, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
