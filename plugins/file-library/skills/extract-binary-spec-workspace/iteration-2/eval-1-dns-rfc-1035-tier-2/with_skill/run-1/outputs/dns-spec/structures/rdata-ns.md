# RDataNS

RDATA payload of a `Type=NS` resource record: a single domain name identifying an authoritative name server for the owner zone. RFC 1035 §3.3.11.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   NSDNAME                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | NSDName | Domain name of an authoritative name server. May be compressed. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` in the enclosing `ResourceRecord` MUST equal the on-wire byte length of `NSDName`.
- Domain name may be compressed (RFC 1035 §3.3 explicitly authorises compression for NS).
- Triggers additional-section processing for an `A` record matching `NSDName` (RFC 1035 §3.3.11).
