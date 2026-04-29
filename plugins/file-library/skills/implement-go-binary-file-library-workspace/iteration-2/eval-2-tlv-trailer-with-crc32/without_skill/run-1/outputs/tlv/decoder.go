package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingReader wraps an io.Reader and tracks the number of bytes consumed.
// It also computes a running CRC32 over every byte it observes so the trailer
// check can compare against bytes preceding the trailer.
type countingReader struct {
	r    io.Reader
	n    int64
	hash hash.Hash32
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 {
		// Update running CRC over consumed bytes.
		_, _ = c.hash.Write(p[:n])
		c.n += int64(n)
	}
	return n, err
}

// decoder reads TLV1 structures from an underlying io.Reader.
type decoder struct {
	r         *countingReader
	byteOrder binary.ByteOrder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         &countingReader{r: r, hash: crc32.NewIEEE()},
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// All decode-time errors must go through this helper so the offset stays accurate.
func (d *decoder) wrapErr(field string, err error) error {
	if err == nil {
		return nil
	}
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readFile reads a TLV1 file from r.
//
// Currently only the trailer is implemented: every byte preceding the last 4
// bytes of input is treated as the body to be CRC32'd, and the final 4 bytes
// are the trailer's IEEE CRC32. On mismatch readFile returns an error wrapping
// ErrChecksumMismatch.
func (d *decoder) readFile() (*File, error) {
	// Read every byte so we can split body || trailer.
	all, err := io.ReadAll(d.r)
	if err != nil {
		return nil, d.wrapErr("File", err)
	}
	if len(all) < 4 {
		return nil, d.wrapErr("Trailer", ErrInvalid)
	}

	bodyLen := len(all) - 4
	body := all[:bodyLen]
	trailerBytes := all[bodyLen:]

	// At this point d.r.hash holds CRC32 of all bytes read (body + trailer).
	// We need CRC32 over body only — recompute since the running hash also
	// consumed the trailer. (Alternative: tee CRC writes to keep two hashes;
	// keep this simple while there's no header/record streaming yet.)
	want := crc32.ChecksumIEEE(body)
	got := d.byteOrder.Uint32(trailerBytes)

	if want != got {
		return nil, d.wrapErr("Trailer.CRC32", ErrChecksumMismatch)
	}

	return &File{Trailer: Trailer{CRC32: got}}, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
