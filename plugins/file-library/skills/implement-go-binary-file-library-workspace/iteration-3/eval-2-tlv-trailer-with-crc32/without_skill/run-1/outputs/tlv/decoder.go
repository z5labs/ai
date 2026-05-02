package tlv

import (
	"encoding/binary"
	"errors"
	"hash"
	"hash/crc32"
	"io"
)

// trailerReader wraps an io.Reader and reserves the last 4 bytes of the stream
// as the trailer. Internally it keeps a circular buffer of up to 5 bytes:
// the oldest 4 are the "candidate trailer" and the 5th (if present) is the
// next body byte. A body byte is emitted only after we have confirmed the
// 5th slot is filled, i.e. that at least one more byte exists beyond the
// current candidate trailer. When the underlying reader hits EOF, whatever
// sits in the 4-byte candidate IS the trailer and is NOT emitted as a body
// byte.
//
// Bytes that ARE emitted as body bytes are folded into a running CRC32, so
// when the body is fully drained the hash covers exactly the bytes preceding
// the trailer.
type trailerReader struct {
	r   io.Reader
	buf [5]byte
	// length is the count of valid bytes in buf, starting at buf[0].
	length      int
	upstreamEOF bool
	n           int64 // bytes emitted as body so far
	hash        hash.Hash32
}

func newTrailerReader(r io.Reader) *trailerReader {
	return &trailerReader{r: r, hash: crc32.NewIEEE()}
}

// fill reads from the underlying reader until the buffer holds 5 bytes or
// upstream EOF is reached.
func (t *trailerReader) fill() error {
	for t.length < 5 && !t.upstreamEOF {
		nRead, err := t.r.Read(t.buf[t.length:])
		t.length += nRead
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.upstreamEOF = true
				break
			}
			return err
		}
	}
	return nil
}

// Read returns body bytes (i.e. bytes preceding the trailer). It returns
// io.EOF when only the trailer remains.
func (t *trailerReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if err := t.fill(); err != nil {
		return 0, err
	}
	// length < 4 with upstream EOF means the stream is malformed: the file
	// must end with at least a 4-byte trailer.
	if t.length < 4 {
		return 0, io.ErrUnexpectedEOF
	}
	// length == 4 means upstream is exhausted (otherwise fill would have
	// pulled a 5th byte). Those 4 bytes are the trailer; body is over.
	if t.length == 4 {
		return 0, io.EOF
	}
	// length == 5: emit buf[0] as a body byte, shift the rest down, and try
	// to refill the tail.
	out := t.buf[0]
	t.buf[0] = t.buf[1]
	t.buf[1] = t.buf[2]
	t.buf[2] = t.buf[3]
	t.buf[3] = t.buf[4]
	t.length = 4
	p[0] = out
	t.n++
	_, _ = t.hash.Write(p[:1])
	return 1, nil
}

// peek reports whether at least one more body byte is available, without
// consuming any bytes from the body stream.
func (t *trailerReader) peekHasBody() (bool, error) {
	if err := t.fill(); err != nil {
		return false, err
	}
	if t.length < 4 {
		return false, io.ErrUnexpectedEOF
	}
	return t.length == 5, nil
}

// trailerBytes returns the 4 bytes held in the candidate-trailer slots. Only
// valid after the body has been fully drained (i.e. peekHasBody returned
// false).
func (t *trailerReader) trailerBytes() ([4]byte, error) {
	if err := t.fill(); err != nil {
		return [4]byte{}, err
	}
	if t.length < 4 {
		return [4]byte{}, io.ErrUnexpectedEOF
	}
	if t.length != 4 {
		return [4]byte{}, errors.New("tlv: trailer requested before body fully consumed")
	}
	var out [4]byte
	copy(out[:], t.buf[:4])
	return out, nil
}

// decoder reads TLV1 structures from an underlying io.Reader.
type decoder struct {
	tr        *trailerReader
	byteOrder binary.ByteOrder
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{
		tr:        newTrailerReader(r),
		byteOrder: binary.BigEndian,
	}
}

// wrapErr funnels every error site into the FieldError -> OffsetError -> leaf chain.
// Always called inside an `if err != nil` branch.
func (d *decoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: d.tr.n, Err: err}}
}

// readBody reads exactly len(buf) body bytes (i.e. bytes preceding the trailer).
// Returns io.ErrUnexpectedEOF if the body ends mid-read.
func (d *decoder) readBody(buf []byte) error {
	read := 0
	for read < len(buf) {
		n, err := d.tr.Read(buf[read:])
		read += n
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}
			return err
		}
	}
	return nil
}

// bodyHasMore reports whether at least one more body byte is available.
func (d *decoder) bodyHasMore() (bool, error) {
	return d.tr.peekHasBody()
}

func (d *decoder) readHeader() (Header, error) {
	var buf [8]byte
	if err := d.readBody(buf[:]); err != nil {
		return Header{}, d.wrapErr("Header", err)
	}
	var h Header
	copy(h.Magic[:], buf[0:4])
	if h.Magic != [4]byte{'T', 'L', 'V', '1'} {
		return Header{}, d.wrapErr("Header.Magic", ErrInvalid)
	}
	h.Version = buf[4]
	if h.Version != 1 {
		return Header{}, d.wrapErr("Header.Version", ErrInvalid)
	}
	h.Flags = buf[5]
	h.Reserved = d.byteOrder.Uint16(buf[6:8])
	if h.Reserved != 0 {
		return Header{}, d.wrapErr("Header.Reserved", ErrInvalid)
	}
	return h, nil
}

func (d *decoder) readRecord() (Record, error) {
	var hdr [3]byte
	if err := d.readBody(hdr[:]); err != nil {
		return Record{}, d.wrapErr("Record", err)
	}
	r := Record{
		Type:   hdr[0],
		Length: d.byteOrder.Uint16(hdr[1:3]),
	}
	if r.Length > 0 {
		r.Value = make([]byte, r.Length)
		if err := d.readBody(r.Value); err != nil {
			return Record{}, d.wrapErr("Record.Value", err)
		}
	}
	return r, nil
}

func (d *decoder) readRecords() ([]Record, error) {
	var records []Record
	for {
		more, err := d.bodyHasMore()
		if err != nil {
			return nil, d.wrapErr("Records", err)
		}
		if !more {
			return records, nil
		}
		rec, err := d.readRecord()
		if err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
}

func (d *decoder) readTrailer() (Trailer, error) {
	expected := d.tr.hash.Sum32()
	buf, err := d.tr.trailerBytes()
	if err != nil {
		return Trailer{}, d.wrapErr("Trailer", err)
	}
	// Account for the trailer bytes so error offsets point at the end of the
	// file rather than the start of the trailer.
	d.tr.n += 4
	got := d.byteOrder.Uint32(buf[:])
	if got != expected {
		return Trailer{}, d.wrapErr("Trailer.CRC32", ErrChecksumMismatch)
	}
	return Trailer{CRC32: got}, nil
}

func (d *decoder) readFile() (*File, error) {
	h, err := d.readHeader()
	if err != nil {
		return nil, err
	}
	records, err := d.readRecords()
	if err != nil {
		return nil, err
	}
	t, err := d.readTrailer()
	if err != nil {
		return nil, err
	}
	return &File{Header: h, Records: records, Trailer: t}, nil
}

// Decode reads a TLV1 file from r.
func Decode(r io.Reader) (*File, error) {
	return newDecoder(r).readFile()
}
