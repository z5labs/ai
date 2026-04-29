# RDataMX

RDATA payload of a `Type=MX` resource record: a 16-bit preference followed by the domain name of a mail exchanger. RFC 1035 §3.3.9.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                  PREFERENCE                   |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                   EXCHANGE                    /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 | Preference | Preference value among MX records at the same owner; **lower is preferred**. RFC 1035 §3.3.9. Big-endian. |
| 2 | variable | DomainName | Exchange | Domain name of a host willing to act as a mail exchanger for the owner. See [`domain-name.md`](domain-name.md). |

## Notes

- `RDLength` MUST equal `2 + on-wire-length-of-Exchange`.
- Triggers additional-section processing of `A` records for `Exchange` (RFC 1035 §3.3.9).

## Ambiguities

> **Ambiguity:** RFC 1035 §3.3 enumerates NS, SOA, CNAME, and PTR as compressible RDATA but is silent about MX. In practice MX `Exchange` is universally compressed; decoders MUST handle compression here regardless.
