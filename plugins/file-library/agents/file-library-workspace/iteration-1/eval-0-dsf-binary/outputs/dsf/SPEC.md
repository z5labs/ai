# DSF Binary Specification Reference

## Overview

The Distribution Scenery Format (DSF) is the X-Plane binary file format for
geospatial scenery: a single DSF describes the appearance and physical
properties of a 1×1 degree section of a planet. DSF is used by X-Plane 8, 9,
and 10. This reference covers DSF master file format **version 1**, the only
version defined.

Source: X-Plane Developer documentation,
<https://developer.x-plane.com/article/dsf-file-format-specification/>
(last updated 2019-10-03; revision history goes back to 2004).

## Conventions

- **Byte order**: little-endian for all multi-byte integer and floating-point
  fields. Floating-point follows IEEE 754 (32-bit `float32` or 64-bit
  `float64`).
- **Atom IDs**: each atom ID is a 32-bit unsigned integer. Although the spec
  often spells them as four-character ASCII codes (`HEAD`, `PROP`, `GEOD`),
  they are read as little-endian 32-bit integers, so an atom ID written as
  `'GEOD'` appears in a byte-level hex view as `D O E G` (`0x44 0x4F 0x45 0x47`).
- **Bit numbering**: LSB-0 (bit 0 is the least significant bit). The spec's
  flag tables use bit *value* (e.g. "1" = bit 0, "2" = bit 1, "4" = bit 2).
- **Size units**: bytes (8-bit octets).
- **Notation**: structure files use `Offset (bytes)` for byte offsets and
  `Bit(s)` for bit-packed sub-fields. Field types are Go-friendly
  (`uint8`/`uint16`/`uint32`/`int8`/`int16`/`int32`/`float32`/`float64`,
  `[N]byte`, or a PascalCase reference to another structure file).

> **Ambiguity:** The spec says "DSF is little endian, but differencing is done
> in the machine's endian format … differencing is done logically on the data,
> not on the file". A round-tripping decoder/encoder must logically operate on
> integer values (not byte buffers) when applying differencing, so the on-disk
> bytes are always written in little-endian regardless of host byte order.
> This reference treats the on-disk encoding of differenced data as
> little-endian; an implementer working from the prose alone might come away
> believing the on-disk bytes vary with host endianness.

## Top-level structure

A DSF file is a "chunky" (atomic) container:

1. **File header** — 12 bytes: 8-byte ASCII cookie `XPLNEDSF` + 32-bit
   little-endian master version (currently `1`).
2. **Atom payload** — a sequence of top-level atoms (`HEAD`, `DEFN`, `GEOD`,
   `DEMS` (X-Plane 10 only), `CMDS`). Each atom carries an 8-byte header
   (32-bit ID + 32-bit total size including the header). Inner atoms-of-atoms
   nest the same way. The atom region ends 16 bytes before the end of the file.
3. **Footer** — 16 bytes: a 128-bit MD5 hash of every byte preceding the footer.

A DSF file is structurally:

```
+----------------------------------------+
|  FileHeader (12 bytes)                 |
|    Cookie [8]byte = 'XPLNEDSF'         |
|    Version uint32 = 1                  |
+----------------------------------------+
|  Atoms ... (variable)                  |
|    HEAD                                |
|    DEFN                                |
|    GEOD                                |
|    [DEMS]   (X-Plane 10)               |
|    CMDS                                |
+----------------------------------------+
|  Footer (16 bytes) = MD5(everything    |
|       above this point)                |
+----------------------------------------+
```

The atom IDs (`HEAD`, `DEFN`, `GEOD`, `DEMS`, `CMDS`) are containers; their
sub-atoms carry the actual payloads.

> **Ambiguity:** The spec does not define the order of top-level atoms.
> Convention from X-Plane's writer is `HEAD`, `DEFN`, `GEOD`, optional `DEMS`,
> `CMDS`, but a decoder must accept any order and treat order-sensitive parts
> (e.g. the `POOL`/`SCAL` pairing inside `GEOD`) per their own rules.

> **Ambiguity:** The spec describes 7Z compression as an option but defines it
> only by its leading bytes: a compressed DSF starts with `PK` or `7z`, an
> uncompressed file starts with `XPLNEDSF`. This reference covers only the
> uncompressed wire format; a decoder for compressed input must dispatch
> based on the magic bytes before delegating to the structures below.

## Structures index

