# DEM flags (`DEMI.Flags`)

Bit-packed `uint16` describing the number type and centric mode of a raster
layer. See [`../structures/demi-atom.md`](../structures/demi-atom.md).

## Number type (bits 1–0)

| Value | Name      | Description |
|---|---|---|
| 0   | FLOAT       | `float32` per pixel — `BytesPerPixel` must be 4. |
| 1   | INT_SIGNED  | Signed integer per pixel — `BytesPerPixel` ∈ {1, 2, 4}. |
| 2   | INT_UNSIGNED| Unsigned integer per pixel — `BytesPerPixel` ∈ {1, 2, 4}. |
| 3   | (reserved)  | Not assigned. |

## Centric mode (bit 2)

| Bit value | Bit position | Name           | Description |
|---|---|---|---|
| 0x0000  | —      | AREA_CENTRIC      | Outer edge of the pixel rectangle aligns with the DSF boundary. |
| 0x0004  | bit 2  | POST_CENTRIC      | Edge pixel value lies exactly on the DSF boundary. |

## Reserved

Bits 15–3 (other than the centric-mode bit 2) are reserved and must be 0.

## Notes

> **Ambiguity:** The spec text says "A flag value of 4 (bit 3) defines the
> data as post-centric"; in LSB-0 numbering, the value 4 is **bit 2**, not
> bit 3. This table follows the value (`0x0004`), which is what a writer
> can produce unambiguously.
