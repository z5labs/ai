# Atom IDs

The known atom IDs in DSF. Each ID is read on the wire as a little-endian
`uint32`; the ASCII spelling is the value the spec uses to refer to the
atom in prose. The `Hex (LE)` column shows what the four bytes look like
in a byte-level hex view of the file.

| Value (ASCII) | Hex (LE)         | Name   | Container | Description |
|---|---|---|---|---|
| `'HEAD'` | `44 41 45 48`   | HEAD   | top-level | Atom-of-atoms; carries `PROP`. |
| `'PROP'` | `50 4F 52 50`   | PROP   | inside `HEAD` | String table of name/value property pairs. |
| `'DEFN'` | `4E 46 45 44`   | DEFN   | top-level | Atom-of-atoms; carries `TERT`/`OBJT`/`POLY`/`NETW`/`DEMN`. |
| `'TERT'` | `54 52 45 54`   | TERT   | inside `DEFN` | String table of `.ter`/`.png`/`.bmp` terrain definition paths. |
| `'OBJT'` | `54 4A 42 4F`   | OBJT   | inside `DEFN` | String table of `.obj` object definition paths. |
| `'POLY'` | `59 4C 4F 50`   | POLY   | inside `DEFN` | String table of polygon definition paths. |
| `'NETW'` | `57 54 45 4E`   | NETW   | inside `DEFN` | String table of `.net` network definition paths. |
| `'DEMN'` | `4E 4D 45 44`   | DEMN   | inside `DEFN` | String table of raster layer names (X-Plane 10 only). |
| `'GEOD'` | `44 4F 45 47`   | GEOD   | top-level | Atom-of-atoms; carries `POOL`/`SCAL`/`PO32`/`SC32`. |
| `'POOL'` | `4C 4F 4F 50`   | POOL   | inside `GEOD` | 16-bit point pool (planar `uint16`). |
| `'SCAL'` | `4C 41 43 53`   | SCAL   | inside `GEOD` | Scaling/offset pairs (`float32`) for the matching `POOL`. |
| `'PO32'` | `32 33 4F 50`   | PO32   | inside `GEOD` | 32-bit point pool (planar `uint32`). |
| `'SC32'` | `32 33 43 53`   | SC32   | inside `GEOD` | Scaling/offset pairs for the matching `PO32`. |
| `'DEMS'` | `53 4D 45 44`   | DEMS   | top-level | Atom-of-atoms; carries `(DEMI, DEMD)` pairs (X-Plane 10). |
| `'DEMI'` | `49 4D 45 44`   | DEMI   | inside `DEMS` | Per-layer raster info record (20 bytes). |
| `'DEMD'` | `44 4D 45 44`   | DEMD   | inside `DEMS` | Per-layer raw raster pixel bytes. |
| `'CMDS'` | `53 44 4D 43`   | CMDS   | top-level | Sequence of variable-length commands. |

## Notes

- IDs consisting entirely of capital ASCII letters and digits are reserved
  for the DSF specification; private extensions should use mixed-case or
  punctuation-bearing IDs.
- The byte order in the `Hex (LE)` column is what a hex-dump tool shows
  for a little-endian `uint32` — i.e., the ASCII bytes appear *reversed*
  relative to the human-readable name. The spec calls this out explicitly:
  "viewing a DSF file in a windows hex editor the GEOD atom would read
  DOEG since 'GEOD' is really a 32-bit character constant".
