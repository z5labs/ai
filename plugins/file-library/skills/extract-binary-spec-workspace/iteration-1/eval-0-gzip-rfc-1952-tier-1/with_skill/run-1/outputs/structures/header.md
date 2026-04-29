# Header

The fixed 10-byte prefix that begins every gzip member. Identifies the
file as gzip, declares the compression method, gates the optional fields
that follow, and records the original modification time, compressor
hints, and source operating system.

## Byte diagram

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     ID1       |     ID2       |      CM       |      FLG      |
+---------------+---------------+---------------+---------------+
|                            MTIME                              |
+---------------------------------------------------------------+
|      XFL      |      OS       |
+---------------+---------------+
```

`ID1`, `ID2`, `CM`, `FLG`, `XFL`, and `OS` are each one byte. `MTIME` is
4 bytes, **little-endian**. Total: 10 bytes.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | uint8 | ID1 | Magic byte 1 — must be `0x1f` |
| 1 | 1 | uint8 | ID2 | Magic byte 2 — must be `0x8b` |
| 2 | 1 | uint8 | CM | Compression method — see [`../encoding-tables/compression-methods.md`](../encoding-tables/compression-methods.md). In practice always `8` (deflate) |
| 3 | 1 | uint8 | FLG | Flag bits gating optional fields — see [`flg.md`](flg.md) |
| 4 | 4 | uint32 | MTIME | Modification time of the original file, Unix seconds since 1970-01-01 UTC; `0` means "no time stamp" |
| 8 | 1 | uint8 | XFL | Extra flags (compressor hints) — see [`../encoding-tables/xfl-deflate.md`](../encoding-tables/xfl-deflate.md) when CM=8 |
| 9 | 1 | uint8 | OS | Source operating system — see [`../encoding-tables/operating-system.md`](../encoding-tables/operating-system.md) |

`MTIME` is a `uint32` in **little-endian** byte order (matches the
format-wide convention).

## Bit fields

Only `FLG` is bit-packed. See the dedicated [`flg.md`](flg.md) for the
per-bit table.

## Versioning notes

- **Version field location**: gzip has no explicit version field in the
  member. The combination `(ID1, ID2)` identifies the format; `CM`
  identifies the compression method (and implicitly the
  RFC 1951 version of deflate).
- **Backward compatibility**: a decoder that does not recognize `CM`
  MUST treat the member as undecodable. A decoder that sees a non-zero
  reserved bit in `FLG` MUST reject the member (see
  [`flg.md`](flg.md)).

## Ambiguities

> **Ambiguity:** The RFC does not state how a decoder should treat
> `MTIME = 0`. By convention this value means "no time stamp available"
> and is preserved as-is rather than being replaced by the current time.

> **Ambiguity:** `XFL` values other than 0, 2, and 4 are unspecified for
> CM=8. Implementations vary on whether to preserve the value or
> normalize to 0. Most decoders ignore `XFL` entirely.
