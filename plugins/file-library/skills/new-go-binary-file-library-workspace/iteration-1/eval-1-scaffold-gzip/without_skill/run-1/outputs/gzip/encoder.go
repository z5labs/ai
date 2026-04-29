package gzip

import (
	"io"
)

// Encoder serializes a *File back to a gzip byte stream.
//
// This is a scaffold: the Encode method is not yet implemented.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns an Encoder that writes a gzip stream to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the given *File to the underlying writer as a gzip stream.
//
// TODO: implement serialization per RFC 1952:
//   - for each member, emit magic (0x1f 0x8b), CM, FLG, MTIME, XFL, OS
//   - if FlagExtra is set, emit XLEN and the extra sub-fields
//   - if FlagName is set, emit a NUL-terminated name
//   - if FlagComment is set, emit a NUL-terminated comment
//   - if FlagHCRC is set, compute and emit the CRC16 of the header bytes
//   - emit the compressed blocks
//   - emit CRC32 and ISIZE (both little-endian 4-byte)
func (e *Encoder) Encode(f *File) error {
	return errNotImplemented("Encoder.Encode")
}

// EncodeHeader writes only a gzip header to the underlying writer.
//
// TODO: implement per RFC 1952 section 2.3.1.
func (e *Encoder) EncodeHeader(h *Header) error {
	return errNotImplemented("Encoder.EncodeHeader")
}

// EncodeTrailer writes only an 8-byte gzip trailer (CRC32 + ISIZE) to the
// underlying writer.
//
// TODO: implement per RFC 1952 section 2.3.1.
func (e *Encoder) EncodeTrailer(t *Trailer) error {
	return errNotImplemented("Encoder.EncodeTrailer")
}
