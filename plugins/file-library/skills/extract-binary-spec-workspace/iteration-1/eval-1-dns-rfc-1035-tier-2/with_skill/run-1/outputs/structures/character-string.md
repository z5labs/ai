# Character String

The `<character-string>` primitive used inside RDATA payloads such as
TXT, HINFO, and the WKS service list. Defined in RFC 1035 §3.3.

## Byte diagram

```
+--------+----------------------------------+
| length |   length octets of payload       |
+--------+----------------------------------+
   1 byte         0..255 octets
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | uint8 | Length | Number of payload octets that follow. Range 0..255. |
| 1 | Length | []byte | Data | Opaque 8-bit octets. |

A `<character-string>` has total wire length `1 + Length` bytes.

## Bit fields

None.

## Variable-length fields

### Data

- **Length determination**: explicit single-byte length prefix.
- **Length prefix counts**: only the payload, not itself.
- **Maximum length**: 255 octets (the prefix is `uint8`).
- **Encoding**: opaque bytes. Often interpreted as ASCII or UTF-8 by
  application code, but the wire format is byte-transparent.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

None.

## Versioning notes

Unchanged since RFC 1035.

## Ambiguities

None.
