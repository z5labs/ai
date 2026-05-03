---
name: review-file-library
description: Audit an existing Go file library package against its SPEC.md, reporting missing coverage, structural gaps, missing round-trip tests, and drift between spec and implementation. Use whenever the user wants to "review", "audit", "check", or "verify" a Go file library — phrases like "audit my package against the spec", "what coverage is `pkg/foo` missing", "does this package still match SPEC.md after the spec changed", or "find drift between spec and code". Works for both text packages (`tokenizer.go`/`parser.go`/`printer.go`) and binary packages (`types.go`/`decoder.go`/`encoder.go`); for binary, walks the full `structures/` + `encoding-tables/` tree, not just `SPEC.md`. Skip when the user wants to add features (use `implement-go-text-file-library` or `implement-go-binary-file-library`) or scaffold a new package (use `new-go-text-file-library` or `new-go-binary-file-library`).
---

You are a read-only auditor that compares an existing Go file library package to its `SPEC.md` and reports findings as a structured markdown summary at `<package>/AUDIT.md`. You never edit source, tests, or the spec — only read them and write the audit. You never load the full `SPEC.md` into your own context: large specs would crowd out the cross-referencing work. Instead, grep `SPEC.md` for section line ranges and hand each phase subagent a `(path, offset, limit)` slice it can read directly. For binary packages, the chunked spec tree (`structures/*.md`, `encoding-tables/*.md`) is passed verbatim — those files are already chunked.

Read `references/text-checklist.md` (text packages) or `references/binary-checklist.md` (binary packages) for the exact audit categories each phase subagent must cover and the finding-line format they must emit. The orchestrator concatenates per-phase findings into the final report; if the subagent prompt is right, the parts are already consistent.

## Inputs

- **Package path** (required) — the Go package directory to audit (e.g. "audit `pkg/kvr`"). Source: user prompt. Validate by listing the directory and detecting the pipeline shape (see `## Detect package shape`).
- **`<package>/SPEC.md`** (required) — Source: filesystem. Sliced by line range per phase. **If `SPEC.md` is missing, stop and tell the user to run `extract-text-spec` or `extract-binary-spec` first.** A spec-less audit cannot detect drift or missing coverage — every finding is spec-driven — so partial value would be misleading.
- **`<package>/structures/*.md`, `<package>/encoding-tables/*.md`** (binary packages only, optional) — Source: filesystem. Per-structure and per-table chunked spec produced by `extract-binary-spec`. When present, listed with `Glob` and the file paths passed verbatim to the relevant phase subagent. When absent, the audit proceeds against `SPEC.md` alone but flags the encoding-table coverage category as "skipped — no encoding-tables/" in the report.

## Outputs

- **`<package>/AUDIT.md`** — the durable audit artifact. Overwritten on re-run (the audit is a snapshot, not an append log). Structure documented in `## Report structure` below.
- **Scratch files** `<package>/_audit_<phase>.md` (one per phase) — written by phase subagents, concatenated into `AUDIT.md`, deleted in Cleanup. If a previous run was interrupted and left any behind, delete them before launching subagents so a stale partial finding cannot leak into the new report.
- **Side effect**: runs `(cd <package> && go test -race ./...)` once at the start to capture test status for the report. Failing tests are usually the cheapest pointer at drift — a test that pins spec behavior and now fails is direct evidence of a spec/code mismatch — so the test summary goes at the top of `AUDIT.md` and is passed to every phase subagent. The `cd` is required — this repo has no root `go.mod`.

## Detect package shape

List the package and check which pipeline files exist:

- If `tokenizer.go`, `parser.go`, and `printer.go` all exist → **text package**, phases are `tokenizer`, `parser`, `printer`.
- If `types.go`, `decoder.go`, and `encoder.go` all exist → **binary package**, phases are `types`, `decoder`, `encoder`.
- If neither set is complete → stop and tell the user the package is not a recognized file-library shape (the user may have meant a different directory, or the package was never scaffolded — point them at `new-go-text-file-library` or `new-go-binary-file-library`).
- If both sets coexist somehow → stop and ask the user which pipeline to audit; this skill audits one pipeline shape per run.

Test files (`*_test.go` siblings) are required for the round-trip test coverage category — if any are missing, note it in the test-status header and continue (subagents will report "no tests for X" findings).

## Before you start

1. Read the package's `CLAUDE.md` (if present) — the audit should reflect package-specific conventions documented there.
2. Detect package shape per `## Detect package shape`.
3. Confirm `<package>/SPEC.md` exists; refuse if not (see `## Inputs`).
4. For binary packages, list `<package>/structures/` and `<package>/encoding-tables/` if present; record the file paths for the phase subagents.
5. Run `(cd <package> && go test -race ./...)` and capture: pass/fail status, count of failing tests, and the first ~10 lines of any failure output. This is the test-status header for `AUDIT.md` and the test context every phase subagent receives.
6. Delete any stale `_audit_*.md` scratch files in the package directory.

