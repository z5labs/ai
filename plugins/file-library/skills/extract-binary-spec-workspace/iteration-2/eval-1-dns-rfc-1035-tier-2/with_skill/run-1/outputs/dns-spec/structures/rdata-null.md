# RDataNULL

RDATA payload of a `Type=NULL` resource record (EXPERIMENTAL): an opaque byte string of any length up to 65535. RFC 1035 §3.3.10.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                  <anything>                   /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | RDLength | []byte | Data | Opaque payload. RFC 1035 §3.3.10: "Anything at all may be in the RDATA field so long as it is 65535 octets or less." |

## Variable-length fields

- **Length determination:** the enclosing `RDLength` (`uint16`) bounds the payload exactly.
- **Maximum length:** 65535 octets (the maximum value of `RDLength`).

## Notes

- NULL records cause no additional-section processing.
- NULL RRs are not allowed in master files (RFC 1035 §3.3.10) — but this reference is silent on master files per user scope.
