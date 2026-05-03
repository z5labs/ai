# Audit: tlv-binary (binary file library)

**Date:** 2026-05-03
**Spec:** SPEC.md (127 lines), structures/ (3 files), encoding-tables/ (1 file)
**Tests:** PASS — `go test -race ./...` reports `ok tlv 1.012s`

Note: tests pass only because every code path returns the `errUnimplemented` sentinel and the three existing tests assert exactly that. The package is freshly-scaffolded; the audit treats untested behavior as unimplemented per the binary-checklist.

## Summary

- 50 findings across 9 categories (3 phases × 3 categories, with the decoder length/offset/checksum sub-categories collapsed)
- Phases: types (12), decoder (19), encoder (19)
- Severity: blockers (48), warnings (2), info (0)

## Types findings

### Missing struct types
- **[blocker]** structures/header.md § Field table — `Header` struct is not declared in `types.go`; spec requires fields `Magic [4]byte`, `Version uint8`, `Flags Flags`, `Reserved uint16`
- **[blocker]** structures/record.md § Field table — `Record` struct is not declared in `types.go`; spec requires fields `Type RecordType`, `Length uint16`, `Value []byte`
- **[blocker]** structures/trailer.md § Field table — `Trailer` struct is not declared in `types.go`; spec requires field `CRC32 uint32`
- **[blocker]** SPEC.md § Overview (lines 5-12) — `File` struct in `types.go` is empty (`type File struct{}`); spec mandates a top-level container holding `Header`, a slice of `Record`, and `Trailer`
- **[blocker]** structures/header.md § Header.Flags bit field — `Flags` type (single-byte bit field) is not declared in `types.go`; missing constants `COMPRESSED = 0x01`, `ENCRYPTED = 0x02`, `SIGNED = 0x04`, and reserved-bits mask `0xF8`
- **[warning]** structures/header.md § Header.Flags bit field — even after `Flags` is added, no `String()` method exists yet; the implement-skill mandates `String()` on every typed-integer/enum so hex-dump test failures are readable
- **[blocker]** `types_test.go` — no test pins `Header`, `Record`, `Trailer`, or `File` field shapes against the spec; per the binary-checklist, untested behavior is treated as unimplemented
- **[blocker]** structures/header.md § Notes — Reserved-must-be-zero invariant has no Go-side representation (no validator/constructor) and no test pins it

### Encoding-table coverage
- **[blocker]** encoding-tables/record-type.md — `RecordType` typed enum is not declared in `types.go`; missing one constant per row: `STRING = 0x01`, `INT = 0x02`, `BLOB = 0x03`, `NESTED = 0x04`
- **[warning]** encoding-tables/record-type.md — when `RecordType` is added, it must have a `String()` method per the binary-checklist; no such method exists in `types.go`
- **[blocker]** encoding-tables/record-type.md § Notes — `INT` length-must-equal-8 invariant has no Go-side representation and no test; the table notes call this out explicitly as a typed-error rejection rule
- **[blocker]** `types_test.go` — no test exercises `RecordType` constant values against `encoding-tables/record-type.md` (e.g. asserting `STRING == 0x01`); without it the table mapping is unverified

### Drift
- (none) — no struct types in `types.go` are absent from `structures/`; the only declared struct (`File`) is named in `structures/`-implied terms (the spec implies a top-level container even though there is no `structures/file.md`). The `File` struct exists but is empty, which is a missing-field blocker captured above, not drift.

## Decoder findings

### Unread fields
- **[blocker]** structures/header.md § Field table — `decoder.go` has no `readHeader` method; `Header.Magic`, `Header.Version`, `Header.Flags`, `Header.Reserved` are all unread (the only `readX` method is `readFile`, which returns `errUnimplemented`)
- **[blocker]** structures/record.md § Field table — `decoder.go` has no `readRecord` method; `Record.Type`, `Record.Length`, `Record.Value` are all unread
- **[blocker]** structures/trailer.md § Field table — `decoder.go` has no `readTrailer` method; `Trailer.CRC32` is not read
- **[blocker]** SPEC.md § Overview (lines 5-12) — `decoder.readFile` is a stub returning `errUnimplemented`; the spec-defined Header → Records → Trailer sequence is not consumed at all (see also: Drift)

### Missing length/offset/checksum checks

