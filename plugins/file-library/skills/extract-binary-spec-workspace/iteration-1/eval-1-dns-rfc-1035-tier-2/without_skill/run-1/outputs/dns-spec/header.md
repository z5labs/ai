# Header Section Format (RFC 1035 Section 4.1.1)

The header is **always 12 octets (96 bits) long**. It is present in every DNS
message (query or response).

## Wire layout

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

Six 16-bit fields = 12 octets total. All multi-octet fields are big-endian.

## Field summary

| Offset (bytes) | Length | Field    | Notes                                  |
|---------------:|-------:|----------|----------------------------------------|
| 0              | 2      | ID       | 16-bit identifier chosen by requester. |
| 2              | 2      | Flags    | Bit-packed; see breakdown below.       |
| 4              | 2      | QDCOUNT  | Unsigned count of question entries.    |
| 6              | 2      | ANCOUNT  | Unsigned count of answer RRs.          |
| 8              | 2      | NSCOUNT  | Unsigned count of authority RRs.       |
| 10             | 2      | ARCOUNT  | Unsigned count of additional RRs.      |

## Field details

### ID (16 bits)

A 16-bit identifier assigned by the program that generates any kind of query.
This identifier is copied to the corresponding reply and can be used by the
requester to match up replies to outstanding queries.

### Flags word (16 bits, offset 2)

The second 16-bit word is bit-packed. Bit 0 is the MSB of byte 2; bit 15 is
the LSB of byte 3.

| Bit(s) | Width | Name   | Description                                       |
|-------:|------:|--------|---------------------------------------------------|
| 0      | 1     | QR     | 0 = query, 1 = response.                          |
| 1-4    | 4     | Opcode | Kind of query. Set by originator, copied to reply.|
| 5      | 1     | AA     | Authoritative Answer (response only).             |
| 6      | 1     | TC     | TrunCation - message was truncated.               |
| 7      | 1     | RD     | Recursion Desired (set by query, copied to reply).|
| 8      | 1     | RA     | Recursion Available (set in response).            |
| 9-11   | 3     | Z      | Reserved for future use. MUST be zero on send.    |
| 12-15  | 4     | RCODE  | Response code.                                    |

### Flag bit semantics

- **QR** - 1-bit field. `0` indicates the message is a query, `1` indicates a
  response.
- **Opcode** - 4-bit field. Specifies the kind of query in this message. This
  value is set by the originator and copied into the response. Defined values:
  - `0` QUERY - a standard query.
  - `1` IQUERY - an inverse query.
  - `2` STATUS - a server status request.
  - `3-15` reserved for future use.
- **AA** (Authoritative Answer) - valid in responses. Indicates that the
  responding name server is an authority for the domain name in question.
  Note that the contents of the answer section may have multiple owner names
  because of aliases; the AA bit corresponds to the name that matches the
  query name, or the first owner name in the answer section.
- **TC** (TrunCation) - specifies that this message was truncated due to
  length greater than that permitted on the transmission channel.
- **RD** (Recursion Desired) - this bit may be set in a query and is copied
  into the response. If RD is set, it directs the name server to pursue the
  query recursively. Recursive query support is optional.
- **RA** (Recursion Available) - set or cleared in a response, denoting
  whether recursive query support is available in the name server.
- **Z** - reserved for future use. Must be zero in all queries and responses.
  3 bits wide.
- **RCODE** (Response Code) - 4-bit field set as part of responses. Defined
  values:
  - `0` No error condition.
  - `1` Format error - the name server was unable to interpret the query.
  - `2` Server failure - the name server was unable to process this query
    due to a problem with the name server.
  - `3` Name Error - meaningful only for responses from an authoritative name
    server, signifies that the domain name referenced in the query does not
    exist.
  - `4` Not Implemented - the name server does not support the requested kind
    of query.
  - `5` Refused - the name server refuses to perform the specified operation
    for policy reasons.
  - `6-15` reserved for future use.

### QDCOUNT (16 bits)

Unsigned 16-bit integer specifying the number of entries in the question
section.

### ANCOUNT (16 bits)

Unsigned 16-bit integer specifying the number of resource records in the
answer section.

### NSCOUNT (16 bits)

Unsigned 16-bit integer specifying the number of name server resource records
in the authority records section.

### ARCOUNT (16 bits)

Unsigned 16-bit integer specifying the number of resource records in the
additional records section.

## Bit-pack reference (Go-style)

Given the flags word as a `uint16` `f`:

```
QR     = (f >> 15) & 0x1     // bit 0 in RFC numbering = MSB
Opcode = (f >> 11) & 0xF     // bits 1-4
AA     = (f >> 10) & 0x1     // bit 5
TC     = (f >>  9) & 0x1     // bit 6
RD     = (f >>  8) & 0x1     // bit 7
RA     = (f >>  7) & 0x1     // bit 8
Z      = (f >>  4) & 0x7     // bits 9-11
RCODE  =  f        & 0xF     // bits 12-15
```

Equivalently, by byte:

- Byte 2 (high byte of flags word):
  - bit 7 (0x80): QR
  - bits 6-3 (0x78): Opcode
  - bit 2 (0x04): AA
  - bit 1 (0x02): TC
  - bit 0 (0x01): RD
- Byte 3 (low byte of flags word):
  - bit 7 (0x80): RA
  - bits 6-4 (0x70): Z
  - bits 3-0 (0x0F): RCODE
