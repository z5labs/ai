# RDataMD

RDATA payload of a `Type=MD` (mail destination) resource record. **OBSOLETE** — superseded by MX (RFC 974). Defined here only for parity with on-wire decoders that may encounter legacy data. RFC 1035 §3.3.4.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   MADNAME                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | MADName | Domain name of a mail-delivery host. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal the on-wire byte length of `MADName`.
- OBSOLETE per RFC 1035 §3.3.4 — encoders SHOULD NOT generate MD; the recommended migration is to translate MD to an MX with preference 0 (RFC 1035 §3.3.4).
- Historically caused additional-section processing for an `A` record matching `MADName`.
