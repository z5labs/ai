# TLV1 Package Audit

Audit of the freshly-scaffolded `tlv-binary` package against `SPEC.md` and the chunked spec tree (`structures/header.md`, `structures/record.md`, `structures/trailer.md`, `encoding-tables/record-type.md`).

## Summary

The package is a scaffold: `types.go` defines only the empty `File` struct plus error wrappers; `decoder.go` and `encoder.go` are entry points that immediately return `errUnimplemented`; tests only assert that the stub returns `errUnimplemented` wrapped in the `FieldError → OffsetError → leaf` chain. **None** of the format-bearing types, fields, decode steps, encode steps, validation rules, or round-trip behaviors required by the spec are implemented yet.

Findings are categorized below. Each finding is tagged with a stable ID and a severity:

- **BLOCKER** — package cannot represent or process a conformant TLV1 file at all.
- **HIGH** — spec-mandated validation/encoding semantics that would silently corrupt or accept bad files.
- **MEDIUM** — coverage gaps (tests, encoding tables) that hide regressions.
- **LOW** — drift / cosmetic mismatches between SPEC.md and the chunked spec.

Total findings: **54** (21 BLOCKER, 19 HIGH, 10 MEDIUM, 4 LOW).

Breakdown by section: types (7), decoder (9), encoder (8), validation rules (14), encoding-table coverage (6), round-trip / test coverage (6), drift (4).

---

## 1. Missing struct types in `types.go`

`types.go` declares only `File struct{}` (empty), `ErrInvalid`, `errUnimplemented`, `OffsetError`, and `FieldError`. Every domain type required by the spec is missing.

| ID    | Severity | Finding |
|-------|----------|---------|
| T-01  | BLOCKER  | `File` has zero fields. Per `SPEC.md` §Overview and `structures/*.md`, it must hold `Header`, `Records []Record`, and `Trailer`. |
| T-02  | BLOCKER  | No `Header` struct. Per `structures/header.md`, must hold `Magic [4]byte`, `Version uint8`, `Flags Flags`, `Reserved uint16`. |
| T-03  | BLOCKER  | No `Record` struct. Per `structures/record.md`, must hold `Type RecordType`, `Length uint16`, `Value []byte`. |
| T-04  | BLOCKER  | No `Trailer` struct. Per `structures/trailer.md`, must hold `CRC32 uint32`. |
| T-05  | HIGH     | No `Flags` named type with mask constants (`COMPRESSED = 0x01`, `ENCRYPTED = 0x02`, `SIGNED = 0x04`, reserved mask `0xF8`). `structures/header.md` types this field as `Flags`, not `uint8`. |
| T-06  | HIGH     | No `RecordType` named type with constants (`STRING = 0x01`, `INT = 0x02`, `BLOB = 0x03`, `NESTED = 0x04`). `structures/record.md` types `Type` as `RecordType`, not `uint8`. |
| T-07  | HIGH     | No typed sentinel errors for the validation rules the spec calls out as "typed error" (bad magic, bad version, non-zero reserved, SIGNED flag set, unknown record type, INT length ≠ 8, CRC mismatch, reserved-flag-bits set). Only `ErrInvalid` and `errUnimplemented` exist, and neither is wired up. The spec repeatedly says "return a typed error wrapping the leaf sentinel" — there are no leaf sentinels to wrap. |

---

## 2. Unread fields in `decoder.go`

`decoder.readFile()` is a single line that returns `errUnimplemented`. **Nothing is read.** Every field below is required by the spec but never decoded.

| ID    | Severity | Field / step (offset · size) |
|-------|----------|------------------------------|
| D-01  | BLOCKER  | `Header.Magic` (0 · 4) — must read 4 bytes and compare to `"TLV1"` (`0x54 0x4C 0x56 0x31`). |
| D-02  | BLOCKER  | `Header.Version` (4 · 1) — must read and reject anything ≠ 1 (`SPEC.md` §Versioning). |
| D-03  | BLOCKER  | `Header.Flags` (5 · 1) — must read and validate (see V-04, V-05). |
| D-04  | BLOCKER  | `Header.Reserved` (6 · 2) — must read big-endian and reject ≠ 0 (`structures/header.md` Notes). |
| D-05  | BLOCKER  | `Record.Type` (offset 0 of record · 1) — must read and validate against the enum. |
| D-06  | BLOCKER  | `Record.Length` (offset 1 · 2, big-endian) — must read. |
| D-07  | BLOCKER  | `Record.Value` (offset 3 · `Length`) — must read exactly `Length` bytes. |
| D-08  | BLOCKER  | Record loop — read records "in order until the bytes preceding the trailer are exhausted" (`SPEC.md` §Overview line 14). The decoder has no record-loop scaffold. There is also no length prefix telling the decoder where the trailer starts; the implementation will need to buffer or peek. The audit flags this as a design decision the implementer must resolve. |
| D-09  | BLOCKER  | `Trailer.CRC32` (4 bytes, big-endian) — never read. |

---

## 3. Unwritten fields in `encoder.go`

`encoder.writeFile()` is a single line that returns `errUnimplemented`. **Nothing is written.** Same field list as the decoder.

