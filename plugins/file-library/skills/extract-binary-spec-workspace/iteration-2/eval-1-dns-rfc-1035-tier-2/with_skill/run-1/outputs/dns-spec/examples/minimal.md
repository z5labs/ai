# Minimal valid DNS message

A standard query for `A` records of `example.com`. Header + one question, no answers, no compression. The smallest realistic DNS query you can decode end-to-end.

```
Offset    Hex                                                ASCII
00000000  12 34 01 00 00 01 00 00  00 00 00 00 07 65 78 61  .4...........exa
00000010  6d 70 6c 65 03 63 6f 6d  00 00 01 00 01           mple.com.....
```

Total length: 29 octets.

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | Header.ID | `0x1234` | Arbitrary transaction ID |
| 2–3 | Header.Flags | `0x0100` | QR=0, Opcode=0 (QUERY), AA=0, TC=0, RD=1, RA=0, Z=0, RCODE=0 |
| 4–5 | Header.QDCOUNT | `0x0001` | One question follows |
| 6–7 | Header.ANCOUNT | `0x0000` | No answers |
| 8–9 | Header.NSCOUNT | `0x0000` | No authority records |
| 10–11 | Header.ARCOUNT | `0x0000` | No additional records |
| 12 | QName label[0].Length | `0x07` (7) | First label length |
| 13–19 | QName label[0].Data | `65 78 61 6d 70 6c 65` | "example" |
| 20 | QName label[1].Length | `0x03` (3) | Second label length |
| 21–23 | QName label[1].Data | `63 6f 6d` | "com" |
| 24 | QName terminator | `0x00` | Root label, ends QName (uncompressed) |
| 25–26 | QType | `0x0001` | A (1) |
| 27–28 | QClass | `0x0001` | IN (1) |

## What this exercises

- 12-byte fixed header layout and big-endian count fields.
- Bit packing of the `Flags` word (RD set; everything else zero).
- A complete uncompressed `DomainName` (two labels and the root terminator).
- 16-bit `QType` and `QClass` fields.
