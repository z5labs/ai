# Minimal valid DSF

The smallest DSF a writer can produce while still satisfying the required
`sim/west`/`sim/east`/`sim/south`/`sim/north` properties: a 12-byte file
header, a `HEAD` containing one `PROP` (the four required edge properties
with placeholder values), then minimal-but-empty `DEFN`, `GEOD`, and `CMDS`
container atoms, and finally the 16-byte MD5 footer.

The atom region's MD5 covers everything from offset `0` up to the start of
the footer. The MD5 bytes shown are illustrative — a real writer must
recompute on save.

```
Offset    Hex                                                ASCII
00000000  58 50 4C 4E 45 44 53 46  01 00 00 00 44 41 45 48  XPLNEDSF....DAEH
00000010  44 00 00 00 50 4F 52 50  3C 00 00 00 73 69 6D 2F  D...PORP<...sim/
00000020  77 65 73 74 00 30 00 73  69 6D 2F 65 61 73 74 00  west.0.sim/east.
00000030  31 00 73 69 6D 2F 73 6F  75 74 68 00 30 00 73 69  1.sim/south.0.si
00000040  6D 2F 6E 6F 72 74 68 00  31 00 4E 46 45 44 08 00  m/north.1.NFED..
00000050  00 00 44 4F 45 47 08 00  00 00 53 44 4D 43 08 00  ..DOEG....SDMC..
00000060  00 00 d4 1d 8c d9 8f 00  b2 04 e9 80 09 98 ec f8  ................
00000070  42 7e                                              B~
```

Total length: **0x72 = 114 bytes**.

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0x00–0x07 | FileHeader.Cookie | `XPLNEDSF` | ASCII magic. |
| 0x08–0x0B | FileHeader.Version | `0x00000001` | Master version 1. |
| 0x0C–0x0F | HEAD atom ID | `'HEAD'` (`44 41 45 48`) | Top-level atom-of-atoms. |
| 0x10–0x13 | HEAD atom Size | `0x00000044` (68) | Includes the 8-byte header and the `PROP` sub-atom. |
| 0x14–0x17 | PROP atom ID | `'PROP'` (`50 4F 52 50`) | String table. |
| 0x18–0x1B | PROP atom Size | `0x0000003C` (60) | 8-byte header + 52-byte string-table payload. |
| 0x1C–0x4F | PROP payload | 8 NUL-terminated strings | `sim/west\0` `0\0` `sim/east\0` `1\0` `sim/south\0` `0\0` `sim/north\0` `1\0`. |
| 0x50–0x53 | DEFN atom ID | `'DEFN'` (`4E 46 45 44`) | Empty container — no `TERT`/`OBJT`/`POLY`/`NETW`/`DEMN`. |
| 0x54–0x57 | DEFN atom Size | `0x00000008` | 8-byte header, empty payload. |
| 0x58–0x5B | GEOD atom ID | `'GEOD'` (`44 4F 45 47`) | Empty container — no `POOL`/`SCAL`/`PO32`/`SC32`. |
| 0x5C–0x5F | GEOD atom Size | `0x00000008` | 8-byte header, empty payload. |
| 0x60–0x63 | CMDS atom ID | `'CMDS'` (`53 44 4D 43`) | No commands. |
| 0x64–0x67 | CMDS atom Size | `0x00000008` | 8-byte header, empty payload. |
| 0x68–0x71 | Footer.MD5 (lo 10 bytes shown) | `D4 1D 8C D9 …` | Illustrative — recompute on save. The shown 10 bytes are the start of the MD5 of the empty string (just to indicate "real MD5 bytes go here"). |

## Notes

The 4-byte `Size` of every empty atom is the minimum value `8` (8-byte
header, zero-byte payload). The decoder must accept that as a non-error
case so an empty `DEFN`/`GEOD`/`CMDS` is round-trippable.
