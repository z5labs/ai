# PoolAtom

`POOL` (sub-atom of `GEOD`). A 16-bit coordinate pool: a
[planar numeric atom](planar-numeric-atom.md) where every element is a
`uint16`. The `ItemCount` is the number of N-tuples; the `PlaneCount` is N.

Indices into a `POOL` are `uint16` (0..65,535). To reconstruct the real
coordinate, each plane's value is multiplied by the corresponding `SCAL`
multiplier and added to the `SCAL` offset (in `float64` precision).

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `PlanarNumeric[uint16]` | Body | Planar `uint16` payload — see [`planar-numeric-atom.md`](planar-numeric-atom.md). |

## Nested structures

Paired by position with the matching `SCAL` atom inside the same `GEOD` —
the N-th `POOL` is scaled by the N-th [`SCAL`](scal-atom.md).
