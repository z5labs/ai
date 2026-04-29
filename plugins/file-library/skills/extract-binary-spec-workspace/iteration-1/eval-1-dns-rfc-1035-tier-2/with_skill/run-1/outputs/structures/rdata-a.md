# RDATA: A (TYPE = 1)

The RDATA payload of an A resource record (Internet IPv4 address).
Defined in RFC 1035 §3.4.1.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                          ADDRESS  (32-bit IPv4 address, network byte order)                   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | [4]byte | Address | IPv4 address. Bytes appear in network order — `192.0.2.1` is `C0 00 02 01`. |

The enclosing RR's RDLENGTH must equal 4. The CLASS for an A record is
typically IN (1); CH (3) is permitted historically but the address
encoding under CH is different and is out of scope here.

## Bit fields

None.

## Variable-length fields

None — RDATA is exactly 4 octets.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

None.

## Versioning notes

Unchanged. AAAA records (IPv6, RFC 3596) use TYPE 28 with a 16-byte
RDATA and are out of scope here.

## Ambiguities

None.
