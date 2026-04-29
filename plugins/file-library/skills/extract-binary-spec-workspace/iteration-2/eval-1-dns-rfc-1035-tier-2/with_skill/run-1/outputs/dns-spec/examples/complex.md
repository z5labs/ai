# Complex DNS response with multiple sections, MX RDATA, and chained compression

Authoritative response to a query for `MX` records of `example.com.`. The response includes:

- 1 question (`example.com. MX IN`)
- 2 answer MX records (priorities 10 and 20)
- 1 authority NS record
- 1 additional A record

Compression appears in three places:
- The two answer MX records and the authority NS record share an owner name via a pointer to the QName.
- The MX `Exchange` `mail.example.com.` (and `mail2.example.com.`) is encoded as `[label] [pointer]` — chained compression: a fresh label followed by a pointer.
- The additional record's owner Name is a pointer **into the middle of an earlier RDATA** (the second MX `Exchange`), illustrating that compression targets need not be earlier owner names.

This is the kind of message a real authoritative server might emit; it stresses the decoder's pointer chasing, the `RDataMX` `Preference + Exchange` layout, and `RDLength`-vs-expanded-name accounting.

```
Offset    Hex                                                ASCII
00000000  ab cd 84 00 00 01 00 02  00 01 00 01 07 65 78 61  .............exa
00000010  6d 70 6c 65 03 63 6f 6d  00 00 0f 00 01 c0 0c 00  mple.com........
00000020  0f 00 01 00 00 0e 10 00  09 00 0a 04 6d 61 69 6c  ............mail
00000030  c0 0c c0 0c 00 0f 00 01  00 00 0e 10 00 0a 00 14  ................
00000040  05 6d 61 69 6c 32 c0 0c  c0 0c 00 02 00 01 00 00  .mail2..........
00000050  0e 10 00 06 03 6e 73 31  c0 0c c0 40 00 01 00 01  .....ns1...@....
00000060  00 00 0e 10 00 04 c0 00  02 19                    ..........
```

Total length: 106 octets (decimal).

## Annotation

### Header (offsets 0–11)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | ID | `0xABCD` | Transaction ID |
| 2–3 | Flags | `0x8400` | QR=1, Opcode=0, AA=1, TC=0, RD=0, RA=0, Z=0, RCODE=0 |
| 4–5 | QDCOUNT | `0x0001` | 1 question |
| 6–7 | ANCOUNT | `0x0002` | 2 answers (the two MX records) |
| 8–9 | NSCOUNT | `0x0001` | 1 authority record |
| 10–11 | ARCOUNT | `0x0001` | 1 additional record |

### Question (offsets 12–28)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 12–24 | QName | `07 "example" 03 "com" 00` | `example.com.`, uncompressed (13 octets) |
| 25–26 | QType | `0x000F` | MX (15) |
| 27–28 | QClass | `0x0001` | IN |

### Answer #1 — MX 10 mail.example.com. (offsets 29–49)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 29–30 | Name | `0xC0 0x0C` | Pointer → offset 12 (`example.com.`) |
| 31–32 | Type | `0x000F` | MX |
| 33–34 | Class | `0x0001` | IN |
| 35–38 | TTL | `0x00000E10` | 3600 |
| 39–40 | RDLength | `0x0009` | 9 octets of RDATA |
| 41–42 | RData.Preference | `0x000A` | 10 |
| 43–49 | RData.Exchange | `04 "mail" C0 0C` | Label "mail" (5 bytes) + pointer (2 bytes) → offset 12; expands to `mail.example.com.` |

`RDLength = 9` = 2 (Preference) + 5 (`"mail"` label as 1+4 bytes) + 2 (pointer). Note the on-wire compressed length, not the expanded length.

### Answer #2 — MX 20 mail2.example.com. (offsets 50–71)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 50–51 | Name | `0xC0 0x0C` | Pointer → offset 12 |
| 52–53 | Type | `0x000F` | MX |
| 54–55 | Class | `0x0001` | IN |
| 56–59 | TTL | `0x00000E10` | 3600 |
| 60–61 | RDLength | `0x000A` | 10 octets |
| 62–63 | RData.Preference | `0x0014` | 20 |
| 64–71 | RData.Exchange | `05 "mail2" C0 0C` | Label "mail2" (6 bytes) + pointer (2 bytes) → offset 12; expands to `mail2.example.com.` |

### Authority — NS ns1.example.com. (offsets 72–89)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 72–73 | Name | `0xC0 0x0C` | Pointer → offset 12 (`example.com.`) |
| 74–75 | Type | `0x0002` | NS |
| 76–77 | Class | `0x0001` | IN |
| 78–81 | TTL | `0x00000E10` | 3600 |
| 82–83 | RDLength | `0x0006` | 6 octets |
| 84–89 | RData.NSDName | `03 "ns1" C0 0C` | `ns1.example.com.` |

### Additional — A mail2.example.com. → 192.0.2.25 (offsets 90–105)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 90–91 | Name | `0xC0 0x40` | Pointer → offset 64 (start of the `"mail2"` label inside Answer #2's Exchange — offset 64 holds the label length octet `0x05`); decoder reads `"mail2"` then follows the inner pointer at 70–71 to offset 12, yielding `mail2.example.com.` |
| 92–93 | Type | `0x0001` | A |
| 94–95 | Class | `0x0001` | IN |
| 96–99 | TTL | `0x00000E10` | 3600 |
| 100–101 | RDLength | `0x0004` | 4 octets |
| 102–105 | RData.Address | `0xC0 0x00 0x02 0x19` | `192.0.2.25` |

## What this exercises

- Multiple sections (Question + 2 Answers + Authority + Additional).
- Three different compression scenarios: pure pointer, label-then-pointer chain, and a pointer into the middle of an earlier RDATA.
- `RDataMX` with both `Preference` and a compressed `Exchange`, and the resulting `RDLength` accounting.
- `RDataNS` with a compressed `NSDName` that itself ends in a pointer.
- Decoder sanity check: pointer-loop guard (decoder must reject any pointer that resolves back to itself or revisits an offset already followed in the current name).
