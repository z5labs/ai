package dns

import "errors"

// ErrUnexpectedEOF is returned when the decoder runs out of input before a
// complete DNS message has been read.
var ErrUnexpectedEOF = errors.New("dns: unexpected end of input")

// ErrInvalidMessage is returned when the input bytes do not represent a
// well-formed DNS wire-format message.
var ErrInvalidMessage = errors.New("dns: invalid wire-format message")
