---
name: implement-go-binary-file-library
description: Implement features for Go binary file library packages that follow a types/decoder/encoder pipeline. Use whenever the user wants to add struct types, decoder methods, encoder methods, bit-field handling, or checksum/integrity logic to a Go package built around `types.go`, `decoder.go`, and `encoder.go` — including phrases like "implement the X header", "add support for Y in the gzip package", "decode the Z field per SPEC.md", or "wire up the FLG bit field", even if the user doesn't say "binary". Skip when the user wants to scaffold a brand-new package (use `new-go-binary-file-library` instead) or when the target package uses the text `tokenizer.go`/`parser.go`/`printer.go` layout (use `implement-go-text-file-library` instead).
---

You are an orchestrator that adds features to an existing Go binary file library package. You prepare context, then delegate each pipeline phase (types → decoder → encoder) to a focused subagent. You never read the full `SPEC.md` into your own context — large specs would crowd out orchestration. Instead, grep `SPEC.md` for section line ranges and hand each subagent a `(path, offset, limit)` slice it can read directly.

Read `references/architecture.md` for the types/decoder/encoder patterns each subagent must follow (especially the `FieldError → OffsetError → leaf` chain, which all error sites funnel through).
Read `references/testing.md` for binary-specific test conventions before launching any subagent.

## Inputs

- **Package path** (required) — the Go package directory the user wants changed (e.g. "implement the FLG bit field in `pkg/gzip`"). Source: user prompt. Validate by listing the directory; if `types.go`, `decoder.go`, `encoder.go`, or any of their `_test.go` siblings are missing, stop and direct the user to `new-go-binary-file-library`.
- **`<package>/SPEC.md`** (optional) — Source: filesystem. When present, sliced by line range per phase. When the path is missing, continue without it — the user's request plus existing source files are the only context. When the path exists but is unreadable (e.g. permissions error), stop and ask the user to fix the path or permissions before continuing.
- **`<package>/structures/*.md`, `<package>/encoding-tables/*.md`** (optional) — Source: filesystem. Pre-chunked spec files produced by `extract-binary-spec`; when present, passed to subagents verbatim, no slicing. When no matching files exist, continue without them. If the `structures/` or `encoding-tables/` directory exists but cannot be read/listed, stop and ask the user to fix the path or permissions before continuing. When a matched path exists but is unreadable, stop and ask the user to fix the path or permissions before continuing.

## Outputs

- **Edits** to `<package>/types.go`, `<package>/types_test.go`, `<package>/decoder.go`, `<package>/decoder_test.go`, `<package>/encoder.go`, `<package>/encoder_test.go` — amended via `Edit`, never recreated wholesale, so prior implementer work is preserved.
- **Stub-test replacement** — the decoder phase deletes/replaces any scaffold-only `TestDecodeStubReturnsErrUnimplemented`-style test; the encoder phase does the same for the `Encode` counterpart. These tests pin the unimplemented chain and start failing the moment the real public API is wired up, so removing them (rather than short-circuiting `Decode`/`Encode` to keep returning `errUnimplemented`) is the only valid resolution.
- **Scratch files** `<package>/_context_types.md` (after Phase 1) and `<package>/_context_decoder.md` (after Phase 2) — overwritten each run, deleted in Cleanup. If a previous run was interrupted and left either file behind, delete them before launching Phase 1 so a stale partial summary cannot leak into the new run.
- **Side effect**: runs `(cd <package> && go test -race ./...)` between phases to verify each phase before launching the next. The `cd` is required — this repo has no root `go.mod`, so each target package's tests must be run from inside that package.

## Before you start

1. Read the package's `CLAUDE.md` (if present) and the repo-root `CLAUDE.md` for project conventions and license-header style.
2. List the package: confirm `types.go`, `decoder.go`, `encoder.go`, and their `_test.go` siblings exist. If they don't, the user wants the `new-go-binary-file-library` scaffold first — say so and stop.
3. Check for `<package>/SPEC.md`. If absent, the user's request and existing source files are the only context — pass them directly to each subagent and skip the partitioning step.
4. Identify scope: which struct types, decoder methods, and encoder methods will change.
5. **Scope gate.** For each phase, sum the line counts of its slice ranges (use the partition table below; for the chunked-input layout, count files under `structures/` and `encoding-tables/` for the relevant phase). If any phase's slices total **more than 600 lines** OR pull **more than 8 chunked files**, partition that phase **along spec-section boundaries** into sub-units of **≤ 300 sliced lines or ≤ 4 chunked files each**, always carrying `Overview` and `Conventions` in every sub-unit (byte order from `Conventions` is load-bearing for every section). Tell the user the partition plan up front — e.g., "decoder phase slices total 950 lines; running it as 3 sub-units of ~3 sections each" — so they can re-scope before any subagent launches. Sub-units run serially per `## Phase chunking` below; if no phase trips the threshold, run each phase as a single subagent call as described in `## Phase order`.
6. **Note any scaffold-only stub tests** (e.g. `TestDecodeStubReturnsErrUnimplemented`, `TestEncodeStubReturnsErrUnimplemented`). See `## Outputs` for how each phase replaces them.
7. **Check the user prompt against the spec.** If the user's request contradicts something in `SPEC.md` (e.g. the spec says reject a flag that the user wants supported), the user's prompt is the active intent — flag the conflict so they can confirm, then implement what the user asked for.
8. **Re-run safety.** This skill is safe to re-run on the same package — see `## Outputs` for what is edited vs. overwritten vs. deleted.

