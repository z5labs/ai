# Typical gzip member with FNAME

A gzip member that records its original file name. Empty input is used so
every byte of the framing remains independently verifiable. `FLG.FNAME` is
the only optional flag set, which inserts a zero-terminated ISO 8859-1 name
between the fixed header and the compressed data.

Original file name: `"a.txt"` (5 bytes plus 1-byte NUL terminator).

```
Offset    Hex                                                ASCII
00000000  1f 8b 08 08 c0 80 1d 5a  00 03 61 2e 74 78 74 00  .......Z..a.txt.
00000010  03 00 00 00 00 00 00 00  00 00                    ..........
```

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0     | ID1   | `0x1F` | gzip magic byte 1 |
| 1     | ID2   | `0x8B` | gzip magic byte 2 |
| 2     | CM    | `0x08` | DEFLATE |
| 3     | FLG   | `0x08` | Only `FNAME` (bit 3) set — see [`../structures/flg.md`](../structures/flg.md) |
| 4–7   | MTIME | `0x5A1D80C0` (LE bytes `c0 80 1d 5a`) | Some Unix timestamp (decimal 1511868096 ≈ 2017-11-28). Stored little-endian. |
| 8     | XFL   | `0x00` | Unspecified |
| 9     | OS    | `0x03` | UNIX — see [`../encoding-tables/os-values.md`](../encoding-tables/os-values.md) |
| 10–14 | FName.Name | `61 2e 74 78 74` | ISO 8859-1 bytes `'a' '.' 't' 'x' 't'` |
| 15    | FName.Terminator | `0x00` | NUL terminator |
| 16–17 | CompressedData | `03 00` | Canonical 2-byte deflate stream for empty input |
| 18–21 | CRC32 | `0x00000000` | CRC-32 of empty uncompressed input |
| 22–25 | ISIZE | `0x00000000` | Uncompressed length 0 |
