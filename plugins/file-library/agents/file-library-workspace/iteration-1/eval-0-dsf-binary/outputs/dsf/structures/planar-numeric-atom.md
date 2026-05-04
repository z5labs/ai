# PlanarNumericAtom

A planar (struct-of-arrays) numeric payload used by `POOL` (`uint16`),
`PO32` (`uint32`), `SCAL` (`float32`), and `SC32` (`float32`).

The "kind" of numeric value (its size, integer/float, signedness) is **not**
stored in the wire bytes — the parent atom's `ID` selects it:

- `POOL` → `uint16` per element
- `PO32` → `uint32` per element
- `SCAL` → `float32` per element (and uses raw encoding only — see below)
- `SC32` → `float32` per element (raw encoding only)

## Encoding

For integer planar atoms (`POOL`, `PO32`):

```
+----------+-----------+--------------------------+--------------------------+ ...
|  uint32  |   uint8   | uint8 enc | data plane 0 | uint8 enc | data plane 1 | ...
| ItemCount| PlaneCount|    (0..3)                |    (0..3)                |
+----------+-----------+--------------------------+--------------------------+ ...
```

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | `uint32` | ItemCount | Number of N-tuples (one per plane) per data plane. |
| 4 | 1 | `uint8`  | PlaneCount | Number of data planes. |
| 5 | variable | per-plane | Planes | `PlaneCount` planes, each preceded by a 1-byte encoding tag. See [`../encoding-tables/plane-encodings.md`](../encoding-tables/plane-encodings.md). |

Each plane's bytes are framed as:

```
+-----------+----------------------+
|  uint8    |  encoded data        |
| Encoding  |  (size depends on    |
|           |   Encoding + element |
|           |   width)             |
+-----------+----------------------+
```

Encoding tags (see [`../encoding-tables/plane-encodings.md`](../encoding-tables/plane-encodings.md)):

- `0` raw — `ItemCount` elements written sequentially, no transformation.
- `1` differenced — each element on the wire is the value minus the
  previous element (first element is the absolute value); reader reconstructs
  by running prefix sum. Wrap modulo the integer width (`uint16` wraps at
  2^16, `uint32` at 2^32).
- `2` RLE — bytes follow the run-length scheme below; values are absolute.
- `3` RLE + differenced — RLE compresses a stream where each value has been
  replaced by `value − previous`. Reverse: RLE-decode, then prefix-sum.

### RLE byte layout (encodings 2 and 3)

The RLE byte is a `uint8` count where the **high bit** is the run/literal flag:

- High bit set (`0x80`): the next 1 element is repeated `count & 0x7F` times
  (so 1 element on the wire produces 1..127 reconstructed elements).
- High bit clear: the next `count & 0x7F` elements are written individually.

Element bytes are written in the parent atom's element width (2 bytes per
element for `POOL`, 4 bytes per element for `PO32`).

For `SCAL` and `SC32` (float planes), the `ItemCount`/`PlaneCount` framing is
**different** — see those structures' files.

> **Ambiguity:** The spec defines RLE only at the byte level ("the run-length
> byte for run-length encoding is an unsigned 8-bit character"). When the
> element width is 2 or 4 bytes, the run/count applies to *elements*, not
> *bytes* — so a `0x83` run byte in a `POOL` (uint16) plane means "the next
> uint16 (2 bytes) repeats 3 times" (6 reconstructed bytes from 2). This
> reading matches X-Plane's behaviour and the structure of differenced
> integer streams.

> **Ambiguity:** A literal byte of `0x00` and a run byte of `0x80` both
> describe a zero-length sequence — the spec does not say which is canonical
> on encode. Either form decodes to no elements; encoders should avoid
> emitting either (just stop the stream when the plane's `ItemCount` is
> reached) and decoders should accept either by treating the count `0` as
> a no-op and continuing.

## Field table

The shape above (ItemCount/PlaneCount + planes) is the canonical form. The
field table for downstream consumers is best expressed in this layered way;
the implementer's `PlanarNumeric` Go type can carry the framing fields and a
slice of planes.

## Variable-length fields

- **Length determination**: implicit — each plane consumes
  `ItemCount * elementWidth` reconstructed bytes. The encoded byte count
  varies with the encoding.
- **Encoding**: see above per encoding tag.
