package gzip

import "errors"

// Sentinel errors returned by the decoder.
var (
	// ErrInvalidMagic indicates the stream did not begin with the gzip
	// magic bytes 0x1f 0x8b.
	ErrInvalidMagic = errors.New("gzip: invalid magic number")

	// ErrUnsupportedCompressionMethod indicates the CM byte specified a
	// compression method other than deflate (8).
	ErrUnsupportedCompressionMethod = errors.New("gzip: unsupported compression method")

	// ErrReservedFlagsSet indicates one of the reserved bits in FLG was set.
	ErrReservedFlagsSet = errors.New("gzip: reserved flag bits set")

	// ErrUnexpectedEOF indicates the stream ended in the middle of a member.
	ErrUnexpectedEOF = errors.New("gzip: unexpected end of stream")

	// ErrHeaderChecksum indicates the FHCRC header CRC-16 did not match.
	ErrHeaderChecksum = errors.New("gzip: header CRC-16 mismatch")
)
