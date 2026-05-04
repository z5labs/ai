# TertAtom

`TERT` (sub-atom of `DEFN`). [String table](string-table-atom.md) listing
external `.ter` terrain definition file paths. `.png` and `.bmp` paths are
also legal as direct terrain textures.

The order of strings assigns the zero-based terrain index used by
`SET DEFINITION 8/16/32` (opcodes 3/4/5) when followed by mesh commands.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `StringTableAtom` | Paths | Slash-separated relative paths with file extensions. |
