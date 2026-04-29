# DNS Message — Overview

Source: RFC 1035 §4.1.

A DNS message is a single flat byte sequence. Every message has the same
top-level layout, regardless of whether it is a query or a response:

```
+---------------------+
|        Header       |  fixed 12 octets
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

- The Header is always present.
- The Header contains four 16-bit count fields (QDCOUNT, ANCOUNT, NSCOUNT,
  ARCOUNT) that say how many entries the four following sections contain.
  Any of those counts may be zero, in which case that section is empty.
- The Question section contains QDCOUNT entries (typically 1 in a query).
  Its entries are not resource records; they have their own format
  (QNAME / QTYPE / QCLASS).
- Answer, Authority, and Additional all share the same per-entry format:
  the standard Resource Record (RR) format described in §4.1.3.

## Byte order

RFC 1035 §2.3.2: All multi-octet integer fields are transmitted in network
byte order (big-endian), most significant octet first. Within a single
octet, bit 0 in the diagrams is the most significant bit.

Implementation note: Go's `encoding/binary.BigEndian` matches this exactly
for `Uint16` / `Uint32` reads and writes.

## Size limits

RFC 1035 §2.3.4:

- Label: 63 octets or less (the high two bits of a label length byte are
  reserved, leaving 6 bits for length, hence 0..63).
- Domain name: 255 octets or less (sum of all label-length bytes plus all
  label bytes plus the terminating zero byte).
- TTL: positive values of a signed 32-bit number (0 .. 2^31 - 1).
- UDP message: 512 octets or less. Longer messages are truncated and the
  TC bit is set in the header. (Transport framing itself is out of scope
  for this spec.)

## Character case

RFC 1035 §2.3.3: Label comparisons are case-insensitive, but original case
should be preserved on the wire. A decoder/encoder does not need to
normalize case; it only needs to round-trip it.
