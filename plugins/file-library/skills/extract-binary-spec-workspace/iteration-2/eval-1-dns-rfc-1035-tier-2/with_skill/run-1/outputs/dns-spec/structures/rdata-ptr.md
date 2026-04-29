# RDataPTR

RDATA payload of a `Type=PTR` resource record: a single domain name pointing into the domain name space. Used in the IN-ADDR.ARPA tree to map IPs back to names. RFC 1035 §3.3.12.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   PTRDNAME                    /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | PTRDName | Target domain name. May be compressed. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal the on-wire byte length of `PTRDName`.
- Domain name may be compressed (RFC 1035 §3.3 explicitly authorises compression for PTR).
- PTR causes no additional-section processing (RFC 1035 §3.3.12).
