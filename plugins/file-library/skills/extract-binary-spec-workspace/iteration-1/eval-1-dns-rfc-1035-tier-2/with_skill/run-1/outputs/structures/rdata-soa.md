# RDATA: SOA (TYPE = 6)

The Start Of Authority RDATA payload — one per zone, returned with
authoritative answers and during zone transfers. Defined in RFC 1035
§3.3.13.

## Byte diagram

```
+-- MNAME (domain-name; primary master server)                  --+
|                                                                 |
+-- RNAME (domain-name; responsible person mailbox)              -+
|                                                                 |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    SERIAL (high)              |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    SERIAL (low)               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    REFRESH (high)             |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    REFRESH (low)              |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    RETRY (high)               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    RETRY (low)                |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    EXPIRE (high)              |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    EXPIRE (low)               |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    MINIMUM (high)             |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    MINIMUM (low)              |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | MNAME | Domain name of the primary master name server for this zone. See [`domain-name.md`](domain-name.md). |
| MNAME end | variable | []byte | RNAME | Domain name of the responsible-person mailbox. The first label encodes the local-part (with `.` represented in the local-part by escaping rules at the master-file layer; on the wire it is just a label). See [`domain-name.md`](domain-name.md). |
| RNAME end | 4 | uint32 | Serial | 32-bit serial number used by secondaries to detect zone changes. Compared with sequence-number arithmetic (RFC 1982). |
| RNAME end + 4 | 4 | int32 | Refresh | Seconds between refresh polls. RFC 1035 says "32 bit time interval"; standard practice and RFC 2181 §8 treat all four time intervals here as signed `int32`. |
| RNAME end + 8 | 4 | int32 | Retry | Seconds before retrying a failed refresh. |
| RNAME end + 12 | 4 | int32 | Expire | Seconds after which a secondary that cannot reach the master should treat the zone as expired. |
| RNAME end + 16 | 4 | uint32 | Minimum | Originally minimum TTL for records from this zone; redefined by RFC 2308 as the negative-cache TTL. |

The enclosing RR's RDLENGTH equals MNAME wire length + RNAME wire length
+ 20.

## Bit fields

None.

## Variable-length fields

MNAME and RNAME — see [`domain-name.md`](domain-name.md). Compression is
permitted within RDATA for SOA per RFC 1035 §4.1.4.

## Conditional / optional fields

None.

## Checksums and integrity

None.

## Padding and alignment

None.

## Nested structures

Two [`domain-name`](domain-name.md) fields followed by five 32-bit
integers.

## Versioning notes

Field layout unchanged. MINIMUM's *meaning* changed in RFC 2308.

## Ambiguities

> **Ambiguity:** Whether REFRESH/RETRY/EXPIRE are signed or unsigned is
> arguable from RFC 1035 alone. They are documented as positive
> intervals; using `int32` mirrors the TTL convention in RFC 2181.
