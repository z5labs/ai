# RDataTXT

RDATA payload of a `Type=TXT` resource record: one or more length-prefixed character strings. RFC 1035 §3.3.14.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   TXT-DATA                    /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | RDLength | []CharacterString | TXTData | One or more concatenated `<character-string>` items. See [`character-string.md`](character-string.md). |

## Variable-length fields

- **Length determination:** the enclosing `RDLength` from `ResourceRecord` bounds the entire payload; within that range the decoder reads `CharacterString` items back-to-back until exactly `RDLength` bytes have been consumed.
- **Minimum count:** RFC 1035 §3.3.14 says "One or more `<character-string>`s", so a TXT RR with `RDLength == 0` is malformed.

## Ambiguities

> **Ambiguity:** RFC 1035 says "one or more"; in practice some encoders emit a single empty `CharacterString` (one byte: `0x00`) for "empty" TXT records. RFC 1035 itself forbids zero `<character-string>`s but is silent on whether a single empty one counts.