| ID    | Severity | Field / step |
|-------|----------|--------------|
| E-01  | BLOCKER  | `Header.Magic` — must always write `"TLV1"`. |
| E-02  | BLOCKER  | `Header.Version` — must always write `1`. |
| E-03  | BLOCKER  | `Header.Flags` — must reject reserved bits set (bits 3-7 / mask `0xF8`) and reject the `SIGNED` bit (`structures/header.md` Notes). |
| E-04  | BLOCKER  | `Header.Reserved` — must always write `0` (`structures/header.md` Notes: "must be zero on write"). |
| E-05  | BLOCKER  | `Record.Type` — write 1 byte. |
| E-06  | BLOCKER  | `Record.Length` — write 2 bytes big-endian. Should validate `len(Value) == Length` and `len(Value) ≤ 65535`. |
| E-07  | BLOCKER  | `Record.Value` — write exactly `Length` bytes. |
| E-08  | BLOCKER  | `Trailer.CRC32` — must compute IEEE CRC32 over every byte from offset 0 up to (but not including) the CRC32 itself (`structures/trailer.md` Notes), then write big-endian. |

---

## 4. Missing length / offset / checksum / validation checks

Validation rules called out in the spec that have no code path today.

| ID    | Severity | Rule (source) |
|-------|----------|---------------|
| V-01  | HIGH     | **Magic check** — readers must reject any file whose first 4 bytes are not `"TLV1"`. (`structures/header.md` field table; `SPEC.md` §Header) |
| V-02  | HIGH     | **Version check** — typed error on Version ≠ 1. (`SPEC.md` §Versioning) |
| V-03  | HIGH     | **Reserved == 0 check** — typed error if `Header.Reserved ≠ 0`. (`structures/header.md` Notes; `SPEC.md` §Conditional and Optional Fields) |
| V-04  | HIGH     | **SIGNED flag rejection** — typed error if bit 2 (`0x04`) is set. (`structures/header.md` Notes; `SPEC.md` §Conditional and Optional Fields) |
| V-05  | HIGH     | **Reserved flag bits == 0 check** — typed error if any of bits 3-7 (mask `0xF8`) is set. (`structures/header.md` Flags table) |
| V-06  | HIGH     | **Unknown RecordType rejection** — typed error if `Type` is not in `{0x01, 0x02, 0x03, 0x04}`. (`structures/record.md` Notes; `encoding-tables/record-type.md` Notes; `SPEC.md` §Encoding Tables) |
| V-07  | HIGH     | **INT length == 8 check** — typed error if `Type == INT` and `Length ≠ 8`. (`encoding-tables/record-type.md` Notes) |
| V-08  | HIGH     | **CRC32 verification** — decoder must compute running CRC32 as it reads and compare against the trailer; mismatch returns a typed error wrapping the leaf sentinel. (`structures/trailer.md` Notes; `SPEC.md` §Checksums and Integrity) |
| V-09  | HIGH     | **CRC32 emission** — encoder must compute IEEE CRC32 over header + records and emit it as the trailer. |
| V-10  | HIGH     | **NESTED recursion** — for `Type = NESTED`, the value is itself a complete TLV1 file (header + records + trailer). Decoder/encoder must recurse (or at minimum surface the structure). (`structures/record.md` Notes; `SPEC.md` §Encoding Tables) |
| V-11  | HIGH     | **Trailer-boundary handling** — records are read "until the bytes preceding the trailer are exhausted." The decoder needs a scheme to know where records end and the trailer begins (typical solution: read whole stream, peel last 4 bytes as trailer, decode prefix as header+records). No scaffolding for this exists. |
| V-12  | MEDIUM   | **Zero-record file** — header followed immediately by trailer must be legal. (`SPEC.md` §Conditional and Optional Fields) Round-trip test required. |
| V-13  | MEDIUM   | **Length == 0 record** — empty `Value` must be legal. (`structures/record.md` Notes) Round-trip test required. |
| V-14  | MEDIUM   | **Short read / truncated input** — every read site (header, record header, record value, trailer) must surface a typed error rather than a generic `io.EOF` at the wrong level. The error-chain plumbing exists (`FieldError → OffsetError → leaf`), but no read sites are wired in. |

---

## 5. Encoding-table coverage (`RecordType` enum)

`encoding-tables/record-type.md` defines four values; **none** are represented in code.

| ID    | Severity | Enum value (mask · name · meaning) |
|-------|----------|-----------------------------------|
| ET-01 | HIGH     | `0x01 STRING` — UTF-8 string. No constant; no decode/encode path. The spec does not require validating UTF-8, so the only obligation is to round-trip the bytes faithfully. |
| ET-02 | HIGH     | `0x02 INT` — big-endian signed int64; **must validate `Length == 8`** (V-07). No constant; no validation. |
| ET-03 | HIGH     | `0x03 BLOB` — opaque bytes. No constant. |
| ET-04 | HIGH     | `0x04 NESTED` — recursive TLV1 file. No constant; no recursive decode/encode (V-10). |
| ET-05 | MEDIUM   | No `String()` method on `RecordType` for diagnostics, and no test asserting that constant values match the spec table (the constants table is the source of truth — a test that pins `STRING == 0x01`, `INT == 0x02`, `BLOB == 0x03`, `NESTED == 0x04` will catch silent renumbering). |

