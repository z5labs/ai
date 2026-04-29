# Minimal valid DNS message

The smallest reasonable DNS message: a standard recursive query for the
A record of the root zone (`.`). 17 bytes total.

The query has ID `0x1234`, Flags `0x0100` (QR=0, Opcode=0 QUERY, RD=1),
QDCOUNT=1, all other counts zero, and a single Question with QNAME =
root (one zero byte), QTYPE=A (1), QCLASS=IN (1).

```
Offset    Hex                                                ASCII
00000000  12 34 01 00 00 01 00 00  00 00 00 00 00 00 01 00  .4..............
00000010  01                                                .
```

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | Header.ID | `0x1234` | Caller-chosen identifier |
| 2–3 | Header.Flags | `0x0100` | QR=0, Opcode=0 (QUERY), AA=0, TC=0, RD=1, RA=0, Z=0, RCODE=0 |
| 4–5 | Header.QDCOUNT | `0x0001` | exactly one question follows |
| 6–7 | Header.ANCOUNT | `0x0000` | no answers |
| 8–9 | Header.NSCOUNT | `0x0000` | no authority RRs |
| 10–11 | Header.ARCOUNT | `0x0000` | no additional RRs |
| 12 | Question.QNAME | `0x00` | root label (zero-length terminator only) |
| 13–14 | Question.QTYPE | `0x0001` | A |
| 15–16 | Question.QCLASS | `0x0001` | IN |

Total length: 17 bytes (12 header + 1 QNAME + 2 QTYPE + 2 QCLASS).
