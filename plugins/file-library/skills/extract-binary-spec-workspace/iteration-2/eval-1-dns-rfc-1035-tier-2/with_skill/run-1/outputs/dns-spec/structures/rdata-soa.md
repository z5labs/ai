# RDataSOA

RDATA payload of a `Type=SOA` resource record: marks the start of a zone of authority and carries zone-management timers. RFC 1035 §3.3.13.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     MNAME                     /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
/                     RNAME                     /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    SERIAL                     |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    REFRESH                    |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                     RETRY                     |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    EXPIRE                     |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    MINIMUM                    |
|                                               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | DomainName | MName | Domain name of the original/primary source of data for this zone. See [`domain-name.md`](domain-name.md). |
| variable | variable | DomainName | RName | Mailbox of the person responsible for the zone, encoded as a domain name (the first label is the local part, the rest is the domain). |
| variable | 4 | uint32 | Serial | Unsigned 32-bit version number of the original zone copy. Compared using sequence-space arithmetic (wraps). |
| variable+4 | 4 | int32 | Refresh | Time interval in seconds before the zone should be refreshed. RFC 1035 §3.3.13: "32 bit time interval". |
| variable+8 | 4 | int32 | Retry | Seconds to wait between failed refresh retries. |
| variable+12 | 4 | int32 | Expire | Upper bound in seconds before the zone is no longer authoritative. |
| variable+16 | 4 | uint32 | Minimum | Lower bound on TTL exported with any RR from this zone. RFC 1035 §3.3.13: "unsigned 32 bit minimum TTL field". |

All multi-octet integers are big-endian.

## Notes

- `RDLength` MUST equal the sum of the on-wire byte sizes of `MName` and `RName` plus 20 (five 32-bit fields).
- Both `MName` and `RName` may be compressed (RFC 1035 §3.3 explicitly authorises compression for SOA).
- All time values are in seconds (RFC 1035 §3.3.13).

## Ambiguities

> **Ambiguity:** RFC 1035 §3.3.13 calls `Refresh`, `Retry`, and `Expire` "32 bit time interval"/"32 bit time value" without specifying signed vs. unsigned. The practical convention is unsigned (negative intervals are nonsense), but a decoder using `int32` matches the spec's TTL precedent in §3.2.1 and rejects negatives at the API. Either Go type is defensible; pick `int32` for symmetry with TTL or `uint32` for the natural value range.

> **Ambiguity:** `RName` is a `<domain-name>` per §3.3.13 but represents an email address whose first label is the mailbox local-part (e.g., `hostmaster.example.com.` means `hostmaster@example.com.`). A decoder MUST decode it as a domain name; presentation conversion to RFC 822 form is up to higher layers.
