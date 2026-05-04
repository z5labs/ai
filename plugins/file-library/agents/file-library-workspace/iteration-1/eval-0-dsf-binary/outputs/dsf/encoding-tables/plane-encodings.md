# Plane encodings

The 1-byte tag that prefixes each data plane inside a
[`POOL`](../structures/pool-atom.md) or [`PO32`](../structures/po32-atom.md)
planar atom. Used in the encoding/decoding pseudocode in
[`../structures/planar-numeric-atom.md`](../structures/planar-numeric-atom.md).

| Value | Name           | Description |
|---|---|---|
| 0   | RAW              | `ItemCount` elements written sequentially, no transformation. |
| 1   | DIFFERENCED      | Each element on the wire is `value − previous`; first element is the absolute value. Reader runs prefix sum, wrapping modulo the element width. |
| 2   | RLE              | Run-length-encoded elements (see scheme below). Values are absolute. |
| 3   | RLE_DIFFERENCED  | RLE wrapper around a differenced stream — RLE-decode first, then prefix-sum. |

## RLE byte layout

Each RLE byte is a `uint8`:

- High bit set (`0x80`): the next 1 element repeats `count & 0x7F` times
  (so a single on-wire element produces 1..127 reconstructed elements).
- High bit clear: the next `count & 0x7F` elements are written individually.

The element width is determined by the parent atom (`POOL` ⇒ 2 bytes,
`PO32` ⇒ 4 bytes). The count `0` is a no-op (zero elements emitted) and
should be avoided by encoders but accepted by decoders.

## Notes

- `SCAL` and `SC32` payloads do **not** use this enum — their bytes are a
  flat `[]float32` with no plane framing.
