# Question Section Format (RFC 1035 Section 4.1.2)

The question section is used to carry the "question" in most queries, i.e.,
the parameters that define what is being asked. The section contains
`QDCOUNT` (usually 1) entries, each in the following format:

## Wire layout

```
                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                                               |
    /                     QNAME                     /
    /                                               /
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     QTYPE                     |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
    |                     QCLASS                    |
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

The slashes denote a variable-length field.

## Fields

### QNAME (variable length)

A domain name represented as a sequence of labels, where each label consists
of a length octet followed by that number of octets. The domain name
terminates with the zero-length label representing the root (a single `0x00`
byte). Note that this field may be an odd number of octets; **no padding is
used**.

See `domain-names.md` for the full label encoding and pointer compression
rules. QNAME may legally use compression pointers, although in practice the
QNAME in a fresh query is usually fully expanded.

### QTYPE (16 bits)

A two-octet, big-endian unsigned integer specifying the type of the query.
The values for this field include all codes valid for a TYPE field, together
with some more general codes which can match more than one type of RR.

See `type-class-values.md` for the value table. Common QTYPE values include
the standard TYPE codes (A=1, NS=2, CNAME=5, SOA=6, PTR=12, MX=15, TXT=16,
etc.) plus the QTYPE-only extensions:

| Value | Mnemonic | Meaning                              |
|------:|----------|--------------------------------------|
| 252   | AXFR     | A request for a transfer of an entire zone. |
| 253   | MAILB    | A request for mailbox-related records (MB, MG, MR). |
| 254   | MAILA    | A request for mail agent RRs (obsolete - see MX). |
| 255   | *        | A request for all records.           |

### QCLASS (16 bits)

A two-octet, big-endian unsigned integer specifying the class of the query.
For example, QCLASS field is `IN` (`1`) for the Internet.

QCLASS includes all CLASS codes plus the wildcard:

| Value | Mnemonic | Meaning            |
|------:|----------|--------------------|
| 1     | IN       | the Internet       |
| 2     | CS       | CSNET (obsolete)   |
| 3     | CH       | CHAOS              |
| 4     | HS       | Hesiod             |
| 255   | *        | any class          |

## Encoding/decoding notes

- Each entry is `len(QNAME) + 4` octets on the wire.
- The Question section as a whole is `sum(entry sizes)` octets, with no
  separators between entries.
- A decoder must walk the QNAME labels first to find the end of QNAME before
  reading the fixed 4 octets for QTYPE/QCLASS.
