# Question

A single entry in the Question section of a DNS message. The header's
QDCOUNT field gives the number of these that follow it. Standard queries
have exactly one question; the format permits more but RFC 1035 §4.1.2
notes that this is essentially never used. Defined in RFC 1035 §4.1.2.

## Byte diagram

The Question entry has a variable-length leading field (QNAME) followed
by two 16-bit fixed fields:

```
+-- QNAME (variable; sequence of length-prefixed labels, terminated by
|         a zero-length root label or a 2-byte compression pointer) --+
|                                                                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     QTYPE                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     QCLASS                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | QNAME | Domain name being queried, in wire form — see [`domain-name.md`](domain-name.md). |
| QNAME end | 2 | uint16 | QTYPE | Type of the query — see [`../encoding-tables/qtypes.md`](../encoding-tables/qtypes.md). |
| QNAME end + 2 | 2 | uint16 | QCLASS | Class of the query — see [`../encoding-tables/qclasses.md`](../encoding-tables/qclasses.md). |

Both QTYPE and QCLASS are 16-bit unsigned values stored in network byte
order. QTYPE is a superset of TYPE (it also accepts AXFR, MAILB, MAILA
and the wildcard `*`); QCLASS is a superset of CLASS (it also accepts
the wildcard `*`).

## Bit fields

None.

## Variable-length fields

### QNAME

- **Length determination**: terminated either by a single octet of value
  0 (the root label) or by a compression pointer whose first byte has
  the two high bits set (`0b11xxxxxx`). The pointer form occupies
  exactly two bytes and ends QNAME.
- **Encoding**: see [`domain-name.md`](domain-name.md). Each label is
  prefixed by a single length byte (1–63); a length byte with the two
  high bits set introduces a 14-bit back-reference into the same
  message.
- **Maximum length**: the wire-form name is bounded at 255 octets
  including the terminating root label or compression pointer.
- **Character encoding**: each label octet is an opaque byte. RFC 1035
  §2.3.1 defines the "preferred name syntax" (LDH — letters, digits,
  hyphens, ASCII case-insensitive) but the wire format itself does not
  enforce this; a decoder must accept any 8-bit values.

QTYPE and QCLASS are fixed-size and follow QNAME immediately with no
padding.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None — the question entry is byte-packed. There is no alignment
requirement, and entries are concatenated directly when QDCOUNT > 1.

## Nested structures

QNAME is a [`domain-name`](domain-name.md). Multiple Question entries (if
QDCOUNT > 1) are written end-to-end with no separator.

## Versioning notes

The structure has not changed since RFC 1035. New QTYPE / QCLASS values
are assigned through IANA but the on-wire encoding is unchanged.

## Ambiguities

> **Ambiguity:** RFC 1035 §4.1.2 allows QDCOUNT > 1 in principle, but
> §4.1.4 and most server behavior assume exactly one question per
> message. A receiver MAY reject messages with QDCOUNT != 1; a strict
> decoder should at least process QDCOUNT entries even though it is
> unusual.
