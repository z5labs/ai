# ObjtAtom

`OBJT` (sub-atom of `DEFN`). [String table](string-table-atom.md) listing
external `.obj` object definition file paths.

The order of strings assigns the zero-based object index used by
`SET DEFINITION 8/16/32` (opcodes 3/4/5) when followed by `OBJECT` /
`OBJECT RANGE` commands.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `StringTableAtom` | Paths | Slash-separated relative paths with file extensions. |
