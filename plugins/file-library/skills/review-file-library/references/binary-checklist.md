# Binary package audit checklist

Each phase subagent emits findings to `<package>/_audit_<phase>.md` using the format and categories below. The orchestrator concatenates these into `AUDIT.md` without further editing, so the headings and finding-line shape here are the durable contract.

## Finding-line format (all phases)

Every finding is one bullet. The bullet starts with a **severity prefix** so the orchestrator can grep counts:

- `- **[blocker]**` — the spec contract is not satisfied. This covers two situations: (a) spec mandates X and the implementation does not have X; (b) the implementation has X but no test pins X against the spec. **Untested behavior is treated as unimplemented** — an encoder method with no round-trip test, a decoder method with no failure-path test asserting the error chain, or a field with no test that exercises it, is a `[blocker]`. The package's tests are how spec compliance is verified; an unverified claim is no claim at all, and the audit refuses to grant it lower severity just because the source compiles.
- `- **[warning]**` — implementation handles X but in a way that differs from the spec at the *behavior* level (i.e., real drift), OR a quality issue that doesn't break the spec contract (e.g., an enum lacks a `String()` method — output is correct, but failure messages are harder to read).
- `- **[info]**` — observation worth recording but not necessarily a defect (e.g., implementation supports more than the spec requires; consider documenting the extension).

After the prefix, cite both sides of the comparison so the reader can jump straight to either:

```
- **[blocker]** structures/header.md § Field table — `Header.Length` field defined in spec but not read by `readHeader` in `decoder.go`
- **[warning]** `encoder.go writeChunk` (line 88) — encode test exists, no round-trip test in `encoder_test.go`
- **[blocker]** structures/trailer.md — CRC field defined; `decoder.go readTrailer` reads bytes but does not validate the checksum
- **[info]** `types.go` defines a `Compression` enum value (`CompressionLZ77`) not listed in encoding-tables/compression.md — verify against vendor docs
```

If a finding spans multiple categories (e.g., a missing struct also breaks the decoder), cite it under the first category and add `(see also: <category>)` rather than duplicating.

## Per-phase output skeleton

Each `_audit_<phase>.md` file uses this exact skeleton — empty categories must still appear with a single bullet `- (none)` so the orchestrator's per-category grep is reliable:

```
## <Phase> findings

### <Category 1>
- ...

### <Category 2>
- ...
```

The orchestrator does not re-order or edit these — the headings here are what the reader sees in `AUDIT.md`.

---

## Types phase

**Source files to read:** `types.go`, `types_test.go`.
**Spec sections received:** Overview, Conventions, Field Definitions, Encoding Tables, Versioning.
**Chunked spec received (if present):** `structures/*.md`, `encoding-tables/*.md` — every file path, both directories.

### Categories

#### Missing struct types
Cross-reference every file in `structures/` (and every entry in the SPEC.md structure index when no `structures/` directory exists) against the exported struct declarations in `types.go`. Each `structures/<name>.md` should correspond to a Go struct named after the file (PascalCase). Missing structs are blockers — the rest of the pipeline cannot represent inputs the spec defines.

Also verify each struct's field list matches the spec's field table for byte order, type, and ordering. A struct that exists but is missing a field is a `[blocker]` under this category, not a separate "drift" finding — both halves are caught at once.

