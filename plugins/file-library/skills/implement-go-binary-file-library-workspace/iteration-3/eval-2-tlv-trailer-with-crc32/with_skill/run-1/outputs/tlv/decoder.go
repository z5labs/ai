package tlv

import (
	"encoding/binary"
	"hash"
	"hash/crc32"
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
	// crc accumulates the IEEE CRC32 of every byte consumed from r
	// before the trailer is read. The trailer bytes themselves are
	// excluded from the running CRC (see readFile).
	crc hash.Hash32
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{
		r:         &countingReader{r: r},
		byteOrder: binary.BigEndian,
		crc:       crc32.NewIEEE(),
	}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// Always called inside an `if err != nil` branch, so it doesn't guard against
// nil errors itself.
func (d *decoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}

// readFile reads a TLV1 file. The current implementation only realizes the
// Trailer pipeline: every byte preceding the trailer is folded into a running
// CRC32, and the trailer is then read and compared. Header and Record
// decoding are out of scope here; once they land, replace the
// "drain everything before the trailer" step with calls to readHeader /
// readRecord — they can use d.crc.Write to keep the running CRC aligned.
func (d *decoder) readFile() (*File, error) {
	// Drain every byte preceding the trailer into the running CRC. We
	// don't know up front how long the body is (records are
	// variable-length and the spec says "read in order until the bytes
	// preceding the trailer are exhausted"), so we buffer the whole
	// stream and treat the last 4 bytes as the trailer.
	buf, err := io.ReadAll(d.r)
	if err != nil {
		return nil, d.wrapErr("File", err)
	}
	if len(buf) < 4 {
		// We've already advanced d.r.n past whatever was available;
		// report Trailer.CRC32 as the failing field.
		return nil, d.wrapErr("Trailer.CRC32", io.ErrUnexpectedEOF)
	}

	body := buf[:len(buf)-4]
	trailerBytes := buf[len(buf)-4:]

	// Fold every pre-trailer byte into the running CRC. Hash.Write never
	// returns an error per the hash.Hash contract, so this is safe.
	_, _ = d.crc.Write(body)

	t, err := d.readTrailer(trailerBytes)
	if err != nil {
		return nil, err
	}
	return &File{Trailer: t}, nil
}

// readTrailer parses the 4-byte big-endian CRC32 trailer and verifies it
// matches the running CRC32 d.crc has accumulated over the preceding bytes.
// On mismatch it returns ErrChecksumMismatch wrapped through wrapErr.
func (d *decoder) readTrailer(b []byte) (Trailer, error) {
	if len(b) < 4 {
		return Trailer{}, d.wrapErr("Trailer.CRC32", io.ErrUnexpectedEOF)
	}
	want := d.crc.Sum32()
	got := d.byteOrder.Uint32(b)
	if got != want {
		return Trailer{}, d.wrapErr("Trailer.CRC32", ErrChecksumMismatch)
	}
	return Trailer{CRC32: got}, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
