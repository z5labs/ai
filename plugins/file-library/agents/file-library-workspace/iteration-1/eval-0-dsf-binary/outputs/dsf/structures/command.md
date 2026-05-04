# Command

A single command record inside the [`CMDS`](cmds-atom.md) atom. Every
command shares the same 1-byte opcode prefix; the body shape depends on
the opcode.

## Encoding

```
+--------+--------------------------------+
| u8     | opcode-specific payload        |
| Opcode |                                |
+--------+--------------------------------+
```

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | `uint8` | Opcode | One of the values in [`../encoding-tables/command-opcodes.md`](../encoding-tables/command-opcodes.md). |
| 1 | variable | (per opcode) | Payload | Layout depends on `Opcode` — see below. |

## Per-opcode payload layouts

All multi-byte fields are little-endian. `N` is taken from a count field
read inline.

### State selection (1–6)

- **`1` COORDINATE POOL SELECT** — `uint16 PoolIndex`. Sets the current
  16-bit pool to `POOL[PoolIndex]`.
- **`2` JUNCTION OFFSET SELECT** — `uint32 Offset`. Added to all 16-bit
  vector indices in subsequent network commands (except `NETWORK CHAIN 32`).
- **`3` SET DEFINITION 8** — `uint8 DefIndex`.
- **`4` SET DEFINITION 16** — `uint16 DefIndex`.
- **`5` SET DEFINITION 32** — `uint32 DefIndex`.
- **`6` SET ROAD SUBTYPE 8** — `uint8 RoadSubtype`.

### Object placement (7–8)

- **`7` OBJECT** — `uint16 CoordIndex`. Places the current-definition object
  at point `CoordIndex` in the current pool.
- **`8` OBJECT RANGE** — `uint16 First, uint16 LastPlusOne`. Places one
  object per index in `[First, LastPlusOne)`.

### Network (9–11)

The current pool for these commands is the most recently selected `PO32`
(via the appropriate command — see [Ambiguity](#ambiguities) below).

- **`9` NETWORK CHAINS** — `uint8 N` then `N × uint16 Indices`. Indices have
  the current `JunctionOffset` added before pool lookup.
- **`10` NETWORK CHAINS RANGE** — `uint16 First, uint16 LastPlusOne`.
  Indices have the `JunctionOffset` added.
- **`11` NETWORK CHAIN 32** — `uint8 N` then `N × uint32 Indices`. The
  `JunctionOffset` is **not** applied (these are absolute 32-bit indices).

### Polygon (12–15)

- **`12` POLYGON** — `uint16 Param, uint8 N, N × uint16 Indices`.
- **`13` POLYGON RANGE** — `uint16 Param, uint16 First, uint16 LastPlusOne`.
- **`14` NESTED POLYGON** — `uint16 Param, uint8 W` (number of windings),
  then for each winding: `uint8 M, M × uint16 Indices`.
- **`15` NESTED POLYGON RANGE** — `uint16 Param, uint8 N, N × uint16 Indices`.
  The polygon has `N − 1` windings; each pair of consecutive indices
  defines `[start, end+1)` for one winding.

### Mesh (16–18, 23–31)

> **Ambiguity:** The spec lists no opcodes for IDs 19–22; those slots are
> simply skipped, going from `18 TERRAIN PATCH FLAGS AND LOD` directly to
> `23 PATCH TRIANGLE`. A defensive decoder must treat 19–22 as unassigned
> and reject them as malformed input until a later spec version defines
> them.

- **`16` TERRAIN PATCH** — no payload. Starts a new patch reusing the prior
  patch's LOD range and flags.
- **`17` TERRAIN PATCH FLAGS** — `uint8 Flags`. Reuses the prior LOD range.
  See [`../encoding-tables/patch-flags.md`](../encoding-tables/patch-flags.md).
- **`18` TERRAIN PATCH FLAGS AND LOD** — `uint8 Flags, float32 NearLOD,
  float32 FarLOD`.
- **`23` PATCH TRIANGLE** — `uint8 N, N × uint16 Indices`. `N` is required
  to be a multiple of 3; each consecutive triple is a triangle.
- **`24` TRIANGLE PATCH CROSS-POOL** — `uint8 N, 2N × uint16 Pairs` where
  each pair is `(PoolIndex, CoordIndex)`. `N` is required to be a multiple
  of 3.
- **`25` PATCH TRIANGLE RANGE** — `uint16 First, uint16 LastPlusOne`. The
  range must be a multiple of 3.
- **`26` PATCH TRIANGLE STRIP** — `uint8 N, N × uint16 Indices`.
- **`27` PATCH TRIANGLE STRIP CROSS-POOL** — `uint8 N, 2N × uint16 Pairs`.
- **`28` PATCH TRIANGLE STRIP RANGE** — `uint16 First, uint16 LastPlusOne`.
- **`29` PATCH TRIANGLE FAN** — `uint8 N, N × uint16 Indices`.
- **`30` PATCH TRIANGLE FAN CROSS-POOL** — `uint8 N, 2N × uint16 Pairs`.
- **`31` PATCH TRIANGLE FAN RANGE** — `uint16 First, uint16 LastPlusOne`.

### Comments (32–34)

Comments embed arbitrary bytes; the only difference between the three
opcodes is the width of the length prefix.

- **`32` COMMENT 8** — `uint8 Length, Length × uint8 Data`. (1..255 byte
  payload.)
- **`33` COMMENT 16** — `uint16 Length, Length × uint8 Data`. (1..65535
  byte payload.)
- **`34` COMMENT 32** — `uint32 Length, Length × uint8 Data`. (1..2^32-1
  byte payload.)

### Reserved

- **`255`** — reserved for future expansion (the spec says so explicitly).
- All other opcodes (`0`, `19–22`, `35–254`) are unassigned. A decoder
  must reject them.

## Variable-length fields

Many commands (`9`, `12`, `14`, `15`, `23`, `24`, `26`, `27`, `29`, `30`,
`32`, `33`, `34`) carry a count or length prefix immediately after the
opcode. Every count/length is unsigned and counts elements (or bytes for
comments), exclusive of the prefix itself.

## Ambiguities

> **Ambiguity:** The spec does not say which `PO32` is current for network
> commands. The natural reading is: `COORDINATE POOL SELECT` (opcode 1)
> takes a unified pool index that selects from the appropriate pool list
> for the next command's family — `POOL` for mesh/object/polygon and
> `PO32` for vector. A defensive implementer should mirror this and
> validate against a known-good fixture before shipping.

> **Ambiguity:** The spec text on `NESTED POLYGON RANGE` (opcode 15) is
> sparse. The reading above ("`N` indices, `N − 1` windings, each
> `[indices[i], indices[i+1])`") matches X-Plane's writer convention.
