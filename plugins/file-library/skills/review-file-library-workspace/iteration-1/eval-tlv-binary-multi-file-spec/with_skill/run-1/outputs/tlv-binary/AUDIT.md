# Audit: tlv-binary (binary file library)

**Date:** 2026-05-03
**Spec:** SPEC.md (127 lines), structures/ (3 files: header.md, record.md, trailer.md), encoding-tables/ (1 file: record-type.md)
**Tests:** PASS

## Summary

- 33 findings across 7 categories
- Phases: types (7), decoder (16), encoder (10)
- Severity: blockers (28), warnings (5), info (0)

> **Context.** The package is freshly scaffolded: `types.go` declares only an empty `File` struct and the error chain, `decoder.go` and `encoder.go` are single-method stubs that return `errUnimplemented`. Tests pass because the only assertions are that the stubs return `errUnimplemented`. As a result every spec-defined structure, field, validation, and round-trip test is reported as missing. Re-run this audit after `Header`/`Record`/`Trailer`/`Flags`/`RecordType` and the `readX`/`writeX` methods land — at that point Drift findings become meaningful.

## Types findings

### Missing struct types

- **[blocker]** structures/header.md § Field table — no `Header` struct declared in `types.go`; spec defines a fixed 8-byte header with `Magic [4]byte`, `Version uint8`, `Flags Flags`, `Reserved uint16` fields, none of which are represented.
- **[blocker]** structures/record.md § Field table — no `Record` struct declared in `types.go`; spec defines a variable-length record with `Type RecordType`, `Length uint16`, `Value []byte` fields, none of which are represented.
- **[blocker]** structures/trailer.md § Field table — no `Trailer` struct declared in `types.go`; spec defines a fixed 4-byte trailer with `CRC32 uint32`, not represented.
- **[blocker]** SPEC.md § Overview (lines 5-12) and structures/header.md § Header.Flags bit field — no `Flags` typed-byte declared in `types.go`; spec mandates a bit field with `COMPRESSED` (0x01), `ENCRYPTED` (0x02), `SIGNED` (0x04) constants, none declared. Missing `String()` method on the enum is also required by the package's binary-skill convention (would be `[warning]` on its own, but the type itself is absent so it rolls into the blocker).
- **[blocker]** SPEC.md § Overview (lines 5-12) — `File` struct in `types.go` (line 10) is empty (`type File struct{}`); spec mandates the top-level decoded representation expose `Header`, `Records []Record`, and `Trailer` fields. Without these the rest of the pipeline cannot represent any TLV1 input.

### Encoding-table coverage

- **[blocker]** encoding-tables/record-type.md § Record.Type values — no `RecordType` typed-uint8 declared in `types.go`; spec defines four constants (`STRING = 0x01`, `INT = 0x02`, `BLOB = 0x03`, `NESTED = 0x04`). Decoder/encoder cannot translate the Type byte without this enum.
- **[warning]** encoding-tables/record-type.md — even after `RecordType` is added, the binary-skill convention requires a `String()` method on every enum for hex-dump test failure messages; flag now so the implementer adds it alongside the type.

### Drift

- (none)

## Decoder findings

### Unread fields

- **[blocker]** structures/header.md § Field table — `Header.Magic` defined in spec but not read by `decoder.go`; `readFile` (line 37) returns `errUnimplemented` without consuming any bytes. Spec also requires the magic be validated against ASCII `"TLV1"` (`0x54 0x4C 0x56 0x31`).
- **[blocker]** structures/header.md § Field table — `Header.Version` defined in spec but not read; SPEC.md § Versioning (line 96) additionally mandates the decoder reject unrecognized versions with a typed error.
- **[blocker]** structures/header.md § Field table — `Header.Flags` defined in spec but not read; SPEC.md § Conditional and Optional Fields (line 84) additionally mandates the decoder fail when the SIGNED flag is encountered.
- **[blocker]** structures/header.md § Field table and § Notes — `Header.Reserved` defined in spec but not read; spec mandates the decoder fail with a typed error when Reserved ≠ 0.
- **[blocker]** structures/record.md § Field table — `Record.Type` defined in spec but not read by any `readRecord` method (no such method exists in `decoder.go`).
- **[blocker]** structures/record.md § Field table — `Record.Length` defined in spec but not read.
- **[blocker]** structures/record.md § Field table — `Record.Value` defined in spec but not read.
- **[blocker]** structures/trailer.md § Field table — `Trailer.CRC32` defined in spec but not read by any `readTrailer` method (no such method exists in `decoder.go`).
- **[blocker]** SPEC.md § Overview (line 14) — spec mandates "Records are read in order until the bytes preceding the trailer are exhausted"; decoder has no record loop. Without it, `File.Records` cannot be populated.

### Missing length/offset/checksum checks