## Partition SPEC.md by line range (do not read the whole file)

```
grep -n '^## ' <package>/SPEC.md       # section headings + line numbers
wc -l <package>/SPEC.md                # last-line marker for the final section
```

Build a `(section, line_start, line_end)` table from that output. Each section ends one line before the next `## ` heading; the final section ends at `wc -l`. Map sections to phases:

| Phase   | Sections to slice                                                                                          |
|---------|------------------------------------------------------------------------------------------------------------|
| types   | Overview, Conventions, Field Definitions, Encoding Tables, Versioning                                      |
| decoder | Overview, Conventions, Field Definitions, Encoding Tables, Conditional/Optional Fields, Checksums, Padding, Examples |
| encoder | Overview, Conventions, Field Definitions, Encoding Tables, Checksums, Padding, Examples                    |

Always include `Conventions` for every phase — byte order is load-bearing for all three.

If `SPEC.md` is paired with `structures/<name>.md` or `encoding-tables/<name>.md` files (the layout produced by `extract-binary-spec`), pass those file paths verbatim to the relevant phase subagent — they are already chunked, so no slicing is needed. The phase-to-directory mapping mirrors the partition table: **all three phases (types, decoder, encoder) consume both `structures/*.md`** (Field Definitions) **and `encoding-tables/*.md`** (Encoding Tables) — types defines the structs and enums each one represents, decoder reads bytes into them, encoder writes bytes back out, and all three need to know layouts and enum tables. The scope-gate file count in step 5 of `## Before you start` follows the same mapping — for each phase, count files in both `structures/` and `encoding-tables/`.

Before launching subagents, grep the slices for `> **Ambiguity:**` callouts and surface them to the user.

## Context summary format

`_context_types.md` and `_context_decoder.md` exist so the next phase's subagent can rely on a small, deterministic snapshot in place of re-reading the upstream source files. Treat them as machine-readable, not narrative — a later subagent must be able to scan the file top-to-bottom and pick out symbols without parsing prose.

**Strict format.** One symbol per line, signature only. No rationale, no examples, no commentary, no code bodies. The only structure permitted is the `## Section` headings shown below. Inside a struct or interface body, one field/method per line is still "one symbol per line"; that is fine. List items in the same order they appear in the source.

**Hard cap: 400 lines per file.** If the summary you would write exceeds 400 lines, the phase's work-unit was sized too large — that is the whole point of the cap. Do not write a longer summary, do not abbreviate to fit, and do not split the summary across files. Stop, tell the user the request needs to be chunked (e.g., "implement the Header struct first, then come back for Records"), and re-launch the phase with the smaller scope.

### `_context_types.md` shape

```
## Structs
<every exported struct, in declaration order; one struct per block; fields one per line, signature only>

## Enums
<for each enum: the type declaration on its own line, then constants one per line, in declaration order>

## Bit-field constants
<for each flag type: the type declaration on its own line, then mask constants one per line>

## Errors
<every exported sentinel/typed error, one per line>
```

Omit any section that has no entries — do not write empty headings.

### `_context_decoder.md` shape

```
## Decode
func Decode(r io.Reader) (*File, error)

## Byte order
<big-endian | little-endian>

## Exported decode helpers
<every new exported decode method/function, one signature per line, in declaration order>
```

## Phase order

Run phases in order. Do not skip ahead. Each phase passes a small `_context_<phase>.md` summary forward.

**If you have an `Agent` / `Task` tool available, spawn a subagent per phase** — it keeps the orchestrator's context lean. **If you don't, run each phase inline yourself**, in the same order, with the same slicing and the same `_context_<phase>.md` summaries between phases. **When the scope gate (step 5 of `## Before you start`) has partitioned a phase, run that phase's sub-units serially per `## Phase chunking` instead of as a single call** — the per-phase descriptions below describe the unpartitioned shape, and `## Phase chunking` explains how a sub-unit varies from it. The discipline (test-first, errors via `wrapErr`, only the slices the phase needs) matters more than who executes the work.

### Phase 1 — types

Spawn a subagent with:
- The slice list: `<spec_path> offset=<line_start> limit=<line_end - line_start + 1>` for every types section above (and any peer `structures/*.md` / `encoding-tables/*.md` paths).
- Source paths: `<package>/types.go`, `<package>/types_test.go`.
- Inline pointers to the **Types** section of `references/architecture.md` and `references/testing.md`.
- A clear description of what struct types, enums, and constants to add or change.

Subagent must read its slices via `Read(path, offset, limit)`, write tests first (size checks via `binary.Size()`, `String()` round-trip for enums, error-chain assertions), confirm tests fail, implement types (every enum gets a `String()` method — non-negotiable; the first hex-dump test failure pays for it), then confirm `go test -race ./...` passes.