- [`structures/file-header.md`](structures/file-header.md) — fixed 12-byte file header (cookie + version)
- [`structures/atom.md`](structures/atom.md) — generic atom envelope (ID + size + payload)
- [`structures/atom-of-atoms.md`](structures/atom-of-atoms.md) — atom whose payload is a concatenation of sub-atoms
- [`structures/string-table-atom.md`](structures/string-table-atom.md) — atom whose payload is a packed list of NUL-terminated strings
- [`structures/planar-numeric-atom.md`](structures/planar-numeric-atom.md) — atom whose payload is a planar (struct-of-arrays) numeric block with optional differencing/RLE
- [`structures/head-atom.md`](structures/head-atom.md) — `HEAD` (atom-of-atoms; carries `PROP`)
- [`structures/prop-atom.md`](structures/prop-atom.md) — `PROP` (string table; pairs of name/value strings)
- [`structures/defn-atom.md`](structures/defn-atom.md) — `DEFN` (atom-of-atoms; carries `TERT`/`OBJT`/`POLY`/`NETW`/`DEMN`)
- [`structures/tert-atom.md`](structures/tert-atom.md) — `TERT` (string table; terrain definition file paths)
- [`structures/objt-atom.md`](structures/objt-atom.md) — `OBJT` (string table; object definition file paths)
- [`structures/poly-atom.md`](structures/poly-atom.md) — `POLY` (string table; polygon definition file paths)
- [`structures/netw-atom.md`](structures/netw-atom.md) — `NETW` (string table; network definition file paths)
- [`structures/demn-atom.md`](structures/demn-atom.md) — `DEMN` (string table; raster layer names; X-Plane 10)
- [`structures/geod-atom.md`](structures/geod-atom.md) — `GEOD` (atom-of-atoms; carries `POOL`/`SCAL`/`PO32`/`SC32`)
- [`structures/pool-atom.md`](structures/pool-atom.md) — `POOL` (planar numeric, `uint16`)
- [`structures/scal-atom.md`](structures/scal-atom.md) — `SCAL` (per-`POOL` scaling/offset pairs as `float32`)
- [`structures/po32-atom.md`](structures/po32-atom.md) — `PO32` (planar numeric, `uint32`)
- [`structures/sc32-atom.md`](structures/sc32-atom.md) — `SC32` (per-`PO32` scaling/offset pairs as `float32`)
- [`structures/dems-atom.md`](structures/dems-atom.md) — `DEMS` (atom-of-atoms; carries `DEMI`/`DEMD` pairs; X-Plane 10)
- [`structures/demi-atom.md`](structures/demi-atom.md) — `DEMI` (raster layer info record)
- [`structures/demd-atom.md`](structures/demd-atom.md) — `DEMD` (raster layer raw pixel bytes)
- [`structures/cmds-atom.md`](structures/cmds-atom.md) — `CMDS` (sequence of variable-length commands)
- [`structures/command.md`](structures/command.md) — generic command framing (1-byte opcode + opcode-specific payload)
- [`structures/footer.md`](structures/footer.md) — fixed 16-byte MD5 file footer

## Encoding tables index

- [`encoding-tables/atom-ids.md`](encoding-tables/atom-ids.md) — defined top-level and nested atom IDs
- [`encoding-tables/command-opcodes.md`](encoding-tables/command-opcodes.md) — opcode → command-name mapping for the `CMDS` atom
- [`encoding-tables/plane-encodings.md`](encoding-tables/plane-encodings.md) — per-plane encoding tag for `POOL`/`PO32` (raw / differenced / RLE / RLE+differenced)
- [`encoding-tables/dem-flags.md`](encoding-tables/dem-flags.md) — `DEMI.Flags` bit layout (number type + post/area-centric)
- [`encoding-tables/patch-flags.md`](encoding-tables/patch-flags.md) — terrain patch bit flags used by `TERRAIN PATCH FLAGS` and `TERRAIN PATCH FLAGS AND LOD`

## Examples index

- [`examples/minimal.md`](examples/minimal.md) — smallest valid DSF (`HEAD`/`PROP` + empty `DEFN`/`GEOD`/`CMDS` + footer)
- [`examples/typical.md`](examples/typical.md) — DSF with one terrain definition, one POOL/SCAL pair (one triangle), and a single `PATCH TRIANGLE` command
- [`examples/complex.md`](examples/complex.md) — DSF that exercises X-Plane 10 raster layers (`DEMN`/`DEMS`/`DEMI`/`DEMD`) plus a `NETW` chain command

## Appendix

- **Maximum sizes**: a definition table (`TERT`/`OBJT`/`POLY`/`NETW`) holds at
  most 65,536 entries. `POOL` indices are `uint16` (max 65,535 points per pool).
  `PO32` indices are `uint32`. Atom payload size is `uint32 size − 8`, so a
  single atom is bounded above by ~4 GB. Junction IDs are 32-bit and start at 1.
- **Related X-Plane file formats** (out of scope here): `.ter` terrain
  definition, `.obj` object definition, `.pol` polygon definition, `.net`
  network definition, `.bmp`/`.png` direct terrain textures.
- **Compatibility**: DSF is intentionally backward compatible but not forward
  compatible. New atoms or new command opcodes break older readers.
- **Version history**: only master version `1` is defined as of 2019-10-03.
