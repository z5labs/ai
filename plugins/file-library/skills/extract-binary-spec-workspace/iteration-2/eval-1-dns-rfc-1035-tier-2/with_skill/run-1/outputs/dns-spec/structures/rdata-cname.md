# RDataCNAME

RDATA payload of a `Type=CNAME` resource record: a single domain name giving the canonical (primary) name for the owner alias. RFC 1035 §3.3.1.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     CNAME                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | CName | Canonical name. May be compressed. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal the on-wire byte length of `CName`.
- Domain name may be compressed (RFC 1035 §3.3 explicitly authorises compression for CNAME).
- A CNAME RR causes no additional-section processing, but a name server MAY restart resolution at the canonical name.
