# RDataWKS

RDATA payload of a `Type=WKS` (Well-Known Services) resource record: an IPv4 address, an IP protocol number, and a bitmap of well-known service ports for that protocol. RFC 1035 §3.4.2.

## Byte diagram

```
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|                    ADDRESS                    |
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
|       PROTOCOL        |                       |
+--+--+--+--+--+--+--+--+                       |
|                                               |
/                   <BIT MAP>                   /
/                                               /
+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | [4]byte | Address | IPv4 address octets in network order. |
| 4 | 1 | uint8 | Protocol | IP protocol number (e.g. 6 = TCP, 17 = UDP). |
| 5 | RDLength − 5 | []byte | BitMap | Bit map, one bit per port. The first octet's MSB (bit 0) corresponds to port 0, the next bit (bit 1) to port 1, and so on. Length MUST be a multiple of 8 bits (i.e., the byte length is unconstrained but each byte covers 8 ports). RFC 1035 §3.4.2. |

All multi-octet integers are big-endian.

## Variable-length fields

- **BitMap length determination:** computed as `RDLength − 5`. RFC 1035 §3.4.2 states the bit map "must be a multiple of 8 bits long" — the byte count is unconstrained beyond that. Bits not represented in the map are implicitly zero (the service is not available).
- **Bit numbering:** within a byte, bit 0 is the **most significant** bit (matches RFC 1035 §2.3.2 — "the bit labeled 0 is the most significant bit"). So if `BitMap[3]` has bit 1 (mask `0x40`) set and `Protocol` is 6 (TCP), then TCP port 25 (= 3*8 + 1) is open. RFC 1035 §3.4.2 example: "if PROTOCOL=TCP (6), the 26th bit corresponds to TCP port 25 (SMTP)."

## Ambiguities

> **Ambiguity:** RFC 1035 §3.4.2 says the bit map "must be a multiple of 8 bits long" but does not state how a decoder should handle a `RDLength` that yields a non-byte-aligned bit map (i.e., none — since RDLength is a byte count, this is automatic). Treat any `RDLength < 5` as malformed.

> **Ambiguity:** RFC 1035 §3.4.2 example "the 26th bit corresponds to TCP port 25" is one-indexed counting from "first bit = port 0". Implementers should verify their bit ordering against this example before shipping.
