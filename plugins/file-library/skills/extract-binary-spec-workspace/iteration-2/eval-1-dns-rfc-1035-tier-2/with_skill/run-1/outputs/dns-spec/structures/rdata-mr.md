# RDataMR

RDATA payload of a `Type=MR` (mail rename) resource record (EXPERIMENTAL): a single domain name redirecting one mailbox to another. RFC 1035 §3.3.8.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   NEWNAME                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | NewName | Domain name of the rename target mailbox. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal the on-wire byte length of `NewName`.
- MR causes no additional-section processing.
- EXPERIMENTAL per RFC 1035 §3.3.8.
