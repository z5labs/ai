// Package dns provides parsing and serialization of the DNS binary
// wire-format messages as defined by RFC 1035 (and related RFCs).
//
// This package follows a Tokenizer -> Parser -> AST -> Printer pipeline:
//
//   - The decoder side reads a byte stream, tokenizes wire-format fields,
//     parses them into AST nodes (Message, Header, Question, ResourceRecord),
//     and exposes a typed Message value.
//   - The encoder side walks AST nodes and prints them back to a byte stream
//     suitable for transmission over UDP/TCP.
//
// The boilerplate types and functions in this package are placeholders;
// concrete RFC 1035 types and behavior are filled in by follow-up work.
package dns
