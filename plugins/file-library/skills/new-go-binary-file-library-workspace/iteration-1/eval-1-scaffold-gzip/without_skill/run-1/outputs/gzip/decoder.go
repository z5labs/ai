package gzip

import (
	"io"
)

// Decoder parses a gzip stream from an io.Reader into a *File.
//
// This is a scaffold: the Decode method is not yet implemented.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a Decoder that reads a gzip stream from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the entire gzip stream from the underlying reader and returns
// the parsed *File.
//
// TODO: implement parsing per RFC 1952:
//   - read and validate the magic bytes (0x1f 0x8b)
//   - read CM, FLG, MTIME (4 bytes, little-endian), XFL, OS
//   - if FLG.FEXTRA, read XLEN (2 bytes, little-endian) then XLEN bytes of subfields
//   - if FLG.FNAME, read a NUL-terminated name
//   - if FLG.FCOMMENT, read a NUL-terminated comment
//   - if FLG.FHCRC, read CRC16 (2 bytes, little-endian) over the header bytes
//   - read the deflate-compressed blocks (treated here as opaque bytes until
//     the trailer is located)
//   - read CRC32 (4 bytes, little-endian) and ISIZE (4 bytes, little-endian)
//   - repeat for each concatenated member until EOF
func (d *Decoder) Decode() (*File, error) {
	return nil, errNotImplemented("Decoder.Decode")
}

// DecodeHeader parses just the gzip header from the underlying reader.
//
// TODO: implement per RFC 1952 section 2.3.1.
func (d *Decoder) DecodeHeader() (*Header, error) {
	return nil, errNotImplemented("Decoder.DecodeHeader")
}

// DecodeTrailer parses the 8-byte gzip trailer (CRC32 + ISIZE) from the
// underlying reader.
//
// TODO: implement per RFC 1952 section 2.3.1.
func (d *Decoder) DecodeTrailer() (*Trailer, error) {
	return nil, errNotImplemented("Decoder.DecodeTrailer")
}
