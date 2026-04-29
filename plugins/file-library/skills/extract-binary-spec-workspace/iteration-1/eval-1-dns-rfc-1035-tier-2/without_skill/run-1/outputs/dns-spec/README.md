# DNS Message Binary Format Specification

Extracted from **RFC 1035 - Domain Names - Implementation and Specification**
(P. Mockapetris, November 1987), Section 4 "MESSAGES" with supporting material
from Sections 2.3.4 (Size limits), 3.1 (Name space definitions), 3.2.2-3.2.5
(TYPE/QTYPE/CLASS/QCLASS values), 3.3 (Standard RRs), and 3.4 (Internet specific RRs).

Source: https://www.rfc-editor.org/rfc/rfc1035.txt

## Scope of this spec

This spec contains only what is needed to implement a Go encoder/decoder for DNS
messages on the wire. Specifically:

- The fixed 12-byte message header, including the bit-packed second 16-bit word
  (QR, Opcode, AA, TC, RD, RA, Z, RCODE).
- The question section entry format (QNAME, QTYPE, QCLASS).
- The resource record format (NAME, TYPE, CLASS, TTL, RDLENGTH, RDATA).
- Domain name encoding using length-prefixed labels and pointer (label)
  compression.
- The standard RR types' RDATA layouts (A, NS, CNAME, SOA, PTR, MX, TXT, HINFO,
  MINFO, MB, MD, MF, MG, MR, NULL, WKS).
- TYPE/QTYPE, CLASS/QCLASS, Opcode, and RCODE numeric values.
- Transport-level message size constraints (UDP 512-byte limit, TCP 2-byte
  length prefix).

Explicitly **out of scope** (per user request and skipped from RFC):
- Master file format (zone files), Section 5.
- Resolver and name-server algorithms / caching, Sections 6 and 7.
- Mail support details beyond RR layouts.
- IANA registry / operational guidance from earlier sections.

## File layout

| File | Contents |
|---|---|
| `byte-order.md` | Bit/byte order conventions used by all diagrams. |
| `message-format.md` | Top-level message structure and the five sections. |
| `header.md` | The 12-byte header, including bit-packed flags. |
| `question.md` | Question section entry format. |
| `resource-record.md` | Common RR wire format (NAME/TYPE/CLASS/TTL/RDLENGTH/RDATA). |
| `domain-names.md` | Label encoding, root, max sizes, and pointer compression. |
| `rdata-formats.md` | Per-TYPE RDATA wire layouts for all standard RRs. |
| `type-class-values.md` | Numeric values for TYPE, QTYPE, CLASS, QCLASS, Opcode, RCODE. |
| `transport.md` | UDP/TCP framing and size limits. |
| `constants.md` | All numeric constants in one place for codegen reference. |
