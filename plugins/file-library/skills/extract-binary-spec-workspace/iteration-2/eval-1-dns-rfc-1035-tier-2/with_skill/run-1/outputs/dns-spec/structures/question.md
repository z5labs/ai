# Question

A single entry in the Question section of a DNS message. Carries the parameters of the query: the name being asked about, the record type, and the class. RFC 1035 §4.1.2.

## Byte diagram

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

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | QName | Domain name being queried; sequence of length-prefixed labels possibly terminated by a compression pointer. See [`domain-name.md`](domain-name.md). |
| variable | 2 | uint16 | QType | Record type or QType being asked. See [`../encoding-tables/qtypes.md`](../encoding-tables/qtypes.md). |
| variable+2 | 2 | uint16 | QClass | Class. Almost always `1` (`IN`). See [`../encoding-tables/qclasses.md`](../encoding-tables/qclasses.md). |

QType and QClass are big-endian uint16. QName has no length prefix and no padding — RFC 1035 §4.1.2: "this field may be an odd number of octets; no padding is used."

## Variable-length fields

- **QName length determination:** self-delimiting per `DomainName` encoding (label-length octets and an optional compression pointer terminate the name). The decoder advances past QName by the on-wire byte count returned from domain-name decoding. RFC 1035 §4.1.2.
- **Maximum QName size:** 255 octets on the wire (uncompressed equivalent). RFC 1035 §2.3.4.

## Nested structures

- [`domain-name.md`](domain-name.md) — for `QName`.

## Notes

- `QType` is a superset of `Type` (RFC 1035 §3.2.3): all wire `Type` values are valid `QType` values, plus `AXFR (252)`, `MAILB (253)`, `MAILA (254)`, and `* (255)`.
- `QClass` is a superset of `Class` (RFC 1035 §3.2.5): adds `* (255)` (any class).
