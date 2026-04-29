# Header

The fixed 12-byte header that begins every DNS message (query and
response, UDP and TCP). Defined in RFC 1035 §4.1.1.

## Byte diagram

Bit positions are MSB-0; bit 0 is the high bit of each octet. Each row of
the diagram is 16 bits / 2 bytes.

```
                                 1  1  1  1  1  1
   0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
 +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
 |                       ID                      |
 +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
 |QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
 +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
 |                    QDCOUNT                    |
 +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
 |                    ANCOUNT                    |
 +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
 |                    NSCOUNT                    |
 +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
 |                    ARCOUNT                    |
 +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 | ID | Caller-chosen identifier echoed in the response so the client can match query/response pairs. |
| 2 | 2 | uint16 | Flags | Packed flags word — see [Bit fields](#bit-fields). |
| 4 | 2 | uint16 | QDCOUNT | Number of entries in the Question section. |
| 6 | 2 | uint16 | ANCOUNT | Number of resource records in the Answer section. |
| 8 | 2 | uint16 | NSCOUNT | Number of resource records in the Authority section. |
| 10 | 2 | uint16 | ARCOUNT | Number of resource records in the Additional section. |

Total: 12 bytes. All multi-byte integers are big-endian (network byte
order).

## Bit fields

The Flags word at offset 2 packs the following sub-fields. Bit numbering
is MSB-0 across the 16-bit word: bit 0 is the high bit of byte 2 (the
first byte of the Flags word), bit 8 is the high bit of byte 3.

| Bit(s) | Name | Description |
|---|---|---|
| 0 | QR | Query/Response. 0 = query, 1 = response. |
| 1–4 | Opcode | 4-bit operation type — see [`../encoding-tables/opcodes.md`](../encoding-tables/opcodes.md). Set by the originator and copied unchanged into the response. |
| 5 | AA | Authoritative Answer. Valid only in responses; 1 means the responding name server is authoritative for the domain name in the question section. |
| 6 | TC | TrunCation. 1 means this message was truncated because its length exceeded the transport's permitted size (typically 512 bytes for UDP). |
| 7 | RD | Recursion Desired. Set by the client; copied into the response. 1 means "please pursue the query recursively". |
| 8 | RA | Recursion Available. Set in responses only. 1 means the responding server supports recursive queries. |
| 9–11 | Z | Reserved. Must be zero in queries and responses (per RFC 1035; later RFCs reuse two of these bits for AD/CD in DNSSEC, out of scope here). |
| 12–15 | RCODE | 4-bit response code — see [`../encoding-tables/rcodes.md`](../encoding-tables/rcodes.md). Meaningful only in responses. |

Equivalent layout viewed as two octets (offset 2 = first flags byte,
offset 3 = second flags byte):

| Byte offset | Bit(s) within byte (MSB-0) | Name |
|---|---|---|
| 2 | 0 | QR |
| 2 | 1–4 | Opcode |
| 2 | 5 | AA |
| 2 | 6 | TC |
| 2 | 7 | RD |
| 3 | 0 | RA |
| 3 | 1–3 | Z |
| 3 | 4–7 | RCODE |

### Encoding examples (Flags word as a `uint16`)

- Standard recursive query (QR=0, Opcode=0, RD=1, all others 0): `0x0100`.
- Standard non-recursive query: `0x0000`.
- Standard query response, no error, recursion available, recursion was
  requested (QR=1, RD=1, RA=1, RCODE=0): `0x8180`.
- Authoritative answer, NXDOMAIN, no recursion (QR=1, AA=1, RCODE=3):
  `0x8403`.

## Variable-length fields

None. The header is always exactly 12 bytes regardless of contents.

## Conditional / optional fields

None. All six 16-bit fields are always present.

## Checksums and integrity

None. DNS relies on the underlying transport (UDP/IP checksums or TCP
checksums) for integrity.

## Padding and alignment

None. The header is naturally 16-bit aligned at all field boundaries when
read sequentially.

## Nested structures

The header is followed by:

- QDCOUNT × [`question`](question.md) entries,
- ANCOUNT × [`resource-record`](resource-record.md) (Answer section),
- NSCOUNT × [`resource-record`](resource-record.md) (Authority section),
- ARCOUNT × [`resource-record`](resource-record.md) (Additional section).

In total the message is 12 bytes plus the variable-length payload of
those four sections.

## Versioning notes

There is no version field. RFC 1035 defines exactly one header layout.
Later RFCs reinterpret the Z bits (AD/CD for DNSSEC), define new RCODEs
beyond 4 bits via the EDNS0 OPT pseudo-RR, and add OPCODE 5 (UPDATE), but
none of those changes alter the on-the-wire header layout described here.

## Ambiguities

> **Ambiguity:** RFC 1035 specifies that all three Z bits "must be zero
> in all queries and responses". Real-world implementations frequently
> tolerate non-zero Z bits because of later RFCs reusing them (RFC 4035
> AD/CD bits). A strict 1035-only decoder may either reject non-zero Z or
> accept and ignore; this reference does not mandate a choice but a
> permissive decoder is generally recommended.

> **Ambiguity:** RFC 1035 originally defined RCODE as a 4-bit field with
> values 0–5 assigned. The IANA registry now assigns higher values that
> require EDNS0 to transmit. A pure RFC 1035 decoder treats values 6–15
> as "unknown error" rather than rejecting the message.
