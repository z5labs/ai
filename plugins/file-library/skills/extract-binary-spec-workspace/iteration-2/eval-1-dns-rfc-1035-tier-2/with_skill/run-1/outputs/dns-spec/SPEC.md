# DNS (RFC 1035) Binary Specification Reference

## Overview

The Domain Name System (DNS) is a distributed name-resolution protocol. This reference covers the on-the-wire **message** format defined by RFC 1035 (November 1987) — the format used by every DNS query and response over UDP and TCP. The reference is scoped to the binary protocol: section 4 (Messages), section 3.2 (RR top-level format and the type/class registries), and section 3.3–3.4.1 (RDATA formats for the standard RR types). The master-file (zone-file) text format from RFC 1035 §5 and the operational sections (§6–§8) are intentionally out of scope; this reference is meant for implementing a Go DNS message encoder/decoder.

Authoritative source: RFC 1035 — `https://www.rfc-editor.org/rfc/rfc1035.txt`.

## Conventions

- **Byte order:** big-endian (network byte order). RFC 1035 §2.3.2: when a multi-octet field represents a numeric quantity, the most significant octet is transmitted first. This applies globally; per-structure files do not restate it.
- **Bit numbering:** MSB-0. The bit labelled 0 in any RFC 1035 wire diagram is the most significant bit of the field. Inside a byte, bit 0 has mask `0x80`.
- **Size units:** octets (8-bit bytes). The RFC uses "octet" interchangeably with "byte"; this reference uses "byte" throughout except in direct quotations.
- **Notation:** wire diagrams reproduce the RFC's ASCII art (`+--+--+...+` rows for bytes, `/.../` for variable-length payloads). Hex values are shown as `0xNN`. Field tables use `Offset (bytes)` and `Bit(s)` to keep byte- and bit-addressed quantities visually distinct.
- **Compression:** domain-name compression (RFC 1035 §4.1.4) is described in [`structures/domain-name.md`](structures/domain-name.md). Compression pointers reference offsets from the start of the message (offset 0 = first byte of the header `ID`).

## Top-level structure

A DNS communication is a single `Message`. Every message is exactly:

```
+---------+-------------+----------+-------------+-------------+
| Header  |  Question   |  Answer  |  Authority  | Additional  |
| 12 B    |  QDCOUNT×Q  | ANCOUNT×R| NSCOUNT×R   | ARCOUNT×R   |
+---------+-------------+----------+-------------+-------------+
```

The 12-byte `Header` carries an ID, bit-packed flags (QR/Opcode/AA/TC/RD/RA/Z/RCODE), and the four counts. Each `Question` (`Q`) is `(QName, QType, QClass)`. Each entry in Answer/Authority/Additional is a `ResourceRecord` (`R`): `(Name, Type, Class, TTL, RDLength, RData)`. `RData` is type-specific; this reference defines one `rdata-*.md` file per type.

Over UDP, the message is sent bare (capped at 512 octets per RFC 1035 §4.2.1). Over TCP, the message is preceded by a 2-octet big-endian length field that excludes the 2 length octets (RFC 1035 §4.2.2).

## Structures index