When the subagent returns, run `go test -race ./...` yourself, then write `_context_types.md` in the strict format from the [Context summary format](#context-summary-format) section. Honor the 400-line cap; if the summary would exceed it, stop and ask the user to chunk the request before relaunching this phase.

### Phase 2 — decoder

Spawn a subagent with:
- Path slices for the decoder sections above.
- `_context_types.md`.
- Source paths: `<package>/decoder.go`, `<package>/decoder_test.go`.
- Pointers to the **Decoder** section of `references/architecture.md` and `references/testing.md`.
- A description of what `readX` methods to add or change.

Subagent must write decode tests first using hex byte literals + `bytes.NewReader`, including failure-path tests that assert the `FieldError → OffsetError → leaf` chain via `errors.Is`/`errors.As`. Every error site funnels through `d.wrapErr`. Each new structure gets its own `readX` method — don't inline record/header reading inside `readFile` even when it would compile, since it makes per-structure failure tests harder to target.

When the subagent returns, run tests yourself, then write `_context_decoder.md` in the strict format from the [Context summary format](#context-summary-format) section. Honor the 400-line cap; if the summary would exceed it, stop and ask the user to chunk the request before relaunching this phase.

### Phase 3 — encoder

Spawn a subagent with:
- Path slices for the encoder sections above.
- `_context_types.md` and `_context_decoder.md`.
- Source paths: `<package>/encoder.go`, `<package>/encoder_test.go`.
- Pointers to the **Encoder** section of `references/architecture.md` and `references/testing.md`.
- A description of what `writeX` methods to add or change.

Subagent must write encode tests first (struct in → bytes out) **plus a round-trip test** (`Encode → Decode → require.Equal`) for every new method. Each new structure gets its own `writeX` method — symmetric to the decoder. Every error site funnels through `e.wrapErr`. When the subagent returns, run `go test -race ./...` yourself for final verification.

### Cleanup

Delete `_context_types.md` and `_context_decoder.md`. Don't leave scratch files in the package.

## Phase chunking

When the scope gate (step 5 of `## Before you start`) has partitioned a phase into N sub-units, run those sub-units **serially** — they all `Edit` the same `types.go` / `decoder.go` / `encoder.go` file, and parallel sub-calls would race each other's edits. The 600-line / 8-chunk threshold is sized so each sub-unit's incremental output stays under the existing 400-line `_context_types.md` / `_context_decoder.md` cap; partitioning is the up-front move that prevents the cap from being hit mid-phase.

Sub-call `i` in a partitioned phase is briefed exactly like the un-partitioned phase (per `### Phase N` above), with three differences:

1. **Narrower slice list.** Only the `(path, offset, limit)` rows for sections this sub-unit covers — plus `Overview` and `Conventions` (always carried) so the high-level shape and byte-order rule stay in view for every sub-call.
2. **Append, don't overwrite, the running summary.** When `i == 1`, `_context_types.md` / `_context_decoder.md` does not yet exist; the sub-call writes it under the strict format from `## Context summary format`. When `i > 1`, the file already holds symbols added by sub-calls 1..i-1; the sub-call **reads** it (capped at 400 lines, so cheap) and **appends** its own new symbols at the end of the relevant `## Section`. Don't duplicate headings; don't reorder existing entries. To keep this append-only protocol consistent with source declaration order, **each sub-unit's new symbols must be added at the end of the relevant source blocks** in `types.go` / `decoder.go` / `encoder.go` — new struct types go after the existing ones, new enum constants go at the end of their const block, new `readX` / `writeX` methods go after the existing ones. If a later sub-unit logically needs a symbol inserted earlier in source, the partitioning was wrong; re-scope the sub-units so each only adds at the end.
3. **No full-`Read` of the growing source file.** Sub-calls `i > 1` treat the running `_context_*.md` as the cross-reference of record for what symbols already exist in `types.go` / `decoder.go` / `encoder.go`. `Edit` adds new symbols without a fresh whole-file read; if a specific helper needs inspection (e.g. to mirror an existing `readX` / `writeX` pattern), `Read` it with `offset` / `limit`, never the whole file. This is the whole point of the gate — once a phase has appended hundreds of lines to its source, re-reading that source in the next sub-call would crowd out the spec slice the sub-call is here for.

After all sub-units complete, the merged `_context_types.md` / `_context_decoder.md` is what the next phase consumes. The 400-line cap still applies to the merged file; **check the cap *before* appending each sub-unit's symbols — sum the existing merged-file line count plus the lines about to be added, and if the total would exceed 400, stop without appending**. The sub-unit cap was sized too generously; re-partition with smaller sub-units, or ask the user to chunk the request further. This pre-append check matches the hard-cap rule in `## Context summary format` — the merged summary never exists in a state over 400 lines.

## Why this shape

The `(path, offset, limit)` form keeps the spec authoritative — no copying, no rewriting, no scratch `_spec_*.md` files to drift out of sync. Each subagent loads only the bytes its phase needs, so a 50-page format spec costs the orchestrator one `grep` and a small table, not 50 pages of context.
