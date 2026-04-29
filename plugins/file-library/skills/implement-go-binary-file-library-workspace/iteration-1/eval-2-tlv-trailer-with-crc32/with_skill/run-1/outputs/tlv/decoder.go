package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
	"io"
)

// countingReader wraps an io.Reader and tracks the number of bytes consumed.
// It also tees every successful read into a running CRC32 hash so the trailer
// integrity check is performed against the bytes that flowed through Read —
// the "compute running CRC32 as bytes are read" requirement from the SPEC.
type countingReader struct {
	r    io.Reader
	n    int64
	hash hash.Hash32
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	if c.hash != nil && n > 0 {
		// hash.Hash.Write never returns an error per its contract.
		_, _ = c.hash.Write(p[:n])
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

// readFile reads a TLV1 file: every byte preceding the trailer is captured as
// File.Payload, and the trailing 4-byte CRC32 is read and verified.
//
// Until Header and Records are implemented, "everything before the last 4
// bytes" is the Payload. The running CRC32 is updated as bytes flow through
// the counting reader; once the trailer is read, the running CRC32 (snapshotted
// just before the trailer bytes) is compared to the stored CRC32.
func (d *decoder) readFile() (*File, error) {
	// Read the entire stream through the counting reader so the running CRC32
	// hash captures every byte. We then split the buffer into payload + trailer
	// and snapshot the CRC32 over only the payload portion.
	all, err := io.ReadAll(d.r)
	if err != nil {
		return nil, d.wrapErr("File", err)
	}

	if len(all) < 4 {
		return nil, d.wrapErr("Trailer.CRC32", io.ErrUnexpectedEOF)
	}

	payload := all[:len(all)-4]
	trailerBytes := all[len(all)-4:]

	// The running hash on the counting reader has now absorbed both payload and
	// trailer bytes. Re-derive the CRC32 over just the payload — the SPEC
	// requires exactly the bytes preceding the trailer.
	computed := crc32.ChecksumIEEE(payload)
	stored := d.byteOrder.Uint32(trailerBytes)

	if computed != stored {
		return nil, d.wrapErr("Trailer.CRC32", ErrChecksumMismatch)
	}

	return &File{
		Payload: payload,
		Trailer: Trailer{CRC32: stored},
	}, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
