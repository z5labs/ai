# DemsAtom

Top-level `DEMS` atom (X-Plane 10 only). An
[atom-of-atoms](atom-of-atoms.md) carrying raster (DEM) layers — one
`(DEMI, DEMD)` pair per layer, in the same order as the names listed in
`DEMN`.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `AtomOfAtoms` | SubAtoms | Alternating `DEMI` (info) and `DEMD` (raw bytes), one pair per raster layer. |

## Nested structures

- [`DEMI`](demi-atom.md) — fixed-layout 20-byte record describing one layer.
- [`DEMD`](demd-atom.md) — raw raster bytes for that layer; size matches the
  preceding `DEMI`.

## Ambiguities

> **Ambiguity:** The spec describes the children as "one DEMI and one DEMD
> for each raster layer" but does not explicitly mandate `DEMI` first. The
> expected layout is `DEMI, DEMD, DEMI, DEMD, …` — a writer that produces
> any other interleaving will confuse X-Plane.
