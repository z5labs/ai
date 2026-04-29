# RDataMG

RDATA payload of a `Type=MG` (mail group member) resource record (EXPERIMENTAL): a single domain-name reference to a mailbox that is a member of the named mail group. RFC 1035 §3.3.6.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   MGMNAME                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | MGMName | Domain name of a mailbox that is a member of the mail group named by the owner. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal the on-wire byte length of `MGMName`.
- MG causes no additional-section processing.
- EXPERIMENTAL per RFC 1035 §3.3.6.