- **[blocker]** structures/record.md § Field table — `Record.Length` is a length field that bounds the read of `Value`; no `readRecord` exists, so no length check exists. Once added, the length must bound the `Value` read.
- **[blocker]** structures/trailer.md § Notes and SPEC.md § Checksums and Integrity (lines 86-88) — `Trailer.CRC32` (IEEE CRC32) is defined; `decoder.go` has no `readTrailer` and no running-CRC32 computation. Spec mandates the decoder compute the running CRC32 as it reads, then compare it to the trailer value, returning a typed error on mismatch.
- **[blocker]** encoding-tables/record-type.md § Notes — for `Type = INT`, the decoder must validate `Length == 8` and surface a typed error otherwise; this validation is not implemented (no record handling exists).
- **[blocker]** encoding-tables/record-type.md § Notes and SPEC.md § Encoding Tables (line 78) — unknown record types must surface as a typed error; not implemented (no record handling exists).
- **[blocker]** structures/record.md § Notes — for `Type = NESTED`, `Value` is itself a complete TLV1 file; recursive decode of nested files is not implemented.

### Drift

- **[warning]** `decoder.go readFile` (line 37) — the only test that exercises the decoder, `TestDecodeStubReturnsErrUnimplemented` in `decoder_test.go`, asserts the stub returns `errUnimplemented`; once the real implementation lands, that test will become drift evidence (it pins stub behavior, not spec behavior). No test currently decodes the spec's Examples bytes (SPEC.md lines 98-127, three example layouts: minimal, typical STRING record, COMPRESSED flag two-record case) — so there is no failure-driving test that pins spec behavior.
- **[warning]** `decoder_test.go` — no failure-path test that asserts the `FieldError → OffsetError → leaf` chain via `errors.Is` / `errors.As` against a *real* decode failure (truncated input, bad magic, bad version, Reserved ≠ 0, CRC mismatch, unknown record type, INT length ≠ 8). The existing `TestDecodeStubReturnsErrUnimplemented` only walks the chain for the stub error; the implement-skill convention requires per-failure-mode coverage.

## Encoder findings

### Unwritten fields

- **[blocker]** structures/header.md § Field table — `Header.Magic` defined in spec but not written by `encoder.go`; `writeFile` (line 36) returns `errUnimplemented` without emitting any bytes. Encoder must write the constant ASCII `"TLV1"` (`0x54 0x4C 0x56 0x31`) regardless of struct contents.
- **[blocker]** structures/header.md § Field table — `Header.Version` defined in spec but not written.
- **[blocker]** structures/header.md § Field table — `Header.Flags` defined in spec but not written.
- **[blocker]** structures/header.md § Field table and § Notes — `Header.Reserved` defined in spec but not written; spec mandates this field be zero on write — the encoder must emit `0x00 0x00` and not trust an arbitrary struct value.
- **[blocker]** structures/record.md § Field table — `Record.Type`, `Record.Length`, and `Record.Value` defined in spec but not written by any `writeRecord` method (no such method exists in `encoder.go`).
- **[blocker]** structures/record.md § Field table — once `writeRecord` exists, `Length` must be *computed* from `len(Value)`, not read from a `Record.Length` struct field; the binary-checklist convention requires length/offset/checksum fields to be derived, not trusted from the struct.
- **[blocker]** structures/trailer.md § Field table and SPEC.md § Checksums and Integrity (lines 86-88) — `Trailer.CRC32` defined in spec but not written; no `writeTrailer` exists. Encoder must compute IEEE CRC32 over every byte from offset 0 (start of header) up to but not including the CRC32 itself, then write the resulting `uint32` big-endian. The CRC must be computed, not trusted from a struct field.
- **[blocker]** SPEC.md § Overview (line 14) — encoder has no record-emission loop; `File.Records` (which itself does not exist on the empty `File` struct) cannot be serialized.

### Round-trip test coverage

- **[warning]** `encoder.go writeFile` (line 36) — `encoder_test.go` contains only `TestEncodeStubReturnsErrUnimplemented`, which asserts the stub returns `errUnimplemented`. There is no round-trip test (`Encode → Decode → require.Equal`) for any structure. The implement-skill convention mandates a round-trip test for every `writeX` method; once `writeHeader`, `writeRecord`, and `writeTrailer` are added, each needs its own round-trip test.
- **[warning]** SPEC.md § Examples (lines 98-127) — no Examples-bytes round-trip test exists (decode the spec's golden bytes, encode back, assert byte-for-byte equality). The spec provides three example layouts (minimal, typical STRING record, COMPRESSED flag two-record case); these are the gold-standard drift detector for catching encoder/decoder symmetric bugs and should each have a test.

### Drift

- (none — encoder is a stub, so there is no implementation behavior to drift from spec yet. Re-audit after `writeHeader`, `writeRecord`, and `writeTrailer` land.)
