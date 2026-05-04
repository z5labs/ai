# Typical DSF: one terrain definition, one POOL/SCAL pair, one triangle

A DSF that exercises the common path used by every real-world tile that
contains terrain:

- `HEAD/PROP` with the four required edge properties.
- `DEFN/TERT` with one terrain definition (`terrain/grass.ter`).
- `GEOD` with one `POOL` (uint16, 1 plane, raw encoded, 3 elements) plus
  the matching `SCAL` (1 plane → 2 floats: multiplier `1.0` and offset
  `0.0`). For brevity, the example uses a *placeholder* coordinate pool —
  real DSFs would store at least lat/lon (2 planes); the decoder doesn't
  reject under-dimensioned pools, only the X-Plane consumer does.
- `CMDS` with `SET DEFINITION 8` (defn 0), `COORDINATE POOL SELECT` (pool
  0), then `PATCH TRIANGLE` with three indices (0, 1, 2).
- 16-byte MD5 footer (illustrative bytes).

This example is hand-laid-out for clarity; the implementation's encoder
will recompute the MD5 on write.

```
Offset    Hex                                                ASCII
00000000  58 50 4C 4E 45 44 53 46  01 00 00 00 44 41 45 48  XPLNEDSF....DAEH
00000010  44 00 00 00 50 4F 52 50  3C 00 00 00 73 69 6D 2F  D...PORP<...sim/
00000020  77 65 73 74 00 30 00 73  69 6D 2F 65 61 73 74 00  west.0.sim/east.
00000030  31 00 73 69 6D 2F 73 6F  75 74 68 00 30 00 73 69  1.sim/south.0.si
00000040  6D 2F 6E 6F 72 74 68 00  31 00 4E 46 45 44 1F 00  m/north.1.NFED..
00000050  00 00 54 52 45 54 17 00  00 00 74 65 72 72 61 69  ..TRET....terrai
00000060  6E 2F 67 72 61 73 73 2E  74 65 72 00 44 4F 45 47  n/grass.ter.DOEG
00000070  18 00 00 00 4C 4F 4F 50  10 00 00 00 03 00 00 00  ....LOOP........
00000080  01 00 00 00 00 01 00 02  00 53 44 4D 43 0E 00 00  .........SDMC...
00000090  00 03 00 01 00 17 03 00  00 01 00 02 00 00 00 00  ................
000000a0  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  ................
000000b0  00 00                                              ..
```

Total length: **0xB2 = 178 bytes**.

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0x00–0x0B | FileHeader | cookie + version 1 | Same as minimal. |
| 0x0C–0x4F | HEAD/PROP | 4 sim properties | Same as minimal. |
| 0x50–0x53 | DEFN.ID | `'DEFN'` | Top-level container. |
| 0x54–0x57 | DEFN.Size | `0x0000001F` (31) | 8-byte header + 23-byte `TERT`. |
| 0x58–0x5B | TERT.ID | `'TERT'` | Terrain definition table. |
| 0x5C–0x5F | TERT.Size | `0x00000017` (23) | 8-byte header + 15-byte payload. |
| 0x60–0x6B | TERT payload | `terrain/grass.ter\0` | 12 ASCII bytes + NUL = 13. (Note: writer rounds to 15 here for alignment-free padding; spec doesn't require padding.) |
| 0x6C–0x6F | GEOD.ID | `'GEOD'` | Coordinate pools container. |
| 0x70–0x73 | GEOD.Size | `0x00000018` (24) | 8-byte header + 16-byte `POOL`. (No `SCAL` in this minimal pool example for brevity — see ambiguity below.) |
| 0x74–0x77 | POOL.ID | `'POOL'` | 16-bit point pool. |
| 0x78–0x7B | POOL.Size | `0x00000010` (16) | 8-byte header + 8-byte body. |
| 0x7C–0x7F | POOL.ItemCount | `0x00000003` | 3 elements per plane. |
| 0x80 | POOL.PlaneCount | `0x01` | 1 plane. |
| 0x81 | plane 0 encoding | `0x00` | RAW. |
| 0x82–0x87 | plane 0 data | `00 00 01 00 02 00` | 3 × `uint16` little-endian: 0, 1, 2. |
| 0x88–0x8B | CMDS.ID | `'CMDS'` | Commands container. |
| 0x8C–0x8F | CMDS.Size | `0x0000000E` (14) | 8-byte header + 6-byte payload. |
| 0x90 | opcode 3 | SET DEFINITION 8 | Following byte is `uint8` defn index. |
| 0x91 | defn idx | `0x00` | Use TERT[0] = `terrain/grass.ter`. |
| 0x92 | opcode 1 | COORDINATE POOL SELECT | Following two bytes are `uint16` pool index. |
| 0x93–0x94 | pool idx | `0x0000` | Use POOL[0]. |
| 0x95 | opcode 23 | PATCH TRIANGLE | Following byte is `uint8 N`. |
| 0x96 | N | `0x03` | 3 coordinate indices follow. |
| 0x97–0x9C | indices | `00 00 01 00 02 00` | 3 × `uint16` little-endian: 0, 1, 2. |
| 0xA2–0xB1 | Footer.MD5 | illustrative | 16 zero bytes here as a placeholder; a real writer fills in the MD5 of the preceding `0xA2` bytes. |

## Notes

> **Ambiguity:** This typical example deliberately omits a `SCAL` atom to
> keep the byte count small. A real DSF must include one `SCAL` per `POOL`
> with `2 × PlaneCount` `float32` entries; an X-Plane reader will refuse a
> `POOL` without a matching `SCAL`. For round-trip testing the decoder
> should still accept the structurally-valid layout above and re-emit it
> identically; the X-Plane-reader-validity check is a layer above the
> decoder.
