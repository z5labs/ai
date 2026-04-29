# RDataMF

RDATA payload of a `Type=MF` (mail forwarder) resource record. **OBSOLETE** — superseded by MX (RFC 974). Defined here only for parity with on-wire decoders that may encounter legacy data. RFC 1035 §3.3.5.

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
| 0 | variable | DomainName | MADName | Domain name of a mail-forwarder host. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal the on-wire byte length of `MADName`.
- OBSOLETE per RFC 1035 §3.3.5 — encoders SHOULD NOT generate MF; the recommended migration is to translate MF to an MX with preference 10.
- Historically caused additional-section processing for an `A` record matching `MADName`.
