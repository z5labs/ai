# Typical DNS response

Authoritative response to the minimal query: `example.com` resolves to one IPv4 address (`192.0.2.10`). The answer's `Name` uses a compression pointer back to the QName at offset 12. Demonstrates the most common case in production traffic.

```
Offset    Hex                                                ASCII
00000000  12 34 85 80 00 01 00 01  00 00 00 00 07 65 78 61  .4...........exa
00000010  6d 70 6c 65 03 63 6f 6d  00 00 01 00 01 c0 0c 00  mple.com........
00000020  01 00 01 00 00 0e 10 00  04 c0 00 02 0a           .............
```

Total length: 45 octets.

## Annotation

### Header (offsets 0–11)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | ID | `0x1234` | Same ID as the query |
| 2–3 | Flags | `0x8580` | QR=1, Opcode=0, AA=1, TC=0, RD=1, RA=1, Z=0, RCODE=0 |
| 4–5 | QDCOUNT | `0x0001` | Question echoed back |
| 6–7 | ANCOUNT | `0x0001` | One answer |
| 8–9 | NSCOUNT | `0x0000` | — |
| 10–11 | ARCOUNT | `0x0000` | — |

`Flags = 0x8580` decodes to `1000 0101 1000 0000` → bit 0 (QR) = 1, bits 1–4 (Opcode) = 0000, bit 5 (AA) = 1, bit 6 (TC) = 0, bit 7 (RD) = 1, bit 8 (RA) = 1, bits 9–11 (Z) = 000, bits 12–15 (RCODE) = 0000.

### Question (offsets 12–28)

Same as in [`minimal.md`](minimal.md): QName = `example.com.`, QType = `A`, QClass = `IN`.

### Answer #1 (offsets 29–44)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 29–30 | Name | `0xC0 0x0C` | Compression pointer: top 2 bits = `11`, OFFSET = 12 → resolves to QName `example.com.` |
| 31–32 | Type | `0x0001` | A |
| 33–34 | Class | `0x0001` | IN |
| 35–38 | TTL | `0x00000E10` | 3600 seconds |
| 39–40 | RDLength | `0x0004` | 4 octets of RDATA |
| 41–44 | RData.Address | `0xC0 0x00 0x02 0x0A` | `192.0.2.10` |

## What this exercises

- A response header with QR, AA, RD, RA all set and RCODE = NOERROR.
- A compression pointer (`0xC00C`) replacing the answer's owner name.
- The `RDataA` structure (4-byte IPv4 address in network order).
- TTL (`0x0E10` = 3600s); decoder treats it as unsigned non-negative.
