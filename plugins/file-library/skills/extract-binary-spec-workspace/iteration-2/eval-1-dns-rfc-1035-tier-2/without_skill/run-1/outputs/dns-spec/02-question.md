# Question Section

Source: RFC 1035 §4.1.2.

The Question section carries the parameters that define what is being
asked. It contains exactly QDCOUNT entries (typically 1 in a query). Each
entry has the following format.

## Wire layout (per entry)

```
                                1  1  1  1  1  1
  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
/                     QNAME                     /   variable length
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     QTYPE                     |   2 octets
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     QCLASS                    |   2 octets
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Fields

### QNAME (variable length)

A domain name represented as a sequence of length-prefixed labels,
terminated by a zero-length label (the root). See `04-domain-names.md`
for the complete label/compression encoding.

Notes:
- May be an odd number of octets; no padding.
- May use compression pointers per §4.1.4 (see `04-domain-names.md`).

### QTYPE (16 bits, big-endian)

A two-octet code which specifies the type of the query. The values for
this field include all codes valid for a TYPE field, plus QTYPE-only
codes that match more than one TYPE. See `05-types-and-classes.md`.

### QCLASS (16 bits, big-endian)

A two-octet code that specifies the class of the query. For Internet
queries this is `IN` (1). See `05-types-and-classes.md`.

## Encoding

1. Encode QNAME (sequence of labels ending in a 0 byte; compression
   optional but typically not used in question entries).
2. Append QTYPE as big-endian uint16.
3. Append QCLASS as big-endian uint16.

## Decoding

1. Decode QNAME starting at the current offset (handle pointers).
2. Read QTYPE as big-endian uint16.
3. Read QCLASS as big-endian uint16.
4. Repeat QDCOUNT times.

There is no length-prefix on the question entry itself; its length is
determined entirely by the length of QNAME plus the fixed 4 trailing
octets (QTYPE + QCLASS).
