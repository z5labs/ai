# Fixed member header (RFC 1952 Section 2.3.1)

The fixed-length portion of the member header is **exactly 10 bytes**, in
this order:

| Offset | Size | Field | Type     | Notes |
|-------:|-----:|-------|----------|-------|
|      0 |    1 | ID1   | u8       | Magic byte 1. Must be `0x1f` (31, `\037`). |
|      1 |    1 | ID2   | u8       | Magic byte 2. Must be `0x8b` (139, `\213`). |
|      2 |    1 | CM    | u8       | Compression method. `0..7` reserved; `8` = DEFLATE. |
|      3 |    1 | FLG   | u8       | Flag bits. See `flg-bits.md`. |
|      4 |    4 | MTIME | u32 LE   | Modification time, Unix seconds. `0` = unset. |
|      8 |    1 | XFL   | u8       | eXtra flags, method-specific. |
|      9 |    1 | OS    | u8       | OS / filesystem on which compression took place. |

## Field details

### ID1, ID2 — gzip magic
- `ID1 = 0x1f`, `ID2 = 0x8b`.
- These are the gzip identification bytes. A compliant decoder must
  reject input that does not start (and that does not start every
  subsequent member) with these two bytes.

### CM — Compression Method
- Values `0..7` are reserved.
- Value `8` denotes DEFLATE, which is what real-world gzip uses.
- A decoder must reject CMs it does not understand.

### FLG — Flag byte
- Bit-packed flags. See `flg-bits.md` for full bit definitions.
- Bits 5, 6, 7 are reserved and must be zero.

### MTIME — Modification time
- Little-endian unsigned 32-bit integer.
- Seconds since the Unix epoch (00:00:00 UTC, 1 Jan 1970).
- If the source had no file mtime, the compressor uses the time at
  which compression started.
- `MTIME = 0` means "no time stamp is available". Decoders should treat
  zero as "unknown" rather than as 1970-01-01.
- Note: the field is unsigned; this format will saturate / wrap at
  2106-02-07 06:28:15 UTC. (The RFC does not specify behaviour past
  2^32 seconds; treat as out-of-band concern.)

### XFL — Extra flags
- Method-specific. For CM = 8 (DEFLATE), the gzip producer encodes
  speed/quality:
  - `XFL = 2` — slowest algorithm / maximum compression.
  - `XFL = 4` — fastest algorithm.
- For other CM values, contents are method-defined.
- Unknown values should be tolerated by decoders.

### OS — Operating system
- Identifies the source filesystem, which can hint at end-of-line
  conventions for text data. Defined values:

    | Value | Meaning |
    |------:|---------|
    |     0 | FAT filesystem (MS-DOS, OS/2, NT/Win32) |
    |     1 | Amiga |
    |     2 | VMS (or OpenVMS) |
    |     3 | Unix |
    |     4 | VM/CMS |
    |     5 | Atari TOS |
    |     6 | HPFS filesystem (OS/2, NT) |
    |     7 | Macintosh |
    |     8 | Z-System |
    |     9 | CP/M |
    |    10 | TOPS-20 |
    |    11 | NTFS filesystem (NT) |
    |    12 | QDOS |
    |    13 | Acorn RISCOS |
    |   255 | unknown |

- A compliant compressor that does not know the source OS should write
  `255` ("unknown").
- Decoders may freely ignore OS.
