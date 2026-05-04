# Atom

The generic atom envelope. Every chunk in the atom region of a DSF file —
top-level (`HEAD`, `DEFN`, `GEOD`, `DEMS`, `CMDS`) and nested (`PROP`, `TERT`,
`POOL`, etc.) — uses this framing.

## Byte diagram

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                          ID (uint32)                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Size (uint32)                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       Payload (Size - 8 bytes)                |
~                              ...                              ~
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | `uint32` | ID | 32-bit ID, little-endian. See [`../encoding-tables/atom-ids.md`](../encoding-tables/atom-ids.md). |
| 4 | 4 | `uint32` | Size | Total atom length in bytes, including these 8 header bytes. |
| 8 | Size − 8 | `[]byte` | Payload | Format depends on `ID`. |

## Variable-length fields

- **Length determination**: length-prefixed (`Size` is the prefix).
- **Length prefix format**: `uint32`, **inclusive** of the 8-byte header.
- **Maximum length**: bounded by `uint32` (~4 GB).
- **Encoding**: opaque bytes; payload format is selected by `ID`.

## Nested structures

- For container atoms (`HEAD`, `DEFN`, `GEOD`, `DEMS`), the payload is a
  concatenation of sub-atoms — see [`atom-of-atoms.md`](atom-of-atoms.md).
- For `PROP`, `TERT`, `OBJT`, `POLY`, `NETW`, `DEMN`, the payload is a
  packed string table — see [`string-table-atom.md`](string-table-atom.md).
- For `POOL`, `PO32`, `SCAL`, `SC32`, the payload is a planar numeric block —
  see [`planar-numeric-atom.md`](planar-numeric-atom.md).
- For `CMDS`, the payload is a sequence of variable-length commands — see
  [`cmds-atom.md`](cmds-atom.md).
- For `DEMI`, the payload is a single fixed-layout record — see
  [`demi-atom.md`](demi-atom.md).
- For `DEMD`, the payload is raw raster bytes — see
  [`demd-atom.md`](demd-atom.md).

## Ambiguities

> **Ambiguity:** The spec says `Size` includes the 8-byte header. Some readers
> in the wild have been observed to interpret `Size` as the payload-only count
> (a known interoperability gotcha for hand-written tools). This reference
> follows the spec text: `Size` is the **total** atom length including the
> 8-byte header.

> **Ambiguity:** Atom IDs less than 8 (i.e., a `Size` < 8) are not addressed
> by the spec. A defensive decoder should reject `Size` < 8 as malformed.
