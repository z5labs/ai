# Header

The fixed 12-octet header that begins every DNS message. Carries the transaction ID, the bit-packed flags (QR/Opcode/AA/TC/RD/RA/Z/RCODE), and the four section counts. RFC 1035 §4.1.1.

## Byte diagram

```
                                1  1  1  1  1  1
  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      ID                       |
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
| 0 | 2 | uint16 | ID | Transaction identifier chosen by the requester; copied unchanged into the response so the requester can correlate replies. |
| 2 | 2 | uint16 | Flags | Packed flags and codes. See [Bit fields](#bit-fields). |
| 4 | 2 | uint16 | QDCOUNT | Number of entries in the Question section. |
| 6 | 2 | uint16 | ANCOUNT | Number of resource records in the Answer section. |
| 8 | 2 | uint16 | NSCOUNT | Number of resource records in the Authority section. |
| 10 | 2 | uint16 | ARCOUNT | Number of resource records in the Additional section. |

All multi-octet integers are big-endian (network byte order).

## Bit fields

The 16-bit `Flags` word at offset 2. Bits are MSB-0 within the 16-bit value (bit 0 is the most significant bit of the high octet).

| Bit(s) | Name | Description |
|---|---|---|
| 0 | QR | Query/Response. 0 = query, 1 = response. |
| 1-4 | Opcode | Query kind. See [`../encoding-tables/opcodes.md`](../encoding-tables/opcodes.md). |
| 5 | AA | Authoritative Answer. Valid only in responses; set when the responding server is authoritative for the question's domain name. |
| 6 | TC | TrunCation. Set when the message was truncated because it was larger than the transport allowed (most commonly when a UDP response would exceed 512 octets). |
| 7 | RD | Recursion Desired. Set in queries to ask the server to resolve recursively; copied unchanged into the response. |
| 8 | RA | Recursion Available. Set in responses by servers that support recursion. |
| 9 | Z | Reserved. MUST be zero in queries and responses. |
| 10 | (Z) | Reserved (continued). |
| 11 | (Z) | Reserved (continued). |
| 12-15 | RCODE | Response code. See [`../encoding-tables/rcodes.md`](../encoding-tables/rcodes.md). |

> RFC 1035 §4.1.1 shows three reserved bits between RA and RCODE under the single label `Z`. Treat them all as the same reserved field (must be zero, ignored on receive).

### Encoding the Flags word as a uint16

Bit position 0 is the MSB of the 16-bit value. The mapping into a `uint16` (when read as a native big-endian integer):

```
Flags uint16 layout:
  bit 15 (MSB)  QR
  bits 14..11   Opcode (4 bits)
  bit 10        AA
  bit  9        TC
  bit  8        RD
  bit  7        RA
  bits  6..4    Z (3 bits, reserved)
  bits  3..0    RCODE (4 bits)
```

Equivalent masks (uint16):

| Field | Mask | Shift |
|---|---|---|
| QR | `0x8000` | 15 |
| Opcode | `0x7800` | 11 |
| AA | `0x0400` | 10 |
| TC | `0x0200` | 9 |
| RD | `0x0100` | 8 |
| RA | `0x0080` | 7 |
| Z | `0x0070` | 4 |
| RCODE | `0x000F` | 0 |

## Conditional / optional fields

- `AA`, `RA` are meaningful only on responses (`QR=1`); MUST be zero on queries.
- `RCODE` is meaningful only on responses; MUST be zero on queries.
- `RD` is set by the requester; the responder copies it.

## Ambiguities

> **Ambiguity:** RFC 1035 §4.1.1 labels three bits `Z` collectively but does not name them individually. Decoders MUST accept any value (be tolerant — later RFCs reused these bits, e.g., AD/CD in RFC 4035) but RFC 1035 itself says they MUST be zero. Implementers targeting RFC 1035 only should preserve received bits but emit zero.
