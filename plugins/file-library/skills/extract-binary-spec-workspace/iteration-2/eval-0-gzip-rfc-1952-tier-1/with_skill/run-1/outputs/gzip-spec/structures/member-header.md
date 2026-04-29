# MemberHeader

Fixed 10-byte identification block at the start of every gzip member. Carries
the magic bytes that identify the stream as gzip, the compression method
selector, the flag byte controlling optional fields, the original
modification time, extra deflate-specific flags, and the source OS.

## Byte diagram

```
 0   1   2   3   4   5   6   7   8   9
+---+---+---+---+---+---+---+---+---+---+
|ID1|ID2|CM |FLG|     MTIME     |XFL|OS |
+---+---+---+---+---+---+---+---+---+---+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | uint8 | ID1 | Fixed magic byte 1: must be `0x1F` (31). |
| 1 | 1 | uint8 | ID2 | Fixed magic byte 2: must be `0x8B` (139). |
| 2 | 1 | uint8 | CM | Compression method. See [`../encoding-tables/compression-methods.md`](../encoding-tables/compression-methods.md). |
| 3 | 1 | uint8 | FLG | Flag bits selecting optional fields. See [`flg.md`](flg.md). |
| 4 | 4 | uint32 | MTIME | Modification time of original input, Unix seconds since 1970-01-01 00:00:00 UTC. `0` means no timestamp available. Stored little-endian. |
| 8 | 1 | uint8 | XFL | Extra flags specific to the compression method. See [`../encoding-tables/xfl-values.md`](../encoding-tables/xfl-values.md). |
| 9 | 1 | uint8 | OS | Source filesystem / OS family. See [`../encoding-tables/os-values.md`](../encoding-tables/os-values.md). |

## Conditional / optional fields

The header itself is fully fixed: every byte is always present. The optional
blocks (`ExtraField`, `FName`, `FComment`, `FHCRC`) follow the header and are
governed by `FLG` — see [`member.md`](member.md) for the per-flag conditions.

## Compliance rules

From RFC 1952 §2.3.1.2:
- A compliant **encoder** must produce correct `ID1`, `ID2`, `CM`, `CRC32`,
  and `ISIZE`. It may set `MTIME = 0`, `XFL = 0`, and `OS = 255`
  ("unknown") when no better value is available. All reserved bits in `FLG`
  must be zero.
- A compliant **decoder** must verify `ID1`, `ID2`, and `CM`, and must error
  on any non-zero reserved `FLG` bit. It must read enough of the optional
  blocks to skip over them, but is not required to interpret `MTIME`, `XFL`,
  `OS`, or `FTEXT`.

## Ambiguities

> **Ambiguity:** RFC 1952 specifies that `MTIME = 0` means "no timestamp
> available". It does not address how a decoder should distinguish that
> sentinel from a legitimate timestamp at the Unix epoch (1970-01-01
> 00:00:00 UTC), since both encode as `0x00000000`. Implementations
> conventionally treat `0` as "unknown".
