# ResourceRecord

The common format used for every entry in the Answer, Authority, and Additional sections. RFC 1035 §3.2.1, §4.1.3.

## Byte diagram

```
                                1  1  1  1  1  1
  0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                                               |
/                                               /
/                      NAME                     /
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      TYPE                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     CLASS                     |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                      TTL                      |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                   RDLENGTH                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     RDATA                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | Name | Owner name — the node to which this RR pertains. See [`domain-name.md`](domain-name.md). |
| variable | 2 | uint16 | Type | RR type code. See [`../encoding-tables/types.md`](../encoding-tables/types.md). |
| variable+2 | 2 | uint16 | Class | RR class code. See [`../encoding-tables/classes.md`](../encoding-tables/classes.md). |
| variable+4 | 4 | int32 | TTL | Cache lifetime in seconds. RFC 1035 §3.2.1: "a 32 bit signed integer". A value of 0 means the RR may be used only for the in-progress transaction and not cached. RFC 1035 §2.3.4 further constrains TTL to "positive values of a signed 32 bit number" — i.e., 0 ≤ TTL ≤ 2³¹−1. |
| variable+8 | 2 | uint16 | RDLength | Length in octets of the RDATA field that follows. RFC 1035 §3.2.1. |
| variable+10 | RDLength | []byte | RData | Type- and class-dependent payload. See per-type structure files (e.g. [`rdata-a.md`](rdata-a.md), [`rdata-ns.md`](rdata-ns.md), ...). |

All multi-octet integers are big-endian.

## Variable-length fields

- **Name length determination:** self-delimited via `DomainName` encoding (RFC 1035 §3.1, §4.1.4). The decoder must use the on-wire byte count of the name, **not** the expanded length, when computing offsets to the next field.
- **RData length determination:** explicit `RDLength` prefix, exclusive of itself, counted in octets.
- **RData length on the wire vs. expanded:** RFC 1035 §4.1.4 — when RDATA contains a compressed domain name, `RDLength` counts the **compressed** length (the bytes actually present on the wire), not the expanded form.

## Conditional / optional fields

- `RData` content is interpreted based on `(Type, Class)`:
  - For class `IN`, type-specific layouts are defined in §3.3 (CNAME, NS, MX, SOA, etc.) and §3.4 (A, WKS).
  - For unknown `(Type, Class)` pairs, decoders MUST still consume `RDLength` bytes verbatim and MAY surface them as opaque `[]byte`.

## Nested structures

- [`domain-name.md`](domain-name.md) — for `Name` and any embedded names inside `RData`.
- Per-type RDATA: see the `rdata-*.md` files indexed in [`SPEC.md`](../SPEC.md).

## Ambiguities

> **Ambiguity:** RFC 1035 §3.2.1 declares TTL "a 32 bit signed integer" while §4.1.3 says it is "a 32 bit unsigned integer", and §2.3.4 limits the value range to "positive values of a signed 32 bit number". The conventional resolution (and what BIND and other implementations do) is to encode it on the wire as a 32-bit big-endian unsigned integer but treat values with the high bit set as zero. A Go decoder safely uses `int32` (preserving sign for the 0..2³¹−1 valid range) and rejects negative values at the API boundary.
