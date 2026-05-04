# Complex DSF: raster layer + vector network

A DSF that stresses the X-Plane 10 raster path and the vector-network
command path simultaneously:

- `HEAD/PROP` with the four required edge properties.
- `DEFN` with one `NETW` (one network definition) and one `DEMN` (one
  raster layer name).
- `GEOD` with one `PO32` (32-bit pool, 4 planes for vector: lon/lat/elev/
  junctionID) and one `SC32` (4 planes ⇒ 8 floats: 4 multipliers and 4
  offsets).
- `DEMS` with one `(DEMI, DEMD)` pair: a 2×2 unsigned 8-bit raster.
- `CMDS` with `SET DEFINITION 8` (defn 0), then `NETWORK CHAIN 32` over a
  pair of 32-bit indices.
- 16-byte MD5 footer (illustrative).

This example does not include a fully byte-accurate hex dump — the file is
about 280 bytes and a real implementer round-trips it via the encoder
rather than hand-crafting it. The structural decomposition below is what
the spec requires.

## Atom layout (sizes are spec-driven; no full hex dump)

```
+ FileHeader (12)
+ HEAD (atom-of-atoms)
  + PROP (string table, 4 required properties)
+ DEFN (atom-of-atoms)
  + NETW (string table, 1 path: "network/highway.net\0")
  + DEMN (string table, 1 name: "elevation\0")
+ GEOD (atom-of-atoms)
  + PO32 (planar uint32: ItemCount=2, PlaneCount=4, four RAW planes,
          each plane carries 2 uint32 elements)
  + SC32 (8 × float32: 4 multipliers and 4 offsets)
+ DEMS (atom-of-atoms)
  + DEMI (Version=1, BPP=1, Flags=2 (INT_UNSIGNED) | 4 (POST_CENTRIC)=6,
          Width=2, Height=2, Scale=1.0, Offset=0.0)
  + DEMD (4 raw bytes: 0x00 0x10 0x20 0x30)
+ CMDS
  + opcode 3 (SET DEFINITION 8) defn=0
  + opcode 1 (COORDINATE POOL SELECT) pool=0  (selects PO32[0])
  + opcode 11 (NETWORK CHAIN 32) N=2 [0x00000001, 0x00000002]
+ Footer (MD5, 16 bytes)
```

## Annotation (selected fields)

| Field | Value | Notes |
|---|---|---|
| `DEMI.Version`        | `0x01`               | Always 1. |
| `DEMI.BytesPerPixel`  | `0x01`               | 1 byte per pixel. |
| `DEMI.Flags`          | `0x0006` (LE: `06 00`) | INT_UNSIGNED (bits 1–0 = 2) plus POST_CENTRIC (bit 2 = 4). |
| `DEMI.Width`          | `0x00000002` (LE)    | 2-pixel-wide raster. |
| `DEMI.Height`         | `0x00000002` (LE)    | 2-pixel-tall raster. |
| `DEMI.Scale`          | `0x3F800000`         | `float32(1.0)` little-endian. |
| `DEMI.Offset`         | `0x00000000`         | `float32(0.0)`. |
| `DEMD` payload size   | 4 bytes              | Width × Height × BytesPerPixel = 2 × 2 × 1 = 4. |
| `NETWORK CHAIN 32`    | opcode `0x0B`        | Followed by `uint8 N=2` then 2 × `uint32` little-endian indices `0x00000001`, `0x00000002`. |

## Notes

> **Ambiguity:** The spec is not explicit about the relationship between
> `COORDINATE POOL SELECT` and the family of pool used by the next command
> (`POOL[i]` vs. `PO32[i]`). The reading used here — the same opcode
> selects from whichever pool list the next command's family uses — is the
> X-Plane writer's convention.

> **Ambiguity:** The interpretation of raster pixel byte order (row-major
> top-to-bottom or bottom-to-top) is "defined by X-Plane" per the spec.
> The byte sequence `0x00 0x10 0x20 0x30` shown above is opaque to the
> decoder/encoder; it round-trips losslessly regardless of orientation.
