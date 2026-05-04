# Audit: dsf (binary file library)

**Date:** 2026-05-04
**Spec:** SPEC.md (139 lines), structures/ (24 files), encoding-tables/ (5 files)
**Tests:** PASS

## Summary

- 41 findings across 3 phases / 9 categories
- Phases: types (16), decoder (15), encoder (10)
- Severity: blockers (37), warnings (3), info (1)

## Types findings

### Missing struct types

- **[blocker]** structures/prop-atom.md — no `PropAtom` (or equivalent) struct in `types.go`; `Atom.Payload` is opaque `[]byte` and `PROP` name/value pairs are not modelled.
- **[blocker]** structures/head-atom.md — no `HeadAtom` struct; `HEAD` is currently just a generic `Atom`.
- **[blocker]** structures/defn-atom.md — no `DefnAtom` struct; `DEFN` is opaque.
- **[blocker]** structures/tert-atom.md — no `TertAtom` (terrain definitions string table) representation.
- **[blocker]** structures/objt-atom.md — no `ObjtAtom` (object definitions string table) representation.
- **[blocker]** structures/poly-atom.md — no `PolyAtom` (polygon definitions string table) representation.
- **[blocker]** structures/netw-atom.md — no `NetwAtom` (network definitions string table) representation.
- **[blocker]** structures/demn-atom.md — no `DemnAtom` (raster layer names) representation.
- **[blocker]** structures/geod-atom.md — no `GeodAtom` (atom-of-atoms with POOL/SCAL/PO32/SC32) representation.
- **[blocker]** structures/pool-atom.md — no `PoolAtom` (planar uint16 pool) representation; planar-numeric encoding (raw/differenced/RLE/RLE+differenced) is not implemented.
- **[blocker]** structures/scal-atom.md — no `ScalAtom` (float32 scaling pairs) representation.
- **[blocker]** structures/po32-atom.md — no `Po32Atom` (planar uint32 pool) representation.
- **[blocker]** structures/sc32-atom.md — no `Sc32Atom` representation.
- **[blocker]** structures/dems-atom.md — no `DemsAtom` representation.
- **[blocker]** structures/demi-atom.md — no `DemiAtom` (20-byte raster layer info record) representation.
- **[blocker]** structures/demd-atom.md — no `DemdAtom` representation.
- **[blocker]** structures/cmds-atom.md — no `CmdsAtom` / `Command` representation; opcode dispatch is not implemented.
- **[blocker]** structures/command.md — no per-opcode payload structures (POLYGON, NETWORK_CHAINS, PATCH_TRIANGLE, COMMENT_8/16/32, …); 30+ command variants are missing.
- **[blocker]** structures/planar-numeric-atom.md — the planar-numeric framing (`ItemCount`, `PlaneCount`, per-plane encoding tag) and the four encoding modes (raw / differenced / RLE / RLE+differenced) have no Go representation.
- **[blocker]** structures/string-table-atom.md — no `StringTableAtom` (or `[]string`) helper exists; string-table walking is not implemented.
- **[blocker]** structures/atom-of-atoms.md — no recursive sub-atom decoding helper; `Atom.Payload` is left as `[]byte`.

### Encoding-table coverage

- **[blocker]** encoding-tables/atom-ids.md — atom-ID constants (`'HEAD'`, `'PROP'`, `'GEOD'`, `'CMDS'`, …) are not declared in `types.go`. Decoder is `Atom`-generic so it does not need them yet, but per-atom dispatch in later iterations will.
- **[blocker]** encoding-tables/command-opcodes.md — no `Opcode` (or similar) typed-integer enum with constants for opcodes 1–18, 23–34, plus the 19–22 unassigned gap and 255 reserved; required before `CmdsAtom` decoding is implemented.
- **[blocker]** encoding-tables/plane-encodings.md — no `PlaneEncoding` enum (`RAW=0`, `DIFFERENCED=1`, `RLE=2`, `RLE_DIFFERENCED=3`).
- **[blocker]** encoding-tables/dem-flags.md — no `DEMFlags` bit-field type (NumberType bits 1–0, PostCentric bit 2).
- **[blocker]** encoding-tables/patch-flags.md — no `PatchFlags` bit-field type (`PHYSICAL=0x01`, `OVERLAY=0x02`).

### Drift

