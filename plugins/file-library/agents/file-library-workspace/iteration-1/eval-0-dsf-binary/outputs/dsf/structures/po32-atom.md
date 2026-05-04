# Po32Atom

`PO32` (sub-atom of `GEOD`). A 32-bit coordinate pool: a
[planar numeric atom](planar-numeric-atom.md) where every element is a
`uint32`. Used by vector network commands so a single pool can carry more
than 65,535 points.

Indices into a `PO32` are `uint32` (0..2^32-1).

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `PlanarNumeric[uint32]` | Body | Planar `uint32` payload — see [`planar-numeric-atom.md`](planar-numeric-atom.md). |

## Nested structures

Paired by position with the matching [`SC32`](sc32-atom.md) atom.
