package tlvx

import (
	"encoding/binary"
	"io"
)

// countingWriter wraps an io.Writer and tracks the number of bytes written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

// encoder writes TLVX structures to an underlying io.Writer.
type encoder struct {
	w         *countingWriter
	byteOrder binary.ByteOrder
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{w: &countingWriter{w: w}, byteOrder: binary.BigEndian}
}

// wrapErr funnels every error site into the FieldError → OffsetError → leaf chain.
// Always called inside an `if err != nil` branch.
func (e *encoder) wrapErr(field string, err error) error {
	return &FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}
}

// writeFile writes a complete TLVX file. Currently only the Header is wired up.
func (e *encoder) writeFile(f *File) error {
	if err := e.writeHeader(&f.Header); err != nil {
		return err
	}
	return nil
}

// writeHeader writes the fixed 16-byte TLVX file header. Validation mirrors
// the decoder: reserved bit 7 of Flags must be clear, Reserved1 must be 0,
// ChecksumAlg must be a defined value, and Version must be 1. Failures funnel
// through wrapErr.
func (e *encoder) writeHeader(h *Header) error {
	if h.Version != Version1 {
		return e.wrapErr("Header.Version", &UnknownVersionError{Version: h.Version})
	}

	if h.Flags&flagsReservedMask != 0 {
		return e.wrapErr("Header.Flags", &ReservedFlagBitError{Field: "Header.Flags", Bit: 7})
	}

	if h.Reserved1 != 0 {
		return e.wrapErr("Header.Reserved1", &ReservedFieldNonZeroError{Field: "Header.Reserved1", Got: h.Reserved1})
	}

	switch h.ChecksumAlg {
	case ChecksumAlgCRC32IEEE,
		ChecksumAlgCRC64ECMA,
		ChecksumAlgSHA256T32,
		ChecksumAlgXXH64,
		ChecksumAlgBLAKE3T32:
		// ok
	default:
		return e.wrapErr("Header.ChecksumAlg", &UnknownChecksumAlgError{Alg: h.ChecksumAlg})
	}

	// Default Magic to MagicTLVX if zero, then sanity-check.
	magic := h.Magic
	if magic == ([4]byte{}) {
		magic = MagicTLVX
	}
	if magic != MagicTLVX {
		return e.wrapErr("Header.Magic", &UnexpectedMagicError{Got: magic})
	}

	if err := binary.Write(e.w, e.byteOrder, magic); err != nil {
		return e.wrapErr("Header.Magic", err)
	}

	rest := struct {
		Version       uint8
		Flags         Flags
		ChecksumAlg   ChecksumAlg
		Reserved1     uint8
		IndexCount    uint16
		ExtCount      uint16
		TrailerOffset uint32
	}{
		Version:       h.Version,
		Flags:         h.Flags,
		ChecksumAlg:   h.ChecksumAlg,
		Reserved1:     h.Reserved1,
		IndexCount:    h.IndexCount,
		ExtCount:      h.ExtCount,
		TrailerOffset: h.TrailerOffset,
	}
	if err := binary.Write(e.w, e.byteOrder, rest); err != nil {
		return e.wrapErr("Header", err)
	}

	return nil
}

// Encode writes f to w as a TLVX file.
func Encode(w io.Writer, f *File) error {
	return newEncoder(w).writeFile(f)
}
