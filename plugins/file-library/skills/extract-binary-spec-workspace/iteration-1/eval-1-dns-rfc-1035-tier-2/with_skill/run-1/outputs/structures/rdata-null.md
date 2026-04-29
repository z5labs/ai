# RDATA: NULL (TYPE = 10)

The NULL RDATA payload — opaque bytes used for experimental purposes.
Defined in RFC 1035 §3.3.10.

## Byte diagram

```
+-- anything (0..65535 octets, opaque) ---------+
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | RDLENGTH | []byte | Data | Opaque bytes, up to 65535 octets. |

## Bit fields

None.

## Variable-length fields

### Data

- **Length determination**: from the enclosing RR's RDLENGTH.
- **Maximum length**: 65535 octets.
- **Encoding**: opaque.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

None.

## Versioning notes

NULL records cannot appear in zone master files (RFC 1035 §3.3.10) and
are explicitly experimental. Still defined.

## Ambiguities

None.
