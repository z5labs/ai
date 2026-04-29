# RDATA: NS / CNAME / PTR / MD / MF / MB / MG / MR (single-name shape)

A family of RDATA payloads that contain exactly one domain name and
nothing else. Defined in RFC 1035 §3.3.11 (NS), §3.3.1 (CNAME), §3.3.12
(PTR), §3.3.4 (MD obsolete), §3.3.5 (MF obsolete), §3.3.3 (MB), §3.3.6
(MG), §3.3.8 (MR).

This file documents the wire shape; consult the field name column for
the per-TYPE meaning.

## Byte diagram

```
+-- NAME (variable; same encoding as a domain name in QNAME) -------+
|                                                                   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | Target | A domain name in wire form — see [`domain-name.md`](domain-name.md). For NS this is NSDNAME (the authoritative name server); for CNAME it is CNAME (the canonical name); for PTR it is PTRDNAME (the pointed-to domain); for MD/MF/MB/MG/MR see RFC 1035 §3.3 for per-type meanings. |

The enclosing RR's RDLENGTH equals the wire length of the encoded name.

## Bit fields

None.

## Variable-length fields

### Target

- **Length determination**: terminated by a root label (zero-length) or
  by a compression pointer.
- **Encoding**: see [`domain-name.md`](domain-name.md).
- **Compression**: permitted (RFC 1035 §4.1.4 names this TYPE family as
  compressible).

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

[`domain-name`](domain-name.md).

## Versioning notes

Unchanged. MD and MF are obsolete (replaced by MX) but their on-the-wire
shape is preserved for backward compatibility.

## Ambiguities

> **Ambiguity:** RFC 1035 declares MD (3) and MF (4) obsolete and says
> a server should reject queries for them with NOTIMP, but the on-the-
> wire format remains defined for compatibility with archived data.
