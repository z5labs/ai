# DemdAtom

`DEMD` (raster layer raw data; sub-atom of `DEMS`). The atom payload is the
flat raster pixel buffer for the layer described by the immediately
preceding [`DEMI`](demi-atom.md). No internal header — `parent.Size − 8` is
exactly the pixel byte count.

The expected size is `DEMI.Width * DEMI.Height * DEMI.BytesPerPixel` bytes.
Pixel order is implementation-defined ("how the raster data is interpreted
is defined by X-Plane") — for the round-trip pipeline, the bytes are passed
through as opaque storage, decoded shape-aware only when an X-Plane consumer
actually needs to render them.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `[]byte` | Pixels | Width × Height × BytesPerPixel bytes. |
