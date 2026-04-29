# DNS Header

Source: RFC 1035 §4.1.1.

The header is a fixed-size, 12-octet (96-bit) structure. It is always
present and always appears at offset 0 of the message.

## Wire layout

```
                                1  1  1  1  1  1
  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      ID                       |   octets 0..1
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |   octets 2..3
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    QDCOUNT                    |   octets 4..5
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ANCOUNT                    |   octets 6..7
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    NSCOUNT                    |   octets 8..9
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ARCOUNT                    |   octets 10..11
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

All fields are 16 bits except the second word (octets 2..3), which is a
bit-packed flags / opcode / rcode word.

## Field-by-field

### ID (octets 0..1, 16 bits)
A 16-bit identifier chosen by the originator of a query. The responder
copies it verbatim into the response so the requester can match replies
to outstanding queries.

### Flags word (octets 2..3, 16 bits, bit-packed)

Bits are numbered MSB = 0 within the 16-bit word. With the word read as
a big-endian uint16, the layout is:

| Bit(s) | Width | Field   | Notes                                        |
|--------|-------|---------|----------------------------------------------|
| 0      | 1     | QR      | 0 = query, 1 = response                      |
| 1..4   | 4     | OPCODE  | Kind of query, see table below               |
| 5      | 1     | AA      | Authoritative Answer (response only)         |
| 6      | 1     | TC      | TrunCation                                   |
| 7      | 1     | RD      | Recursion Desired (set by query, copied)     |
| 8      | 1     | RA      | Recursion Available (response only)          |
| 9..11  | 3     | Z       | Reserved, MUST be zero in queries/responses  |
| 12..15 | 4     | RCODE   | Response code (response only), see table     |

Equivalent bitmask view of the flags word as a `uint16`:

| Field   | Mask     | Shift |
|---------|----------|-------|
| QR      | 0x8000   | 15    |
| OPCODE  | 0x7800   | 11    |
| AA      | 0x0400   | 10    |
| TC      | 0x0200   | 9     |
| RD      | 0x0100   | 8     |
| RA      | 0x0080   | 7     |
| Z       | 0x0070   | 4     |
| RCODE   | 0x000F   | 0     |

Encoding example: `flags = (qr<<15) | (opcode<<11) | (aa<<10) | (tc<<9) |
(rd<<8) | (ra<<7) | (z<<4) | rcode`.

Decoding example: `qr = (flags >> 15) & 0x1`, `opcode = (flags >> 11) & 0xF`,
`aa = (flags >> 10) & 0x1`, etc.

### OPCODE values (4 bits)

| Value | Mnemonic | Meaning                                       |
|-------|----------|-----------------------------------------------|
| 0     | QUERY    | Standard query                                |
| 1     | IQUERY   | Inverse query                                 |
| 2     | STATUS   | Server status request                         |
| 3..15 |          | Reserved for future use                       |

### RCODE values (4 bits)

| Value | Meaning                                                          |
|-------|------------------------------------------------------------------|
| 0     | No error condition                                               |
| 1     | Format error — name server could not interpret the query         |
| 2     | Server failure — internal problem in the name server             |
| 3     | Name Error (NXDOMAIN) — domain name in query does not exist;     |
|       | meaningful only from an authoritative name server                |
| 4     | Not Implemented — server does not support this kind of query     |
| 5     | Refused — server refused for policy reasons                      |
| 6..15 | Reserved for future use                                          |

### QDCOUNT (octets 4..5, unsigned 16-bit)
Number of entries in the Question section.

### ANCOUNT (octets 6..7, unsigned 16-bit)
Number of resource records in the Answer section.

### NSCOUNT (octets 8..9, unsigned 16-bit)
Number of resource records in the Authority section.

### ARCOUNT (octets 10..11, unsigned 16-bit)
Number of resource records in the Additional section.

## Encoder rules

- Always emit exactly 12 octets.
- Z MUST be written as zero.
- Encoder MAY refuse to encode flag combinations that violate semantics
  (e.g. AA set on a query) but the wire format itself does not police it.

## Decoder rules

- Read exactly 12 octets at offset 0.
- Treat the flags word as a single big-endian uint16 and unpack with the
  masks above.
- Z bits SHOULD be ignored on decode (do not reject just because Z != 0,
  since later RFCs reuse those bits, e.g. AD/CD in RFC 4035; out of scope
  here but worth knowing for future-proofing).
- The four COUNT fields tell the decoder how many entries to read in each
  of the following four sections.