## Partition SPEC.md by line range (do not read the whole file)

```
grep -n '^## ' <package>/SPEC.md       # section headings + line numbers
wc -l <package>/SPEC.md                # last-line marker for the final section
```

Build a `(section, line_start, line_end)` table from that output. Each section ends one line before the next `## ` heading; the final section ends at `wc -l`.

**Text packages** — same partition as `implement-go-text-file-library`:

| Phase     | Sections to slice                                                      |
|-----------|------------------------------------------------------------------------|
| tokenizer | Overview, Lexical Elements (Tokens) and all its subsections, Examples  |
| parser    | Overview, Structure (Grammar), Semantics, Examples                     |
| printer   | Overview, Structure (Grammar), Semantics, Examples                     |

**Binary packages** — same partition as `implement-go-binary-file-library`:

| Phase   | Sections to slice                                                                                          |
|---------|------------------------------------------------------------------------------------------------------------|
| types   | Overview, Conventions, Field Definitions, Encoding Tables, Versioning                                      |
| decoder | Overview, Conventions, Field Definitions, Encoding Tables, Conditional/Optional Fields, Checksums, Padding, Examples |
| encoder | Overview, Conventions, Field Definitions, Encoding Tables, Checksums, Padding, Examples                    |

Always include `Overview` for every phase (frames what the section is for). Text always includes `Examples` (the cheapest sanity check on user-facing behavior). Binary always includes `Conventions` (byte order is load-bearing for all three phases).

For binary packages, every phase subagent additionally receives the full `structures/*.md` and `encoding-tables/*.md` file lists — those files are already chunked, no slicing needed. The mapping mirrors `implement-go-binary-file-library`: all three binary phases consume both directories.

## Phase order

Phases are independent for an audit (no per-phase context summary is forwarded — unlike implementation, the audit is read-only and each phase reports against the spec, not against the previous phase's output). **If you have an `Agent` / `Task` tool available, dispatch all phase subagents in parallel** — independent reads, independent writes, no coordination needed. **If you don't, run each phase inline yourself in the order listed above.**

### Each phase subagent gets

- The package path and the test-status summary (so findings can reference failing test names).
- The slice list: `<spec_path> offset=<line_start> limit=<line_end - line_start + 1>` for every section in this phase's row of the partition table above.
- For binary packages: the full list of `structures/*.md` and `encoding-tables/*.md` paths.
- Source paths for this phase (text: `tokenizer.go` + `tokenizer_test.go` for the tokenizer phase, etc.; binary: `types.go` + `types_test.go` for the types phase, etc.).
- The package's `CLAUDE.md` path (if present) — package conventions affect what counts as drift.
- Inline pointer to the relevant section of `references/text-checklist.md` or `references/binary-checklist.md` — this defines the audit categories and the finding-line format.
- Output path: `<package>/_audit_<phase>.md`.

The subagent reads its slices via `Read(path, offset, limit)`, reads source files in full (audit needs the whole file to find what's missing), reads its checklist section, then writes `_audit_<phase>.md` in the format documented there. The subagent does not edit source, does not run tests, and does not modify the spec.

## Report structure

After all phase subagents complete, assemble `<package>/AUDIT.md` in this order:

```
# Audit: <package> (<text|binary> file library)

**Date:** <YYYY-MM-DD>
**Spec:** SPEC.md (<N> lines)[, structures/ (<N> files), encoding-tables/ (<N> files)]
**Tests:** <PASS | FAIL — N failing>

<if FAIL: a fenced block with the first ~10 lines of failing-test output>

## Summary

- <total findings count> findings across <N> categories
- Phases: tokenizer (N), parser (N), printer (N)   ← or types/decoder/encoder for binary
- Severity: blockers (N), warnings (N), info (N)

<concatenated _audit_<phase>.md contents, one phase per ## section, in phase order>
```

Concatenate without re-reading section bodies — you wrote the test-status header and the summary, the phase subagents wrote the body sections. The summary counts come from grepping each `_audit_<phase>.md` for finding lines (the checklist mandates a stable bullet prefix per severity, so a `grep -c` per severity is enough).

## Cleanup

Delete `_audit_<phase>.md` scratch files in the package. **Keep `AUDIT.md`** — it's the durable artifact for this run, intended to be diffed against future audits and committed if the user wants a record of when each finding was first observed.

## Why this shape

Audit findings are spec-driven: every "missing X" or "drift between spec and Y" finding cites a spec section. Slicing `SPEC.md` by line range keeps the spec authoritative — no scratch copies, no rewriting, no out-of-sync excerpts. Each phase subagent loads only the bytes its phase needs, so a 50-page format spec costs the orchestrator one `grep` and a small table, not 50 pages of context. Running `go test -race` once up front is cheap and produces direct drift evidence — a failing test that pins spec behavior is more reliable than any grep-based heuristic.
