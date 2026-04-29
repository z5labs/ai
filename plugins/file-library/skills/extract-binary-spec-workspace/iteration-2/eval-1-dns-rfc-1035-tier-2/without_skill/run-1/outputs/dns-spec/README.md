# DNS Message Binary Format Spec (RFC 1035)

This directory contains an extracted, implementation-oriented specification
of the DNS message wire format from RFC 1035, Section 4 ("Messages"), with
supporting material pulled from Section 2.3 (conventions/limits) and
Sections 3.1-3.4 (labels, TYPE/CLASS values, RDATA formats).

Source: https://www.rfc-editor.org/rfc/rfc1035.txt
RFC 1035 — Domain Implementation and Specification, P. Mockapetris, Nov 1987.

## Scope

In scope (everything needed to encode/decode DNS messages on the wire):

- Section 4.1.1 Header section format (12 bytes, bit-packed flags)
- Section 4.1.2 Question section format
- Section 4.1.3 Resource record format (Answer/Authority/Additional)
- Section 4.1.4 Message compression (label pointers)
- Section 3.1 Domain-name label encoding
- Section 3.2.2/3.2.3 TYPE and QTYPE values
- Section 3.2.4/3.2.5 CLASS and QCLASS values
- Section 3.3 / 3.4 RDATA layouts for the standard RR types
- Section 2.3.2 Data transmission order (network byte order, big-endian)
- Section 2.3.4 Size limits

Out of scope (per user request):

- Master / zone file textual format
- Operational behavior of resolvers and name servers
- IN-ADDR.ARPA, registry, and historical/registry material
- Section 4.2 Transport (UDP/TCP framing) — except UDP 512-byte limit noted
  in size limits

## File index

| File                       | Contents                                            |
|----------------------------|-----------------------------------------------------|
| `00-overview.md`           | Top-level message layout and shared conventions     |
| `01-header.md`             | 12-byte header, all flag bits, opcodes, rcodes      |
| `02-question.md`           | Question section entries (QNAME/QTYPE/QCLASS)       |
| `03-resource-record.md`    | RR wire format common to Answer/Authority/Additional|
| `04-domain-names.md`       | Label encoding + message compression (pointers)     |
| `05-types-and-classes.md`  | TYPE/QTYPE and CLASS/QCLASS code tables             |
| `06-rdata-formats.md`      | Per-type RDATA layouts (A, NS, CNAME, SOA, PTR, MX, |
|                            | TXT, HINFO, MINFO, MB, MG, MR, MD, MF, NULL, WKS)   |
| `07-encoding-rules.md`     | Byte order, size limits, edge cases, gotchas        |

## How to use this spec

Implementers building a Go encoder/decoder should be able to write code
directly from these documents. Each section cites the originating RFC 1035
subsection so you can cross-reference the source if anything is ambiguous.

Where the RFC is silent on edge cases that matter for a robust decoder
(e.g. pointer loops, name length validation across compression), this is
called out explicitly in `07-encoding-rules.md` rather than invented.