#### Length checks
- **[blocker]** structures/record.md § Field table — `Record.Length` is a count of subsequent `Value` bytes; no decoder code uses it to bound a `Value` read because no `readRecord` exists
- **[blocker]** SPEC.md § Overview line 14 — "Records are read in order until the bytes preceding the trailer are exhausted" requires the decoder to track the total file length (or peek for the trailer's leading position) before each record loop iteration; no such bound exists in `decoder.go`
- **[blocker]** encoding-tables/record-type.md § Notes — for `Type = INT` the decoder must validate `Length == 8` and surface a typed error otherwise; no such check exists
- **[blocker]** structures/record.md § Field table — `Length` is `uint16` so values 0..65535 are syntactically valid; no test asserts that an `INT` record with `Length != 8` is rejected with the typed `FieldError → OffsetError → leaf` chain

#### Offset checks
- (none) — TLV1 has no inter-structure offsets (records and trailer are positional, not pointer-addressed); category not applicable to this format

#### Checksum checks
- **[blocker]** structures/trailer.md § Notes / SPEC.md § Checksums and Integrity (lines 86-88) — decoder must compute a running IEEE CRC32 over every byte from offset 0 up to (but not including) the CRC32 itself, then compare to the value in the trailer; `decoder.go` neither computes nor compares CRC32 (no `hash/crc32` import, no running checksum field on `decoder`)
- **[blocker]** `decoder_test.go` — no failure-path test asserts the `FieldError → OffsetError → leaf` chain via `errors.Is`/`errors.As` for a CRC32 mismatch; per the binary-checklist, missing rejection-rule tests are blockers and each rule gets its own bullet

### Drift
- **[blocker]** SPEC.md § Conditional and Optional Fields line 83 / structures/header.md § Notes — readers must fail with a typed error if `Header.Reserved != 0`; `decoder.go` has no such rejection rule and `decoder_test.go` has no test pinning the rejection
- **[blocker]** SPEC.md § Conditional and Optional Fields line 84 / structures/header.md § Notes — readers must fail if the `SIGNED` flag bit is set (extension not specified in this version); `decoder.go` has no such rejection rule and `decoder_test.go` has no test pinning the rejection
- **[blocker]** SPEC.md § Versioning line 96 — decoder must reject files with an unrecognized `Header.Version` using a typed error; `decoder.go` has no version check and `decoder_test.go` has no rejection test
- **[blocker]** structures/header.md § Field table — decoder must reject files whose `Header.Magic` is not the ASCII bytes `54 4C 56 31` ("TLV1"); `decoder.go` has no magic check and `decoder_test.go` has no rejection test
- **[blocker]** encoding-tables/record-type.md § Notes — unknown `RecordType` values must surface as a typed error; `decoder.go` has no such rejection rule and `decoder_test.go` has no rejection test
- **[blocker]** SPEC.md § Conditional and Optional Fields line 82 — a file with zero records is legal (header followed immediately by trailer); `decoder_test.go` has no test for the zero-record case (matches the SPEC.md § Examples "Minimal" fixture, lines 100-108)
- **[blocker]** SPEC.md § Examples (lines 110-118, "Typical: one STRING record") — no decoder test loads the spec's golden bytes and asserts the resulting `File` struct; the Examples section is the format's gold-standard fixture
- **[blocker]** SPEC.md § Examples (lines 120-127, "Complex: COMPRESSED flag set, two records") — no decoder test loads these golden bytes either; flag handling and multi-record decoding are unverified
- **[blocker]** `decoder_test.go` — the only test (`TestDecodeStubReturnsErrUnimplemented`) is a stub-marker that asserts the placeholder error chain, not spec behavior; no `TestDecodeHeader`, `TestDecodeRecord`, `TestDecodeTrailer`, or full-file decode test exists

(Tests pass — `go test -race ./...` is `ok tlv 1.012s` — but they pass only because every code path returns `errUnimplemented` and the only assertion is `ErrorIs(..., errUnimplemented)`. Per binary-checklist § Test-status integration, no failing-test findings are added; the absence of real tests is captured above as `[blocker]` rejection-rule and round-trip gaps instead.)

## Encoder findings

### Unwritten fields
- **[blocker]** structures/header.md § Field table — `encoder.go` has no `writeHeader` method; `Header.Magic`, `Header.Version`, `Header.Flags`, `Header.Reserved` are all unwritten (the only `writeX` method is `writeFile`, which returns `errUnimplemented`)
- **[blocker]** structures/record.md § Field table — `encoder.go` has no `writeRecord` method; `Record.Type`, `Record.Length`, `Record.Value` are all unwritten
- **[blocker]** structures/trailer.md § Field table — `encoder.go` has no `writeTrailer` method; `Trailer.CRC32` is not written
- **[blocker]** structures/record.md § Field table — once `writeRecord` exists it must compute `Length` from `len(Value)` rather than copying `r.Length` verbatim (otherwise a caller forgetting to set `Length` produces a structurally invalid file); no such computation exists yet
- **[blocker]** structures/trailer.md § Notes / SPEC.md § Checksums and Integrity (lines 86-88) — `writeTrailer` must compute the IEEE CRC32 over every byte from offset 0 up to (but not including) the CRC32 itself, then write the value; `encoder.go` has no `hash/crc32` import and no running checksum tracking on the `encoder` struct
- **[blocker]** structures/header.md § Notes — encoder must enforce `Header.Reserved == 0` on write (or unconditionally write zero); no such logic exists
- **[blocker]** SPEC.md § Conditional and Optional Fields line 84 — encoder should reject (or refuse to set) the `SIGNED` flag bit since this version does not specify the signature layout; no such check exists

### Round-trip test coverage
- **[blocker]** `encoder.go writeFile` — no round-trip test in `encoder_test.go`; the only test (`TestEncodeStubReturnsErrUnimplemented`) asserts the placeholder error chain, not `Encode → Decode → require.Equal`
- **[blocker]** structures/header.md § Field table — no round-trip test exercises `Header` (would also need `Header` to exist as a struct first; see types phase)
- **[blocker]** structures/record.md § Field table — no round-trip test exercises a `Record`; in particular no round-trip covers `STRING`, `INT`, `BLOB`, or `NESTED` record types from `encoding-tables/record-type.md`
- **[blocker]** structures/trailer.md § Field table — no round-trip test exercises `Trailer`; CRC32 compute-on-write / verify-on-read symmetry is unverified
- **[blocker]** structures/header.md § Header.Flags bit field — no round-trip test exercises `Flags`; `COMPRESSED`, `ENCRYPTED`, and the SIGNED-rejection paths are all unexercised
- **[blocker]** SPEC.md § Conditional and Optional Fields line 82 — no round-trip test exercises the zero-records case (header immediately followed by trailer); matches the SPEC.md § Examples "Minimal" fixture but no test loads it
- **[blocker]** SPEC.md § Examples (lines 100-127) — no Examples-bytes round-trip test exists for any of the three example fixtures (Minimal, Typical-STRING, Complex-COMPRESSED-two-records); per the binary-checklist, missing Examples-bytes round-trip tests are blockers because the Examples section is the spec's gold-standard fixture

### Drift
- **[blocker]** SPEC.md § Examples (lines 102-108, "Minimal") — encoder output for an empty `File` is not pinned against the spec's golden bytes (`54 4C 56 31 01 00 00 00` + 4-byte CRC32); without an Examples-bytes test, drift between the encoder's CRC32 computation and the spec's CRC32 definition can pass round-trip silently if the decoder shares the same bug
- **[blocker]** SPEC.md § Examples (lines 114-118, "Typical: one STRING record") — encoder output for a single-STRING-record `File` is not pinned against the spec's golden header bytes (`01 00 05 68 65 6C 6C 6F`)
- **[blocker]** SPEC.md § Examples (lines 122-127, "Complex") — encoder output for a two-record file with the `COMPRESSED` flag set is not pinned against the spec's golden bytes; flag-byte placement (offset 5 in the header) and inter-record contiguity (no padding per SPEC.md § Padding and Alignment) are unverified
- **[blocker]** SPEC.md § Conventions line 18 — big-endian byte order is mandated for all multi-byte integers; the encoder declares `byteOrder: binary.BigEndian` but never uses it (only `writeFile` exists, returning `errUnimplemented`), so byte-order compliance is unpinned by any test
- **[blocker]** SPEC.md § Padding and Alignment (lines 90-92) — there is no padding between header/records/trailer; no encoder test asserts the absence of padding (would be caught by an Examples-bytes round-trip test, listed above)

(Tests pass — `go test -race ./...` is `ok tlv 1.012s` — but `TestEncodeStubReturnsErrUnimplemented` only asserts the stub's placeholder error chain, so no failing-test findings are added per binary-checklist § Test-status integration; the lack of real coverage is captured as `[blocker]` round-trip gaps above.)
