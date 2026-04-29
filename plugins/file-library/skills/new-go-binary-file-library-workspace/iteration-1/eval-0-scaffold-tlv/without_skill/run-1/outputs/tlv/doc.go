// Package tlv provides primitives for encoding and decoding simple
// Type-Length-Value (TLV) records.
//
// A TLV record consists of:
//
//   - Type:   identifies the record kind.
//   - Length: number of bytes in the value payload.
//   - Value:  raw payload bytes whose interpretation depends on Type.
//
// The exact wire format (integer widths, endianness, framing) is intentionally
// left for the caller to define by filling in the constants and helpers in
// this package.
package tlv
