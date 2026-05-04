# Sc32Atom

`SC32` (sub-atom of `GEOD`). Scaling/offset pairs for the matching
[`PO32`](po32-atom.md). Same encoding as [`SCAL`](scal-atom.md) — a flat
`float32` array of `2 × PlaneCount` values (multiplier, offset, …) — but
applied to the `PO32` element values rather than `POOL` element values.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `[]float32` | Pairs | `2 × matchingPo32.PlaneCount` floats — multiplier, offset, multiplier, offset, … |

## Nested structures

Pairs by position with the matching [`PO32`](po32-atom.md).