For symmetry, the `Header.Flags` bit field is also a small encoding table:

| ID    | Severity | Flag (mask · name) |
|-------|----------|-------------------|
| ET-06 | HIGH     | `0x01 COMPRESSED`, `0x02 ENCRYPTED`, `0x04 SIGNED` constants are missing; reserved mask `0xF8` is missing. (V-04, V-05 cannot be enforced cleanly without these.) |

---

## 6. Round-trip / test coverage

The current test suite has three tests, all asserting the stub error chain. There are **zero** decode tests against real bytes, **zero** encode tests against real bytes, and **zero** round-trip tests.

The spec ships three worked examples (`SPEC.md` §Examples) that should each become a fixture-based round-trip test, plus tests for every validation rule above.

| ID    | Severity | Missing test |
|-------|----------|-------------|
| RT-01 | MEDIUM   | **Round-trip: minimal file** — header + trailer, zero records. (`SPEC.md` example "Minimal: header + trailer") |
| RT-02 | MEDIUM   | **Round-trip: typical file** — one STRING record `"hello"`. (`SPEC.md` example "Typical: one STRING record") |
| RT-03 | MEDIUM   | **Round-trip: COMPRESSED + two records** — STRING + INT, including the INT-length-must-be-8 path. (`SPEC.md` example "Complex: COMPRESSED flag set, two records") |
| RT-04 | MEDIUM   | **Round-trip: every record type** — a file containing one STRING, one INT, one BLOB, one NESTED record. Covers ET-01..ET-04 and the recursion path V-10. |
| RT-05 | MEDIUM   | **Negative tests** — one per validation rule V-01 through V-09 (bad magic, bad version, non-zero reserved, SIGNED set, reserved flag bits set, unknown RecordType, INT with Length ≠ 8, CRC mismatch). Each must assert `errors.Is(err, <leaf>)` and the `FieldError`/`OffsetError` wrapping. |
| RT-06 | MEDIUM   | **Truncation tests** — input cut off inside the header, between header and first record, inside a record header, inside a record value, and inside the trailer. Each must surface a typed error at the right offset (V-14). |

---

## 7. Drift between `SPEC.md` and the chunked spec tree

The chunked spec is the source of truth for several details that `SPEC.md` is looser about. The implementer must follow the chunked tree.

| ID    | Severity | Drift |
|-------|----------|-------|
| DR-01 | LOW      | **`Header.Flags` type name.** `SPEC.md` §Header (line 33) types Flags as `uint8`; `structures/header.md` (line 13) types it as `Flags` (named type). The chunked tree wins — implement a named `Flags` type. |
| DR-02 | LOW      | **`Record.Type` type name.** `SPEC.md` §Record (line 53) types Type as `uint8`; `structures/record.md` (line 11) types it as `RecordType` (named type). The chunked tree wins — implement a named `RecordType` type. |
| DR-03 | LOW (informational) | **NESTED recursion call-out.** `structures/record.md` Notes call out NESTED-as-recursive-TLV1 explicitly; `SPEC.md` §Encoding Tables only mentions it in the table cell. The decoder/encoder design must take the recursive interpretation, not the "opaque bytes" interpretation. |
| DR-04 | LOW (informational) | **INT Length == 8 validation.** `encoding-tables/record-type.md` Notes mandate this validation; `SPEC.md` mentions "Length must be 8" only inside the table cell. Implementer must treat it as a hard validation, not a comment. |

(DR-03 and DR-04 are not contradictions — the chunked spec strengthens what `SPEC.md` says. They're tagged as drift so the implementer doesn't accidentally pick the weaker reading.)

---

## Suggested order of implementation

1. Land the type system (T-01..T-07): `Header`, `Record`, `Trailer`, `File`, `Flags`, `RecordType`, plus leaf sentinel errors.
2. Add the constant-table tests (ET-05) so renumbering is caught immediately.
3. Implement encoder first (E-01..E-08, V-09) — encoder is simpler because there's no input to validate; it just needs CRC and length bookkeeping. Drives the round-trip fixtures.
4. Implement decoder (D-01..D-09) with all validation rules (V-01..V-08, V-10, V-11) wired through `wrapErr`.
5. Add round-trip tests (RT-01..RT-04) and negative tests (RT-05..RT-06, V-12..V-14).
6. Run `go test -race ./...` to confirm.

---

## Out of scope (intentionally not flagged)

- The mechanics of zlib / AES indicated by `COMPRESSED` / `ENCRYPTED`. Both flags only describe how `Record.Value` payloads were *produced*; the spec does not require the codec to inflate or decrypt. Implement as bit-level pass-through unless the implementer's PR adds a separate spec for the payload codecs.
- Streaming decode (the spec lets the decoder buffer; no streaming requirement).
- The `SPEC.md` worked-example CRC32 values (line 107 explicitly says "illustrative; recompute when wiring up tests").