- [`structures/message.md`](structures/message.md) — the top-level container (header + four record-bearing sections) and transport framing.
- [`structures/header.md`](structures/header.md) — fixed 12-byte header with ID, bit-packed Flags (QR/Opcode/AA/TC/RD/RA/Z/RCODE), and the four section counts.
- [`structures/question.md`](structures/question.md) — entry in the Question section: QName + QType + QClass.
- [`structures/resource-record.md`](structures/resource-record.md) — common Answer/Authority/Additional entry: Name + Type + Class + TTL + RDLength + RData.
- [`structures/domain-name.md`](structures/domain-name.md) — recursive label encoding and §4.1.4 compression pointers (with decode pseudocode).
- [`structures/character-string.md`](structures/character-string.md) — single-octet-prefixed binary string used inside HINFO and TXT.
- [`structures/rdata-a.md`](structures/rdata-a.md) — A: 4-byte IPv4 address.
- [`structures/rdata-ns.md`](structures/rdata-ns.md) — NS: single domain name (NSDNAME).
- [`structures/rdata-cname.md`](structures/rdata-cname.md) — CNAME: single domain name.
- [`structures/rdata-ptr.md`](structures/rdata-ptr.md) — PTR: single domain name (PTRDNAME).
- [`structures/rdata-mx.md`](structures/rdata-mx.md) — MX: 16-bit Preference + Exchange domain name.
- [`structures/rdata-soa.md`](structures/rdata-soa.md) — SOA: MNAME + RNAME + 5×32-bit timers.
- [`structures/rdata-txt.md`](structures/rdata-txt.md) — TXT: one or more CharacterStrings.
- [`structures/rdata-hinfo.md`](structures/rdata-hinfo.md) — HINFO: CPU + OS CharacterStrings.
- [`structures/rdata-minfo.md`](structures/rdata-minfo.md) — MINFO (EXPERIMENTAL): RMAILBX + EMAILBX domain names.
- [`structures/rdata-mb.md`](structures/rdata-mb.md) — MB (EXPERIMENTAL): single domain name (MADNAME).
- [`structures/rdata-mg.md`](structures/rdata-mg.md) — MG (EXPERIMENTAL): single domain name (MGMNAME).
- [`structures/rdata-mr.md`](structures/rdata-mr.md) — MR (EXPERIMENTAL): single domain name (NEWNAME).
- [`structures/rdata-md.md`](structures/rdata-md.md) — MD (Obsolete): single domain name (MADNAME); migrate to MX.
- [`structures/rdata-mf.md`](structures/rdata-mf.md) — MF (Obsolete): single domain name (MADNAME); migrate to MX.
- [`structures/rdata-null.md`](structures/rdata-null.md) — NULL (EXPERIMENTAL): opaque bytes up to 65535.
- [`structures/rdata-wks.md`](structures/rdata-wks.md) — WKS: 4-byte IPv4 + 1-byte protocol + bitmap of well-known ports.

## Encoding tables index

- [`encoding-tables/opcodes.md`](encoding-tables/opcodes.md) — 4-bit Opcode in the header Flags word.
- [`encoding-tables/rcodes.md`](encoding-tables/rcodes.md) — 4-bit RCode in the header Flags word.
- [`encoding-tables/types.md`](encoding-tables/types.md) — 16-bit RR Type values defined by RFC 1035.
- [`encoding-tables/qtypes.md`](encoding-tables/qtypes.md) — 16-bit QType values (Type values plus AXFR/MAILB/MAILA/`*`).
- [`encoding-tables/classes.md`](encoding-tables/classes.md) — 16-bit RR Class values (IN/CS/CH/HS).
- [`encoding-tables/qclasses.md`](encoding-tables/qclasses.md) — 16-bit QClass values (Class values plus `*`).

## Examples index

- [`examples/minimal.md`](examples/minimal.md) — 29-byte standard query for `example.com. A IN`, no compression.
- [`examples/typical.md`](examples/typical.md) — 45-byte authoritative response with one A answer, owner-name compression pointer.
- [`examples/complex.md`](examples/complex.md) — 106-byte authoritative MX response with two answers, an authority NS, and an additional A; demonstrates label-then-pointer chained compression and a pointer into the middle of an earlier RDATA.

## Appendix

### Implementation limits

| Limit | Value | Source |
|---|---|---|
| Label length | 63 octets | RFC 1035 §2.3.4, §3.1 |
| Domain name length on wire (label bytes + length bytes, including root terminator) | 255 octets | RFC 1035 §2.3.4 |
| TTL value range | 0 to 2³¹−1 (positive 32-bit signed) | RFC 1035 §2.3.4, §3.2.1 |
| UDP message size | 512 octets | RFC 1035 §4.2.1 |
| TCP message size | bounded by the 2-octet length prefix (≤ 65535) | RFC 1035 §4.2.2 |
| RDLength | 0 to 65535 | RFC 1035 §3.2.1 |
| CharacterString payload | 0 to 255 octets | RFC 1035 §3.3 |
| Compression pointer offset range | 0 to 16383 (14-bit) | RFC 1035 §4.1.4 |

### Related RFCs (informational, out of scope here)

- RFC 1034 — Domain Names – Concepts and Facilities (companion to RFC 1035).
- RFC 974 — Mail routing and the domain system (defines MX preference semantics; obsoletes MD/MF).
- RFC 3425 — Obsoleting IQUERY (Opcode 1).
- RFC 6891 — EDNS0 (extends UDP message size beyond 512 octets via OPT pseudo-RR).
- RFC 3596 — AAAA RR for IPv6 (Type 28).

### IANA registry pointers

- DNS Parameters: <https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml>
  - DNS TYPEs / QTYPEs and DNS CLASSes / QCLASSes.
  - DNS RCODEs.
  - DNS OpCodes.

### Version history

This reference describes RFC 1035 (November 1987) only. Later revisions to the DNS wire format (EDNS0, DNSSEC, etc.) are explicit out-of-scope per the user task.
