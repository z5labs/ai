package tlv

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Decoder reads TLV records from an underlying io.Reader.
//
// The default wire format mirrors Encoder:
//
//   - Type:   uint16, big-endian
//   - Length: uint32, big-endian
//   - Value:  Length raw bytes
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a Decoder that reads records from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads a single Record from the underlying reader.
//
// Decode returns io.EOF when the reader is at a clean record boundary and
// has no more data. A short read in the middle of a record is reported as
// ErrShortRead wrapping io.ErrUnexpectedEOF.
func (d *Decoder) Decode() (Record, error) {
	var hdr [6]byte
	n, err := io.ReadFull(d.r, hdr[:])
	if err != nil {
		if errors.Is(err, io.EOF) && n == 0 {
			return Record{}, io.EOF
		}
		return Record{}, fmt.Errorf("%w: header: %w", ErrShortRead, err)
	}

	t := Type(binary.BigEndian.Uint16(hdr[0:2]))
	length := binary.BigEndian.Uint32(hdr[2:6])

	rec := Record{Type: t}
	if length == 0 {
		return rec, nil
	}

	rec.Value = make([]byte, length)
	if _, err := io.ReadFull(d.r, rec.Value); err != nil {
		return Record{}, fmt.Errorf("%w: value: %w", ErrShortRead, err)
	}
	return rec, nil
}
