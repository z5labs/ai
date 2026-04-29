# Typical DNS response

A standard A-record response to a query for `www.example.com.`. Exercises
header flags for an authoritative response and label compression in the
Answer section's NAME (the answer points back into the question's QNAME
rather than re-encoding the name).

The response carries one Question (echoed back) and one Answer record:
A `www.example.com.` -> `192.0.2.1` with a TTL of 3600 seconds.

```
Offset    Hex                                                ASCII
00000000  12 34 85 80 00 01 00 01  00 00 00 00 03 77 77 77  .4...........www
00000010  07 65 78 61 6d 70 6c 65  03 63 6f 6d 00 00 01 00  .example.com....
00000020  01 c0 0c 00 01 00 01 00  00 0e 10 00 04 c0 00 02  ................
00000030  01                                                .
```

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | Header.ID | `0x1234` | Echoed from the original query |
| 2–3 | Header.Flags | `0x8580` | QR=1, Opcode=0, AA=1, TC=0, RD=1, RA=1, Z=0, RCODE=0 |
| 4–5 | Header.QDCOUNT | `0x0001` | One question |
| 6–7 | Header.ANCOUNT | `0x0001` | One answer |
| 8–9 | Header.NSCOUNT | `0x0000` | No authority |
| 10–11 | Header.ARCOUNT | `0x0000` | No additional |
| 12–28 | Question.QNAME | `03 'www' 07 'example' 03 'com' 00` | Three labels + root terminator |
| 29–30 | Question.QTYPE | `0x0001` | A |
| 31–32 | Question.QCLASS | `0x0001` | IN |
| 33–34 | Answer.NAME | `0xC0 0x0C` | Compression pointer: top two bits `11`, offset `0x000C` = 12, which is the first byte of the Question.QNAME — references `www.example.com` |
| 35–36 | Answer.TYPE | `0x0001` | A |
| 37–38 | Answer.CLASS | `0x0001` | IN |
| 39–42 | Answer.TTL | `0x00000E10` | 3600 seconds |
| 43–44 | Answer.RDLENGTH | `0x0004` | 4 bytes of RDATA |
| 45–48 | Answer.RDATA | `C0 00 02 01` | IPv4 192.0.2.1 |

Total length: 49 bytes.

### Notes on compression

`0xC0 0x0C` in bytes 33–34 is the canonical compression pointer. In
binary the first byte is `11000000`: the two high bits flag a pointer,
and the remaining 14 bits across both bytes form the offset
`(0xC0 & 0x3F) << 8 | 0x0C` = `0 | 12` = 12. Byte offset 12 is the
first byte of QNAME (the `0x03` length prefix of the `www` label).
The decoder follows the pointer, walks the labels until the root
terminator, and obtains `www.example.com.`.