For enums declared in the spec (typically inside `## Versioning` or as a header field's allowed values), confirm a corresponding Go typed-integer with the right constants exists, **and** that it has a `String()` method — the implement skill mandates `String()` on every enum for hex-dump test failure messages, so missing `String()` is a `[warning]`.

#### Encoding-table coverage
For every file in `encoding-tables/` (and every table defined inline in `SPEC.md` when no `encoding-tables/` directory exists), confirm `types.go` represents the table as either a typed enum with one constant per row, or as a documented lookup map. Tables with no Go representation are blockers — the decoder/encoder cannot translate the bytes without them.

If `encoding-tables/` is absent and SPEC.md does not define any inline tables, this category is `- (skipped — no encoding-tables defined in spec)`.

#### Drift
Struct types in `types.go` that are not represented in `structures/` or in the SPEC.md structure index — likely a leftover from a removed format version, or an undocumented extension. Flag as `[warning]` so the user can decide whether to delete the type or update the spec.

Field types that disagree with the spec's field table (e.g., spec says `uint32`, struct uses `uint16`) are `[blocker]` drift — a wrong type breaks the wire format silently.

---

## Decoder phase

**Source files to read:** `decoder.go`, `decoder_test.go`.
**Spec sections received:** Overview, Conventions, Field Definitions, Encoding Tables, Conditional and Optional Fields, Checksums and Integrity, Padding and Alignment, Examples.
**Chunked spec received (if present):** `structures/*.md`, `encoding-tables/*.md`.

### Categories

#### Unread fields
For every field in every `structures/<name>.md` (or every field in SPEC.md's field tables when no chunked layout), confirm the corresponding `readX` method in `decoder.go` reads bytes into the corresponding struct field. A field defined in the spec but not assigned in the decoder is a `[blocker]` — every defined byte must be consumed.

Walk the decoder method body and check the struct-field assignments; the implement skill mandates one `readX` method per structure, so each method's responsibility is clear-cut.

#### Missing length/offset/checksum checks
Three sub-checks, each producing its own findings:

- **Length checks.** For every field whose spec entry says "length", "size", "count", or implies a count of subsequent records, verify the decoder uses the value to bound a read or loop. A length field that's read but ignored is a `[blocker]` — malformed inputs will overrun the buffer.
- **Offset checks.** For every field whose spec entry is an offset to another structure (common in container formats: PNG chunk offsets, BMP DIB-header offsets, ELF section-header offsets), verify the decoder seeks to that offset before reading the target. An offset that's read but never used is a `[blocker]`.
- **Checksum checks.** For every CRC, Adler-32, MD5, or other integrity field defined in the spec's `## Checksums and Integrity` section (or implied in a structure's field table), verify `decoder.go` validates the value after reading. A checksum that's read but not compared is a `[blocker]` — silent corruption.

If the spec says "checksum is informational, do not validate" (some formats do), the decoder reading-without-validating is correct; flag as `[info]` rather than `[blocker]` and cite the spec line that justifies it.

#### Drift
Decoder methods whose accept/reject behavior differs from the spec — extra bytes consumed, missing optional-field handling, wrong byte order. The Conventions section is the source of truth for byte order; the Conditional and Optional Fields section is the source of truth for which fields are optional and under what condition.

Failing tests that pin spec behavior (decode the spec's example bytes, expect a known struct) are direct drift evidence — cite the failing test name from the test-status header.

If `decoder_test.go` lacks a failure-path test that asserts the `FieldError → OffsetError → leaf` chain via `errors.Is`/`errors.As` for each spec-defined rejection rule (bad magic, wrong version, reserved ≠ 0, unknown record type, CRC mismatch, etc.), flag as `[blocker]` — the implement skill mandates these because the chain is the only way callers can locate decode failures by field path or byte offset, and a missing failure-path test means the rejection rule is unverified. Each missing rejection-rule test is its own finding (don't aggregate into one bullet) so the implementer can tick them off.

---

## Encoder phase

**Source files to read:** `encoder.go`, `encoder_test.go`.
**Spec sections received:** Overview, Conventions, Field Definitions, Encoding Tables, Checksums and Integrity, Padding and Alignment, Examples.
**Chunked spec received (if present):** `structures/*.md`, `encoding-tables/*.md`.

### Categories

#### Unwritten fields
Symmetric to the decoder's "unread fields" check. For every field in every `structures/<name>.md`, confirm the corresponding `writeX` method in `encoder.go` writes the field. A field defined in the spec but not written is a `[blocker]` — the output bytes won't round-trip.

Also confirm length, offset, and checksum fields are *computed* by the encoder, not read from the struct verbatim — a length field that just writes `s.Length` will produce a structurally invalid encoding the moment the caller forgets to set it. The encoder must compute length/offset/checksum from the actual data being written.

#### Round-trip test coverage
For every `writeX` method in `encoder.go`, confirm `encoder_test.go` has at least one round-trip test — `Encode → Decode → require.Equal` — that exercises that method. **Missing round-trip tests are blockers.** The implement skill mandates them for every encoder method because direct tests pin output bytes but cannot detect encoder/decoder asymmetry — a package can pass every direct test on both sides while failing round-trip if encoder and decoder share a bug, and that mutual error is exactly the failure mode round-trip tests exist to catch. Treating the gap as a warning underplays it.

If only direct tests exist (struct in → expected bytes out, no round-trip), flag the specific method as `[blocker] no round-trip test`. If the round-trip exists but only for a trivial case (zero-value struct, single-field record, no flags set), flag as `[blocker] round-trip exists but does not exercise <field>` — a partial round-trip leaves the unexercised field in the same untested state as no test at all.

#### Drift
Encoder methods whose output bytes differ from what the spec's Examples section shows — wrong padding, wrong byte order in a sub-field, wrong checksum algorithm. Round-trip tests catch most of this directly; this category exists for cases where the round-trip passes (because the decoder has the same bug) but the output doesn't match the spec's golden bytes.

If the package has any test that loads the spec's Examples bytes and decodes-then-encodes-back, comparing to the original bytes, that's the gold-standard drift detector — flag missing such tests as `[blocker] no Examples-bytes round-trip test`. The Examples section exists to be tested against; not testing against it is leaving the spec's own gold-standard fixture on the table.

---

## Test-status integration

The orchestrator passes the result of `go test -race ./...` to every phase subagent. If tests are failing:

- The first ~10 lines of failure output is in the test-status header at the top of `AUDIT.md`.
- Each phase subagent must scan failing test names for tests that belong to its phase (e.g., the decoder phase scans for `TestDecode*`, `TestRead*`) and add a `[blocker]` finding under the relevant category referencing the failing test by name. A failing test is direct, runtime-verified drift evidence — the cheapest finding the audit produces.

If tests pass, no test-failure findings are added; the test-status header in `AUDIT.md` is the only mention.

## What the audit does not do

- Does not propose fixes — findings cite the gap, not the patch.
- Does not edit the spec, source, or tests.
- Does not run benchmarks, fuzz tests, or coverage tools.
- Does not score the package — every finding stands on its own; no aggregate "grade".
