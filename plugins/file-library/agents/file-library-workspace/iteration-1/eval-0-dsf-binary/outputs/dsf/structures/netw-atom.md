# NetwAtom

`NETW` (sub-atom of `DEFN`). [String table](string-table-atom.md) listing
external `.net` network definition file paths.

X-Plane 8/9 only accept one network definition per DSF file; multiple road
types are encoded via vector subtypes (set by opcode 6 — `SET ROAD SUBTYPE 8`).
X-Plane 10 lifts this to multiple `.net` definitions.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `StringTableAtom` | Paths | Slash-separated relative paths with file extensions. |
