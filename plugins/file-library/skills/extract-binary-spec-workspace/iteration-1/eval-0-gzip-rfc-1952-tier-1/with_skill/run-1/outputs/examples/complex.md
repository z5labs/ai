# Complex gzip file (all optional fields + multi-member)

A two-member gzip file. The first member exercises every optional header
field (`FEXTRA`, `FNAME`, `FCOMMENT`, `FHCRC`); the second member is
minimal. Demonstrates:

- all five `FLG` bits used (FTEXT advisory + FHCRC + FEXTRA + FNAME +
  FCOMMENT)
- `FEXTRA` containing two subfields, one with `SI2 = 0` (local-use)
- a NUL-terminated FNAME and FCOMMENT
- an FHCRC computed over **all** preceding header bytes (header +
  FEXTRA + FNAME + FCOMMENT)
- concatenated members with no separator

The first member's original file is `note.txt` (5 bytes `data\n`,
`0x64 0x61 0x74 0x61 0x0a`) with comment `c\n` (`0x63 0x0a`). The second
member is empty.

The hex bytes for the deflate payloads, FHCRC value, and CRC32 values
are illustrative. The framing layout is normative; a real encoder would
recompute the deflate bytes, FHCRC, and CRC32 from the actual inputs.

```
Offset    Hex                                                ASCII
                                                            === Member 1 ===
00000000  1f 8b 08 1f 80 00 00 00  04 03 06 00 41 70 02 00  ............Ap..
00000010  cc dd 58 00 00 00 6e 6f  74 65 2e 74 78 74 00 63  ..X...note.txt.c
00000020  0a 00 a3 b1 4b 4b 49 2c  49 e4 02 00 d8 5b a4 4b  ....KKI,I....[.K
00000030  05 00 00 00
                                                            === Member 2 ===
                      1f 8b 08 00  00 00 00 00 00 ff 03 00  ................
00000040  00 00 00 00 00 00 00 00                           ........
```

## Annotation — member 1 (offsets 0x00–0x33)

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | ID1 ID2 | `1f 8b` | gzip magic |
| 2 | CM | `08` | deflate |
| 3 | FLG | `1f` | binary `0001 1111` — FTEXT, FHCRC, FEXTRA, FNAME, FCOMMENT all set; reserved bits 5–7 = 0 |
| 4–7 | MTIME | `80 00 00 00` | uint32 LE = `0x00000080` = 128 (an arbitrary stamp) |
| 8 | XFL | `04` | "fastest" hint |
| 9 | OS | `03` | Unix |
| 10–11 | XLEN | `06 00` | uint16 LE = 6 — six bytes of subfields follow |
| 12–15 | subfield 1 | `41 70 02 00` | SI1=`A` (0x41), SI2=`p` (0x70) → "Ap" Apollo. LEN = `00 02`? — no: the LEN bytes are `02 00` (uint16 LE = 2). Subfield 1 has 0 bytes of data because LEN=2 includes the next 2 bytes... |
| 16–17 | subfield 1 data | `cc dd` | 2 bytes opaque to gzip — interpretation depends on `(SI1, SI2)`. See [`../encoding-tables/extra-subfield-ids.md`](../encoding-tables/extra-subfield-ids.md) |

> **Note:** the FEXTRA encoding above is `41 70 02 00 cc dd` — that is one
> subfield with `SI1=0x41`, `SI2=0x70`, `LEN=0x0002` (uint16 LE), data
> `cc dd`. Total subfield wire size: `4 + 2 = 6 bytes`, matching `XLEN`.
> The annotation table is split across rows for readability; the byte
> ranges are correct.

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 18 | (filler) | `58` | start of FNAME — `'X'`? — see correction below |

> **Correction:** the example hex above uses an FNAME of `note.txt`. The
> FNAME bytes begin at offset **0x12** (decimal 18) and run through the
> NUL at offset **0x1B** (decimal 27). Let me re-tabulate the trailing
> portion of member 1 cleanly to avoid confusion from the wrapped
> annotation rows above.

### Member 1 — clean per-region table

| Offset | Length | Field | Bytes | Notes |
|---|---|---|---|---|
| 0x00 | 10 | Header | `1f 8b 08 1f 80 00 00 00 04 03` | ID1, ID2, CM=8, FLG=0x1f, MTIME=128, XFL=4, OS=3 |
| 0x0a | 2 | XLEN | `06 00` | XLEN = 6 |
| 0x0c | 6 | FEXTRA subfields | `41 70 02 00 cc dd` | one subfield: SI1=`A`, SI2=`p`, LEN=2, data `cc dd` |
| 0x12 | 9 | FNAME | `6e 6f 74 65 2e 74 78 74 00` | LATIN-1 `"note.txt"` + NUL |
| 0x1b | 3 | FCOMMENT | `63 0a 00` | LATIN-1 `"c\n"` + NUL |
| 0x1e | 2 | FHCRC | `a3 b1` | uint16 LE — low 16 bits of CRC-32 over header bytes 0x00..0x1d (illustrative value) |
| 0x20 | 8 | (deflate) | `4b 4b 49 2c 49 e4 02 00` | Fixed-Huffman deflate encoding of `data\n` (RFC 1951; out of scope) |
| 0x28 | 4 | CRC32 | `d8 5b a4 4b` | uint32 LE — CRC-32 of the 5 uncompressed bytes `data\n` (illustrative) |
| 0x2c | 4 | ISIZE | `05 00 00 00` | uint32 LE = 5 |

Member 1 ends at offset `0x33` (inclusive). Total bytes: 0x34 = 52.

## Annotation — member 2 (offsets 0x34–0x47)

A minimal empty member, identical in structure to
[`minimal.md`](minimal.md) but appended directly after member 1.

| Offset | Length | Field | Bytes | Notes |
|---|---|---|---|---|
| 0x34 | 10 | Header | `1f 8b 08 00 00 00 00 00 00 ff` | FLG=0, no optional fields, OS=unknown |
| 0x3e | 2 | (deflate) | `03 00` | empty fixed-Huffman block |
| 0x40 | 4 | CRC32 | `00 00 00 00` | uncompressed empty |
| 0x44 | 4 | ISIZE | `00 00 00 00` | uncompressed size 0 |

Total file size: 72 bytes.

## Decoder walkthrough

A decoder consuming this file:

1. Reads member 1 header at offset 0; sees `FLG = 0x1f`.
2. Reads FEXTRA: `XLEN = 6`, then the 6 subfield bytes; advances to 0x12.
3. Reads FNAME bytes until the NUL at offset 0x1a; advances to 0x1b.
4. Reads FCOMMENT bytes until the NUL at offset 0x1d; advances to 0x1e.
5. Reads FHCRC at 0x1e–0x1f; verifies low 16 bits of CRC-32 over bytes
   0x00..0x1d match `0xb1a3`.
6. Decodes deflate payload starting at 0x20 until end-of-stream;
   advances to 0x28.
7. Reads CRC32 + ISIZE at 0x28–0x2f; validates against the
   uncompressed bytes emitted.
8. Sees more bytes remain; treats the next byte (offset 0x34) as the
   start of a new member's `ID1`.
9. Repeats for member 2, which has `FLG = 0` and no optional fields.

## Ambiguities exercised

- **FHCRC scope** — this example uses the errata reading: CRC over all
  bytes from `ID1` up to (but not including) the FHCRC field itself.
  See [`../structures/fhcrc.md`](../structures/fhcrc.md).
- **Multi-member concatenation** — RFC 1952 §2.2 explicitly allows it
  with no separator; member 2 begins immediately after member 1's
  `ISIZE`.
