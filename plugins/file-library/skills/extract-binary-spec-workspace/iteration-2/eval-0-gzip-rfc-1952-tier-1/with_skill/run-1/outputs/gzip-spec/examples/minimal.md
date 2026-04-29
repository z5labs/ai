# Minimal valid gzip member

The smallest legal gzip member: empty original input, no optional metadata
fields, deflate compression. All `FLG` bits are clear, so the only blocks
present are the fixed 10-byte header, a 2-byte deflate stream encoding "empty
input", and the 8-byte trailer. Total: 20 bytes.

The body `03 00` is the canonical deflate encoding for an empty input: a
single final fixed-Huffman block containing only the end-of-block symbol.

```
Offset    Hex                                                ASCII
00000000  1f 8b 08 00 00 00 00 00  00 ff 03 00 00 00 00 00  ................
00000010  00 00 00 00                                       ....
```

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0     | ID1   | `0x1F` | gzip magic byte 1 |
| 1     | ID2   | `0x8B` | gzip magic byte 2 |
| 2     | CM    | `0x08` | DEFLATE — see [`../encoding-tables/compression-methods.md`](../encoding-tables/compression-methods.md) |
| 3     | FLG   | `0x00` | All flag bits clear — no optional fields follow |
| 4–7   | MTIME | `0x00000000` | "no timestamp available" |
| 8     | XFL   | `0x00` | unspecified deflate level |
| 9     | OS    | `0xFF` | UNKNOWN — see [`../encoding-tables/os-values.md`](../encoding-tables/os-values.md) |
| 10–11 | CompressedData | `03 00` | Canonical 2-byte deflate stream encoding the empty input |
| 12–15 | CRC32 | `0x00000000` | CRC-32 of the 0-byte uncompressed input is `0x00000000` |
| 16–19 | ISIZE | `0x00000000` | Uncompressed input length, 0 bytes |
