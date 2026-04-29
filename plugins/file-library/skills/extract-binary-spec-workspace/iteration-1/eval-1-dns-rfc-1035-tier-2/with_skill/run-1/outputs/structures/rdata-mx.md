# RDATA: MX (TYPE = 15)

The Mail Exchange RDATA payload — a 16-bit preference followed by a
domain name. Defined in RFC 1035 §3.3.9.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                  PREFERENCE                   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
+-- EXCHANGE (domain-name) ---------------------+
|                                               |
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 | Preference | Lower values are preferred when multiple MX records exist for the same name. Big-endian. |
| 2 | variable | []byte | Exchange | Domain name of the mail-exchange host. See [`domain-name.md`](domain-name.md). |

## Bit fields

None.

## Variable-length fields

Exchange — see [`domain-name.md`](domain-name.md). Compression is
permitted (RFC 1035 §4.1.4).

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

[`domain-name`](domain-name.md).

## Versioning notes

Unchanged.

## Ambiguities

None.
