# RDataA

RDATA payload of a `Type=A`, `Class=IN` resource record: a single 32-bit IPv4 address. RFC 1035 §3.4.1.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ADDRESS                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

(One 32-bit field, transmitted as 4 octets in network byte order.)

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | [4]byte | Address | IPv4 address octets in network order: `Address[0]` is the most significant octet (e.g., `10.2.0.52` encodes as `0x0A 0x02 0x00 0x34`). |

## Notes

- `RDLength` in the enclosing `ResourceRecord` MUST equal 4.
- A host with multiple IPv4 addresses appears as multiple A records, one per address (RFC 1035 §3.4.1).
