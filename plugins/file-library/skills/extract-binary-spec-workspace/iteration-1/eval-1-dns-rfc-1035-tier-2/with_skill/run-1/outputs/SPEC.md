# DNS Message Binary Specification Reference

## Overview

This reference describes the wire format of Domain Name System (DNS) messages
as defined in RFC 1035 (Domain Names — Implementation and Specification, P.
Mockapetris, November 1987). It covers only the binary message format from
RFC 1035 §4 (Messages) — the 12-byte header, question section, resource
records, and the encoding of domain names (including label compression).

Out of scope (intentionally excluded): the master file (zone file) text
format from §5, registry / operational guidance from §§1–3 and §§6–8, and
all later DNS extensions (EDNS0/RFC 6891, DNSSEC, IDN, etc.). RDATA layouts
are limited to the record types defined in RFC 1035 itself.

## Conventions

- **Byte order**: network byte order (big-endian) for every multi-octet
  numeric field. (RFC 1035 §2.3.2 and §3.2.1.) When this reference says
  "16-bit unsigned" it means a `uint16` read MSB-first off the wire.
- **Bit numbering**: MSB-0. Bit 0 is the most significant bit of an octet,
  bit 7 is the least significant. This matches the diagrams in RFC 1035,
  which number bits left-to-right starting at 0. Bit 0 of octet 2 of the
  Flags field is therefore the QR flag.
- **Size units**: 8-bit octets. RFC 1035 uses "octet" throughout; this
  reference uses "byte" interchangeably.
- **Notation**: Field tables use byte offsets relative to the start of the
  containing structure. Bit-field tables use MSB-0 bit positions within a
  named byte or 16-bit word, and the table header explicitly states which.
- **Message size limits**: 512 bytes maximum over UDP (RFC 1035 §2.3.4 and
  §4.2.1). Messages longer than 512 bytes that are returned over UDP must
  set the TC (truncation) flag. Over TCP the message is preceded by a
  separate 16-bit length field that is *not* part of the DNS message
  itself; see [`structures/tcp-framing.md`](structures/tcp-framing.md).
- **Label / name limits**: a single label is 1–63 octets; a complete
  domain name (the wire-form sequence of labels including the terminating
  zero-length root label) is at most 255 octets. (§2.3.4.)

## Top-level structure

Every DNS message — query or response, UDP or TCP — has the same five-part
layout:

```
+---------------------+
|        Header       |  fixed 12 bytes
+---------------------+
|       Question      |  QDCOUNT entries
+---------------------+
|        Answer       |  ANCOUNT resource records
+---------------------+
|      Authority      |  NSCOUNT resource records
+---------------------+
|      Additional     |  ARCOUNT resource records
+---------------------+
```

The header carries four 16-bit counts (QDCOUNT, ANCOUNT, NSCOUNT, ARCOUNT)
that define how many entries follow in each of the four variable-length
sections. The Question section contains
[`question`](structures/question.md) entries; the Answer, Authority, and
Additional sections all contain
[`resource-record`](structures/resource-record.md) entries with the same
layout but different roles.

Domain names anywhere in a message use the encoding described in
[`structures/domain-name.md`](structures/domain-name.md), which permits
back-references (compression pointers) into earlier parts of the same
message.

## Structures index

- [`structures/header.md`](structures/header.md) — fixed 12-byte message header with ID, Flags word, and the four section counts.
- [`structures/question.md`](structures/question.md) — query record carrying QNAME / QTYPE / QCLASS.
- [`structures/resource-record.md`](structures/resource-record.md) — generic resource record format used in Answer / Authority / Additional sections.
- [`structures/domain-name.md`](structures/domain-name.md) — wire encoding for domain names: length-prefixed labels, root terminator, and compression pointers.
- [`structures/character-string.md`](structures/character-string.md) — `<character-string>`: single-byte length followed by that many octets, used inside RDATA.
- [`structures/tcp-framing.md`](structures/tcp-framing.md) — 2-byte length prefix that brackets a DNS message when sent over TCP.
- [`structures/rdata-a.md`](structures/rdata-a.md) — RDATA for the A record (IPv4 address).
- [`structures/rdata-ns.md`](structures/rdata-ns.md) — RDATA for NS (and the structurally identical CNAME/PTR records).
- [`structures/rdata-soa.md`](structures/rdata-soa.md) — RDATA for the SOA record.
- [`structures/rdata-mx.md`](structures/rdata-mx.md) — RDATA for the MX record.
- [`structures/rdata-txt.md`](structures/rdata-txt.md) — RDATA for the TXT record (one or more `<character-string>`s).
- [`structures/rdata-hinfo.md`](structures/rdata-hinfo.md) — RDATA for HINFO (CPU + OS strings).
- [`structures/rdata-minfo.md`](structures/rdata-minfo.md) — RDATA for MINFO (mailbox responsible / error-mailbox).
- [`structures/rdata-wks.md`](structures/rdata-wks.md) — RDATA for WKS (Well Known Services bitmap).
- [`structures/rdata-null.md`](structures/rdata-null.md) — RDATA for NULL (opaque experimental record).

## Encoding tables index

- [`encoding-tables/opcodes.md`](encoding-tables/opcodes.md) — 4-bit Opcode field in the header (QUERY, IQUERY, STATUS).
- [`encoding-tables/rcodes.md`](encoding-tables/rcodes.md) — 4-bit RCODE field in the header (NoError, FormErr, ServFail, NXDomain, NotImp, Refused).
- [`encoding-tables/types.md`](encoding-tables/types.md) — TYPE values that appear in resource records (A, NS, CNAME, SOA, ...).
- [`encoding-tables/qtypes.md`](encoding-tables/qtypes.md) — QTYPE values (TYPE plus AXFR, MAILB, MAILA, `*`).
- [`encoding-tables/classes.md`](encoding-tables/classes.md) — CLASS values (IN, CS, CH, HS).
- [`encoding-tables/qclasses.md`](encoding-tables/qclasses.md) — QCLASS values (CLASS plus `*`).

## Examples index

- [`examples/minimal.md`](examples/minimal.md) — smallest possible message: a query for `A` of the root zone with no answers.
- [`examples/typical.md`](examples/typical.md) — a typical A-record response showing label compression for the answer's NAME.
- [`examples/complex.md`](examples/complex.md) — a multi-RR response with answer + authority + additional sections, exercising compression and an SOA record.

## Appendix

### Implementation limits (RFC 1035 §2.3.4)

| Limit | Value |
|---|---|
| Single label | 63 octets |
| Domain name (wire form) | 255 octets |
| TTL | 32-bit signed; positive values only (treat MSB=1 as zero) |
| UDP message | 512 octets |

### Related documents

- RFC 1034 — Domain Names: Concepts and Facilities (companion).
- RFC 1035 — this document; the binary wire format.
- RFC 1123 §6.1 — host requirements clarifications.
- RFC 2181 — clarifications to the DNS specification (TTL signedness, etc.).
- RFC 6891 — EDNS(0) extension mechanism (out of scope here).
- IANA "Domain Name System (DNS) Parameters" registry — current TYPE,
  CLASS, OPCODE, RCODE assignments.

### Version history

This reference covers only the original RFC 1035 message format. Later
RFCs add new TYPEs, new RCODEs (>= 16 via OPT), EDNS0 framing, and DNSSEC
records; they are not extracted here.