- **[warning]** types.go declares a placeholder `Kind` enum with `KindUnknown`/`KindExample` — this is scaffold residue and not represented anywhere in `structures/`. Remove once a real DSF enum (e.g. `Opcode`) replaces it.

## Decoder findings

### Unread fields

- **[blocker]** structures/prop-atom.md — `decoder.go` does not parse `PROP` payloads into name/value pairs; `Atom.Payload` is left opaque.
- **[blocker]** structures/defn-atom.md — `decoder.go` does not walk `DEFN` sub-atoms (`TERT`/`OBJT`/`POLY`/`NETW`/`DEMN`).
- **[blocker]** structures/tert-atom.md — string-table walking not implemented; `TERT` payload not parsed.
- **[blocker]** structures/objt-atom.md — same — `OBJT` payload not parsed.
- **[blocker]** structures/poly-atom.md — same — `POLY` payload not parsed.
- **[blocker]** structures/netw-atom.md — same — `NETW` payload not parsed.
- **[blocker]** structures/demn-atom.md — same — `DEMN` payload not parsed.
- **[blocker]** structures/pool-atom.md — `decoder.go` does not implement `POOL` planar-numeric decoding (item-count + plane-count + per-plane encoding + RAW/DIFFERENCED/RLE/RLE_DIFFERENCED).
- **[blocker]** structures/scal-atom.md — `SCAL` not parsed into `[]float32`.
- **[blocker]** structures/po32-atom.md — `PO32` planar-numeric decoding not implemented.
- **[blocker]** structures/sc32-atom.md — `SC32` not parsed.
- **[blocker]** structures/demi-atom.md — `DEMI`'s 20-byte fixed-layout record (Version, BPP, Flags, Width, Height, Scale, Offset) not parsed.
- **[blocker]** structures/demd-atom.md — `DEMD` raw-pixel-byte payload not extracted (it is currently captured as opaque `Atom.Payload` bytes; that is fine for round-trip but the `Width × Height × BytesPerPixel` length check is missing).
- **[blocker]** structures/cmds-atom.md / command.md — opcode dispatch and per-opcode payload decoding are not implemented; no `readCommand` method exists.

### Missing length/offset/checksum checks

- **[blocker]** SPEC.md § Conventions — Atom `Size` is verified against `>= 8` and `<= remaining` but the `RLE` count semantics inside `POOL`/`PO32` planes (count of *elements* vs *bytes*) is not implemented and therefore cannot be checked.
- (none — top-level container's length checks are in place: `Size < 8`, `Size > remaining`, total file `< 28` truncation, and footer MD5 are all verified.)

### Drift

- (none — the top-level container decoder agrees with the spec's `## Conventions` and `## Top-level structure` sections.)

## Encoder findings

### Unwritten fields

- **[blocker]** structures/prop-atom.md — `encoder.go` cannot serialize a `PropAtom` (no such type yet).
- **[blocker]** structures/defn-atom.md — same gap.
- **[blocker]** structures/geod-atom.md — same gap; planar-numeric encoding (RAW/DIFFERENCED/RLE/RLE_DIFFERENCED) not implemented.
- **[blocker]** structures/dems-atom.md — same gap.
- **[blocker]** structures/cmds-atom.md / command.md — no per-opcode `writeCommand` methods; encoder cannot construct a `CMDS` payload from typed commands.
- **[blocker]** structures/demi-atom.md — `DEMI` encoder method missing (20-byte fixed-layout record).
- **[blocker]** structures/planar-numeric-atom.md — encoder does not compute the per-plane RLE byte stream or apply differencing.

### Round-trip test coverage

- **[info]** The current `TestEncodeDecodeRoundTrip` covers the top-level container at the `Atom`-as-`[]byte` level and confirms byte-identical round-trip — strong for the implemented surface, but does not exercise any per-atom payload semantics (because none are implemented).
- **[blocker]** structures/* — once typed sub-atom encoders land, each needs its own `Encode → Decode → Equal` round-trip test; none exist today.
- **[blocker]** examples/typical.md and examples/complex.md — no test loads either example's bytes and asserts `Decode → Encode → bytes-equal`. The `examples/minimal.md` annotation is also untested. Examples-bytes round-trip is the spec's own gold standard fixture; not testing against it is leaving compliance unverified.

## Notes

This audit was run against an iteration-1 implementation that intentionally
stops at the top-level container. The blockers above are honest reflections
of the unimplemented surface, not regressions: they list what would need to
land in iterations 2..N to get full DSF spec coverage.
