# PolyAtom

`POLY` (sub-atom of `DEFN`). [String table](string-table-atom.md) listing
external polygon definition file paths. Polygon definitions can be facades
(`.fac`) or forests (`.for`); X-Plane infers the type from the extension.

The order of strings assigns the zero-based polygon index used by
`SET DEFINITION 8/16/32` (opcodes 3/4/5) when followed by `POLYGON`,
`POLYGON RANGE`, `NESTED POLYGON`, or `NESTED POLYGON RANGE` commands.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `StringTableAtom` | Paths | Slash-separated relative paths with file extensions. |
