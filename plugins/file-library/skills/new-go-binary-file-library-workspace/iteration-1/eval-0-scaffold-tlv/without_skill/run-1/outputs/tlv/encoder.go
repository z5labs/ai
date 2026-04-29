package tlv

import (
	"encoding/binary"
	"io"
)

// Encoder writes TLV records to an underlying io.Writer.
//
// The default wire format is:
//
//   - Type:   uint16, big-endian
//   - Length: uint32, big-endian
//   - Value:  Length raw bytes
//
// Adjust the constants and Encode method to match your TLV variant.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns an Encoder that writes records to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes a single Record to the underlying writer.
func (e *Encoder) Encode(r Record) error {
	var hdr [6]byte
	binary.BigEndian.PutUint16(hdr[0:2], uint16(r.Type))
	binary.BigEndian.PutUint32(hdr[2:6], uint32(len(r.Value)))

	if _, err := e.w.Write(hdr[:]); err != nil {
		return err
	}
	if len(r.Value) == 0 {
		return nil
	}
	_, err := e.w.Write(r.Value)
	return err
}
