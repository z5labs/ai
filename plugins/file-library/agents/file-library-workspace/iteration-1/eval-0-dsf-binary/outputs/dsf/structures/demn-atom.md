# DemnAtom

`DEMN` (sub-atom of `DEFN`; **X-Plane 10 only**).
[String table](string-table-atom.md) listing the names of each raster layer
encoded inside the file's `DEMS` atom.

The order of names matches the order of `(DEMI, DEMD)` pairs inside `DEMS` —
the first name corresponds to the first `DEMI/DEMD` pair, etc.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | parent.Size − 8 | `StringTableAtom` | Names | Layer names; how each layer is interpreted is decided by X-Plane. |
