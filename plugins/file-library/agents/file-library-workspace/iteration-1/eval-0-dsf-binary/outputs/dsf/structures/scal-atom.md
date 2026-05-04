# ScalAtom

`SCAL` (sub-atom of `GEOD`). The scaling/offset pairs that turn a
[`POOL`](pool-atom.md)'s raw `uint16` plane values into real coordinates.

## Encoding

The payload is a flat array of `float32` values: 2 floats per plane in the
matching `POOL` (multiplier first, then offset). For a `POOL` with N planes,
the matching `SCAL` carries `2N` floats, i.e. `8N` bytes.

```
+----------+----------+ +----------+----------+ ... +----------+----------+
| Mult[0]  | Off[0]   | | Mult[1]  | Off[1]   |     | Mult[N-1]| Off[N-1] |
| float32  | float32  | | float32  | float32  |     | float32  | float32  |
+----------+----------+ +----------+----------+ ... +----------+----------+
```

For each plane `p`, the real value of element `i` is:

```
real(p, i) = float64(POOL[p, i]) * Mult[p] + Off[p]
```

The conversion is performed in `float64` per the spec.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `[]float32` | Pairs | `2 × matchingPool.PlaneCount` floats — multiplier, offset, multiplier, offset, … |

## Variable-length fields

- **Length determination**: implicit; `parent.Size − 8` bytes / 4 bytes per
  `float32`. Decoder may also validate against `2 × PlaneCount` of the
  matching `POOL`.

## Nested structures

Pairs by position with the matching [`POOL`](pool-atom.md) inside `GEOD`.
