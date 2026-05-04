# DefnAtom

Top-level `DEFN` atom. An [atom-of-atoms](atom-of-atoms.md) carrying the
definition tables: `TERT`, `OBJT`, `POLY`, `NETW`, and (X-Plane 10) `DEMN`.

Each sub-atom is a [string table](string-table-atom.md) of forward-slash-
delimited file paths (extensions included). The order of strings in each
table assigns a zero-based index used by the `CMDS` opcodes (see
[`set-definition-8`](../encoding-tables/command-opcodes.md)). A single
definition table is capped at 65,536 entries.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `AtomOfAtoms` | SubAtoms | Some subset of `TERT`, `OBJT`, `POLY`, `NETW`, `DEMN`. Each absent table is treated as empty. |

## Nested structures

- [`TERT`](tert-atom.md) — terrain definition file paths (`.ter` / `.png` / `.bmp`).
- [`OBJT`](objt-atom.md) — object definition file paths (`.obj`).
- [`POLY`](poly-atom.md) — polygon definition file paths (`.fac`/`.for`/etc.).
- [`NETW`](netw-atom.md) — network definition file paths (`.net`).
- [`DEMN`](demn-atom.md) — raster layer names (X-Plane 10).
