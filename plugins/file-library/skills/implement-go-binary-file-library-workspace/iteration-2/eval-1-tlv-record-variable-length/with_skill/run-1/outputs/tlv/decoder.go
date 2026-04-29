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

// readFile reads records sequentially until the underlying reader is exhausted.
// (Header/Trailer wiring is deferred per the current implementation scope.)
func (d *decoder) readFile() (*File, error) {
	f := &File{}
	for {
		// Peek for end-of-stream by attempting to read the Type byte.
		var typeByte [1]byte
		_, err := io.ReadFull(d.r, typeByte[:])
		if errors.Is(err, io.EOF) {
			return f, nil
		}
		if err != nil {
			return nil, d.wrapErr("Record.Type", err)
		}

		rec := Record{Type: RecordType(typeByte[0])}

		if err := binary.Read(d.r, d.byteOrder, &rec.Length); err != nil {
			if errors.Is(err, io.EOF) {
				err = io.ErrUnexpectedEOF
			}
			return nil, d.wrapErr("Record.Length", err)
		}

		if rec.Length > 0 {
			rec.Value = make([]byte, rec.Length)
			if _, err := io.ReadFull(d.r, rec.Value); err != nil {
				if errors.Is(err, io.EOF) {
					err = io.ErrUnexpectedEOF
				}
				return nil, d.wrapErr("Record.Value", err)
			}
		}

		f.Records = append(f.Records, rec)
	}
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
