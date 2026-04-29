# Complex DNS response

A response that exercises:
- multiple answers,
- an authority section with an NS record using compression,
- an additional section with an A glue record using compression,
- and an SOA in the authority section showing two embedded names.

Question: `www.example.com.` IN A.

Answers:
- `www.example.com.` IN A `192.0.2.1` (TTL 3600)

Authority:
- `example.com.` IN NS `ns1.example.com.` (TTL 86400)
- `example.com.` IN SOA `ns1.example.com.` `hostmaster.example.com.`
  serial=1, refresh=7200, retry=3600, expire=1209600, minimum=300
  (TTL 86400)

Additional:
- `ns1.example.com.` IN A `192.0.2.53` (TTL 3600)

The example is laid out so that the compression pointers used in each
record point to earlier offsets in the message.

```
Offset    Hex                                                ASCII
00000000  12 34 85 80 00 01 00 01  00 02 00 01 03 77 77 77  .4...........www
00000010  07 65 78 61 6d 70 6c 65  03 63 6f 6d 00 00 01 00  .example.com....
00000020  01 c0 0c 00 01 00 01 00  00 0e 10 00 04 c0 00 02  ................
00000030  01 c0 10 00 02 00 01 00  01 51 80 00 06 03 6e 73  .........Q....ns
00000040  31 c0 10 c0 10 00 06 00  01 00 01 51 80 00 23 c0  1..........Q..#.
00000050  3d 0a 68 6f 73 74 6d 61  73 74 65 72 c0 10 00 00  =.hostmaster....
00000060  00 01 00 00 1c 20 00 00  0e 10 00 12 75 00 00 00  ..... ......u...
00000070  01 2c c0 3d 00 01 00 01  00 00 0e 10 00 04 c0 00  .,.=............
00000080  02 35                                             .5
```

## Annotation

### Header (offsets 0..11)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | ID | `0x1234` | |
| 2–3 | Flags | `0x8580` | QR=1, AA=1, RD=1, RA=1, RCODE=0 |
| 4–5 | QDCOUNT | `0x0001` | one question |
| 6–7 | ANCOUNT | `0x0001` | one answer |
| 8–9 | NSCOUNT | `0x0002` | two authority RRs |
| 10–11 | ARCOUNT | `0x0001` | one additional RR |

### Question (offsets 12..32)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 12–28 | QNAME | `03 'www' 07 'example' 03 'com' 00` | name `www.example.com.` |
| 29–30 | QTYPE | `0x0001` | A |
| 31–32 | QCLASS | `0x0001` | IN |

Note: byte 12 begins the `www` label; byte 16 begins the `example`
label (`0x07`); byte 24 begins the `com` label (`0x03`); byte 28 is the
root terminator.

### Answer #1 (offsets 33..48): A record for www.example.com

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 33–34 | NAME | `0xC0 0x0C` | pointer to offset 12 (`www.example.com.`) |
| 35–36 | TYPE | `0x0001` | A |
| 37–38 | CLASS | `0x0001` | IN |
| 39–42 | TTL | `0x00000E10` | 3600 |
| 43–44 | RDLENGTH | `0x0004` | 4 bytes |
| 45–48 | RDATA | `C0 00 02 01` | 192.0.2.1 |

### Authority #1 (offsets 49..60): NS record for example.com

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 49–50 | NAME | `0xC0 0x10` | pointer to offset 16 (`example.com.`) |
| 51–52 | TYPE | `0x0002` | NS |
| 53–54 | CLASS | `0x0001` | IN |
| 55–58 | TTL | `0x00015180` | 86400 |
| 59–60 | RDLENGTH | `0x0006` | 6 bytes |
| 61–66 | RDATA (NSDNAME) | `03 'ns1' C0 10` | label `ns1` then a pointer to `example.com.` at offset 16 |

### Authority #2 (offsets 67..113): SOA record for example.com

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 67–68 | NAME | `0xC0 0x10` | pointer to offset 16 (`example.com.`) |
| 69–70 | TYPE | `0x0006` | SOA |
| 71–72 | CLASS | `0x0001` | IN |
| 73–76 | TTL | `0x00015180` | 86400 |
| 77–78 | RDLENGTH | `0x0023` | 35 bytes (2 MNAME + 13 RNAME + 20 fixed integers) |
| 79–80 | RDATA.MNAME | `0xC0 0x3D` | pointer to offset 61 (`ns1.example.com.` — the name written into Authority #1's RDATA) |
| 81–93 | RDATA.RNAME | `0A 'hostmaster' C0 10` | label `hostmaster` then a pointer to `example.com.` at offset 16 |
| 94–97 | RDATA.SERIAL | `0x00000001` | 1 |
| 98–101 | RDATA.REFRESH | `0x00001C20` | 7200 |
| 102–105 | RDATA.RETRY | `0x00000E10` | 3600 |
| 106–109 | RDATA.EXPIRE | `0x00127500` | 1209600 |
| 110–113 | RDATA.MINIMUM | `0x0000012C` | 300 |

### Additional #1 (offsets 114..129): A record for ns1.example.com

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 114–115 | NAME | `0xC0 0x3D` | pointer to offset 61 (`ns1.example.com.`) |
| 116–117 | TYPE | `0x0001` | A |
| 118–119 | CLASS | `0x0001` | IN |
| 120–123 | TTL | `0x00000E10` | 3600 |
| 124–125 | RDLENGTH | `0x0004` | 4 bytes |
| 126–129 | RDATA | `C0 00 02 35` | 192.0.2.53 |

Total length: 130 bytes.

### Notes

- Pointers `0xC0 0x0C`, `0xC0 0x10`, and `0xC0 0x3D` all use the
  high-2-bits-set form. The 14-bit offset is computed as
  `((byte0 & 0x3F) << 8) | byte1`, giving 12, 16, and 61 respectively.
- Within Authority #1's RDATA the name `ns1.example.com.` is encoded
  with a literal `ns1` label followed by a pointer to `example.com.` at
  offset 16. That whole 6-byte fragment (`03 'n' 's' '1' C0 10`) starts
  at offset 61 and is itself the target for the SOA's MNAME pointer
  (`0xC0 0x3D` = 61) and the Additional A record's NAME pointer.
- All TTL values fit in 31 bits, so the int32 vs uint32 distinction
  does not matter for the bytes shown.
