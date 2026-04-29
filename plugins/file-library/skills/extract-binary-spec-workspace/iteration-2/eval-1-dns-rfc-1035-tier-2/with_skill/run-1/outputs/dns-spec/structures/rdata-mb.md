# RDataMB

RDATA payload of a `Type=MB` (mailbox) resource record (EXPERIMENTAL): a single domain name identifying a host that holds the specified mailbox. RFC 1035 §3.3.3.

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
| 0 | variable | DomainName | MADName | Domain name of a host that has the specified mailbox. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal the on-wire byte length of `MADName`.
- MB causes additional-section processing for an `A` record matching `MADName`.
- EXPERIMENTAL per RFC 1035 §3.3.3.
