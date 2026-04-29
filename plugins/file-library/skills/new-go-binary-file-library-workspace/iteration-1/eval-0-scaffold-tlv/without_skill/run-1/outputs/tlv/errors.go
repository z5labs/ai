package tlv

import "errors"

// Sentinel errors returned by the decoder. Callers should compare with
// errors.Is so that decoders can wrap these with additional context.
var (
	// ErrShortRead is returned when the underlying reader returns fewer
	// bytes than the header or declared length requires.
	ErrShortRead = errors.New("tlv: short read")

	// ErrInvalidLength is returned when a record advertises a length that
	// is invalid for the configured wire format (e.g. negative, or larger
	// than a configured maximum).
	ErrInvalidLength = errors.New("tlv: invalid length")

	// ErrUnknownType is returned by higher-level helpers when a Type is
	// encountered that the caller has not registered. The base Decoder
	// does not return this error on its own.
	ErrUnknownType = errors.New("tlv: unknown type")
)
