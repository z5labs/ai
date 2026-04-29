# DNS Message Format (RFC 1035 Section 4.1)

A DNS message is a single self-contained unit transmitted over UDP or TCP. All
queries and responses share the same overall structure.

## Top-level layout

```
    +---------------------+
    |        Header       |
    +---------------------+
    |       Question      | the question for the name server
    +---------------------+
    |        Answer       | RRs answering the question
    +---------------------+
    |      Authority      | RRs pointing toward an authority
    +---------------------+
    |      Additional     | RRs holding additional information
    +---------------------+
```

The five sections appear in the order shown above and are concatenated with
no padding or separators.

## Section counts

The Header (always present, fixed 12 bytes) carries four 16-bit unsigned
counts that determine how many entries follow in each variable-length section:

| Header field | Section it counts | Entry type |
|---|---|---|
| `QDCOUNT` | Question | Question entries |
| `ANCOUNT` | Answer | Resource records |
| `NSCOUNT` | Authority | Resource records |
| `ARCOUNT` | Additional | Resource records |

Any of these counts MAY be zero, in which case the corresponding section is
empty (zero bytes).

## Section content rules

- The **Question** section contains `QDCOUNT` question entries (each is a
  QNAME + QTYPE + QCLASS triple - see `question.md`).
- The **Answer**, **Authority**, and **Additional** sections each contain a
  sequence of resource records in the common RR wire format (see
  `resource-record.md`). The number of RRs in each is given by `ANCOUNT`,
  `NSCOUNT`, and `ARCOUNT` respectively.
- The same wire format is used for queries and responses. In a typical query
  only the Question section is populated; in a response, all four counts may
  be non-zero.

## Decoder/encoder implications

- A decoder MUST parse the header first to learn the counts, then iterate the
  four variable-length sections in order.
- Section boundaries are **not** marked by length prefixes; they are inferred
  by counting `QDCOUNT`/`ANCOUNT`/`NSCOUNT`/`ARCOUNT` entries.
- Because RR `NAME` and RDATA fields can contain compression pointers
  referring to earlier offsets in the message (see `domain-names.md`), the
  decoder must keep the entire received message buffer addressable while
  parsing.
