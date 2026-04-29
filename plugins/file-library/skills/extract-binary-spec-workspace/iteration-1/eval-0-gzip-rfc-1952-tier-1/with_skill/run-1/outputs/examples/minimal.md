# Minimal valid gzip member

The smallest legal gzip member: fixed 10-byte header, no optional fields,
a deflate stream encoding zero bytes of input, and the 8-byte trailer.
Twenty bytes total.

This example exercises:
- the magic bytes (`1f 8b`) and `CM = 8` (deflate)
- `FLG = 0x00` — no optional fields
- `MTIME = 0` — no time stamp
- `XFL = 0`, `OS = 0xff` (unknown)
- a minimal "empty input" deflate stream
- `CRC32 = 0` and `ISIZE = 0` for an empty payload

```
Offset    Hex                                                ASCII
00000000  1f 8b 08 00 00 00 00 00  00 ff 03 00 00 00 00 00  ................
00000010  00 00 00 00                                       ....
```

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | ID1 ID2 | `1f 8b` | gzip magic — see [`../structures/header.md`](../structures/header.md) |
| 2 | CM | `08` | deflate — see [`../encoding-tables/compression-methods.md`](../encoding-tables/compression-methods.md) |
| 3 | FLG | `00` | no optional fields set — see [`../structures/flg.md`](../structures/flg.md) |
| 4–7 | MTIME | `00 00 00 00` | uint32 LE — no time stamp |
| 8 | XFL | `00` | no compression-quality hint — see [`../encoding-tables/xfl-deflate.md`](../encoding-tables/xfl-deflate.md) |
| 9 | OS | `ff` | unknown — see [`../encoding-tables/operating-system.md`](../encoding-tables/operating-system.md) |
| 10–11 | (deflate) | `03 00` | Deflate fixed-Huffman BFINAL block carrying zero literal bytes (RFC 1951; out of scope here) |
| 12–15 | CRC32 | `00 00 00 00` | uint32 LE; CRC-32 of empty input = `0x00000000` — see [`../structures/trailer.md`](../structures/trailer.md) |
| 16–19 | ISIZE | `00 00 00 00` | uint32 LE; uncompressed size 0 |

This is the framing a Go encoder must produce for `gzip.Writer.Write(nil)
+ Close()` on empty input (modulo the OS byte, which Go writes as `255`).
