# DemiAtom

`DEMI` (raster layer info; sub-atom of `DEMS`). A fixed-layout record
describing one raster layer's dimensions, encoding, and post-load scaling.

## Byte diagram

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|  Version (u8) | BPP (u8)      |          Flags (u16)          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          Width (u32)                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          Height (u32)                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          Scale (f32)                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          Offset (f32)                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

The total payload is 20 bytes (`uint8 + uint8 + uint16 + 4*uint32 = 4 +
4 + 4 + 4 + 4 = 20`). Combined with the 8-byte atom header, a `DEMI` atom
on the wire is exactly 28 bytes.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | `uint8`   | Version       | Always `1`. |
| 1 | 1 | `uint8`   | BytesPerPixel | Bytes per raster pixel. Must be 1, 2, or 4. |
| 2 | 2 | `uint16`  | Flags         | See [Bit fields](#bit-fields) and [`../encoding-tables/dem-flags.md`](../encoding-tables/dem-flags.md). |
| 4 | 4 | `uint32`  | Width         | Width in pixels (east–west). |
| 8 | 4 | `uint32`  | Height        | Height in pixels (north–south). |
| 12 | 4 | `float32` | Scale         | Multiplier applied to each pixel post-load. |
| 16 | 4 | `float32` | Offset        | Added after scaling. Final value = `pixel * Scale + Offset`. |

## Bit fields

`Flags` is a 16-bit little-endian value with the following bit layout:

| Bit(s) | Name | Description |
|---|---|---|
| 1-0 | NumberType   | 0 = `float32` (then `BytesPerPixel` must be 4); 1 = signed integer (1, 2, or 4 bytes); 2 = unsigned integer (1, 2, or 4 bytes); 3 = reserved. |
| 2   | PostCentric  | If set, edge pixels lie on the DSF's geometric boundary (post-centric). If clear, the pixel rectangles' outer edges lie on the boundary (area-centric). |
| 15-3 | Reserved | Must be 0. |

See [`../encoding-tables/dem-flags.md`](../encoding-tables/dem-flags.md) for the
defined values.

> **Ambiguity:** The spec text says "A flag value of 4 (bit 3) defines the data
> as post-centric." Bit value 4 corresponds to **bit 2** in LSB-0 numbering,
> not bit 3. This reference uses bit 2 (matching `0x0004`); a careful
> implementer should confirm against a known-good fixture before committing.
