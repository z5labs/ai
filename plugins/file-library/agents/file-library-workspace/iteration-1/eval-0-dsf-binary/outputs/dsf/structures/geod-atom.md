# GeodAtom

Top-level `GEOD` atom: an [atom-of-atoms](atom-of-atoms.md) carrying the
coordinate pools for the file. Children are sequences of:

- [`POOL`](pool-atom.md) — 16-bit point pool (planar `uint16`).
- [`SCAL`](scal-atom.md) — 16-bit scaling/offset pairs as `float32`. There
  must be exactly one `SCAL` per `POOL`, in the same order.
- [`PO32`](po32-atom.md) — 32-bit point pool (planar `uint32`); used for
  vector commands.
- [`SC32`](sc32-atom.md) — 32-bit scaling/offset pairs. Exactly one per
  `PO32`, in the same order.

There must be an equal count of `POOL` and `SCAL` atoms inside `GEOD`, and
likewise an equal count of `PO32` and `SC32`. Pairing is positional: the
N-th `POOL` is scaled by the N-th `SCAL`; the N-th `PO32` by the N-th `SC32`.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `AtomOfAtoms` | SubAtoms | Some interleaving of `POOL`/`SCAL`/`PO32`/`SC32`. |

## Ambiguities

> **Ambiguity:** The spec does not require `POOL`/`SCAL` (or `PO32`/`SC32`)
> to be physically interleaved on the wire — only that their **order**
> matches. A defensive decoder should pair them by position after collecting
> all sub-atoms, not by adjacency.
