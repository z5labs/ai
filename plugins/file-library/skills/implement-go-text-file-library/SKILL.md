---
name: implement-go-text-file-library
description: Implement features for Go text file library packages that follow a tokenizer/parser/printer pipeline. Use whenever the user wants to add token types, parser rules, AST nodes, or printer logic to a Go package built around `tokenizer.go`, `parser.go`, and `printer.go` — including phrases like "tokenize the X keyword", "parse the Y block", "format the Z node", "add support for comments", or "implement the spec section on records", even if the user doesn't say "text". Skip when the user wants to scaffold a brand-new package (use `new-go-text-file-library` instead) or when the target package uses the binary `types.go`/`decoder.go`/`encoder.go` layout (use `implement-go-binary-file-library` instead).
---

You are an orchestrator that adds features to an existing Go text file library package. You prepare context, then delegate each pipeline phase (tokenizer → parser → printer) to a focused subagent. You never read the full `SPEC.md` into your own context — large specs would crowd out orchestration. Instead, grep `SPEC.md` for section line ranges and hand each subagent a `(path, offset, limit)` slice it can read directly.

Read `references/architecture.md` for the tokenizer/parser/printer patterns each subagent must follow (especially the **inner action loop** rule for complex types — flat for-loops with switches do not scale and must be rejected in review).
Read `references/testing.md` for text-specific test conventions before launching any subagent.

## Inputs

- **Package path** (required) — the Go package directory the user wants changed (e.g. "implement comments in `pkg/kvr`"). Source: user prompt. Validate by listing the directory; if `tokenizer.go`, `parser.go`, `printer.go`, or any of their `_test.go` siblings are missing, stop and direct the user to `new-go-text-file-library`.
- **`<package>/SPEC.md`** (optional) — Source: filesystem. When present, sliced by line range per phase. When the path is missing, continue without it — the user's request plus existing source files are the only context. When the path exists but is unreadable (e.g. permissions error), stop and ask the user to fix the path or permissions before continuing.
- **`<package>/tokens/*.md`, `<package>/grammar/*.md`** (optional) — Source: filesystem. Optional pre-chunked spec layout (prepared manually or by an external tool — `extract-text-spec` itself produces a single `SPEC.md`, not these per-section files); when present, passed to subagents verbatim, no slicing. When no matching files exist, continue without them. If the `tokens/` or `grammar/` directory exists but cannot be read/listed, stop and ask the user to fix the path or permissions before continuing. When a matched path exists but is unreadable, stop and ask the user to fix the path or permissions before continuing.

## Outputs

- **Edits** to `<package>/tokenizer.go`, `<package>/tokenizer_test.go`, `<package>/parser.go`, `<package>/parser_test.go`, `<package>/printer.go`, `<package>/printer_test.go` — amended via `Edit`, never recreated wholesale, so prior implementer work is preserved.
- **Scratch files** `<package>/_context_tokens.md` (after Phase 1) and `<package>/_context_ast.md` (after Phase 2) — overwritten each run, deleted in Cleanup. If a previous run was interrupted and left either file behind, delete them before launching Phase 1 so a stale partial summary cannot leak into the new run.
- **Side effect**: runs `(cd <package> && go test -race ./...)` between phases to verify each phase before launching the next. The `cd` is required — this repo has no root `go.mod`, so each target package's tests must be run from inside that package.

## Before you start

1. Read the package's `CLAUDE.md` (if present) and the repo-root `CLAUDE.md` for project conventions and license-header style.
2. List the package: confirm `tokenizer.go`, `parser.go`, `printer.go`, and their `_test.go` siblings exist. If they don't, the user wants the `new-go-text-file-library` scaffold first — say so and stop.
3. Check for `<package>/SPEC.md` and the optional chunked layout (`<package>/tokens/`, `<package>/grammar/`). The four input combinations are:
   - **Both absent** — the user's request and existing source files are the only context. Pass them directly to each subagent and skip the scope gate (step 5); there's nothing to count.
   - **`SPEC.md` present, chunked layout absent** — slice `SPEC.md` per the partition table; the scope gate counts per-phase slice-line totals (the chunked-file count is 0, so only the line trigger can fire).
   - **Both present** — slice `SPEC.md` *and* pass the chunked files verbatim; the scope gate counts both per phase and trips when either threshold is exceeded.
   - **Chunked layout present, `SPEC.md` absent** — not a supported combination. The chunked layout *supplements* `SPEC.md`; the always-carry sections (`Overview`, `Examples`) live only in `SPEC.md`, so without it `## Phase chunking`'s carry rule (sub-units must always include `Overview` and `Examples`) cannot be satisfied. Stop and ask the user to add a `SPEC.md` — even a minimal one containing just those two sections is enough — before proceeding.
4. Identify scope: which token types, AST nodes, parser rules, and printer rules will change.
5. **Scope gate.** For each phase, sum the line counts of its slice ranges (use the partition table below; for the chunked-input layout, count files under `tokens/` and `grammar/` for the relevant phase). If any phase's slices total **more than 600 lines** OR pull **more than 8 chunked files**, partition that phase **along spec-section boundaries** into sub-units of **≤ 300 sliced lines or ≤ 4 chunked files each**, always carrying `Overview` and `Examples` in every sub-unit. Tell the user the partition plan up front — e.g., "Phase 2's slices total 920 lines; running it as 3 sub-units of ~3 sections each" — so they can re-scope before any subagent launches. Sub-units run serially per `## Phase chunking` below; if no phase trips the threshold, run each phase as a single subagent call as described in `## Phase order`.
6. **Check the user prompt against the spec.** If the user's request contradicts something in `SPEC.md` (e.g. the spec rejects a syntax the user wants supported), the user's prompt is the active intent — flag the conflict so they can confirm, then implement what the user asked for.
7. **Re-run safety.** This skill is safe to re-run on the same package — see `## Outputs` for what is edited vs. overwritten vs. deleted.

## Partition SPEC.md by line range (do not read the whole file)

```
grep -n '^## ' <package>/SPEC.md       # section headings + line numbers
wc -l <package>/SPEC.md                # last-line marker for the final section
```

Build a `(section, line_start, line_end)` table from that output. Each section ends one line before the next `## ` heading; the final section ends at `wc -l`. Map sections to phases:

| Phase     | Sections to slice                                                                                           |
|-----------|-------------------------------------------------------------------------------------------------------------|
| tokenizer | Overview, Lexical Elements (Tokens) and all its subsections, Examples                                       |
| parser    | Overview, Structure (Grammar), Semantics, Examples                                                          |
| printer   | Overview, Structure (Grammar), Semantics, Examples                                                          |

Always include `Overview` for every phase — it carries the high-level shape that frames the section the subagent is reading. Always include `Examples` — they're the cheapest sanity check on whether the implementation matches user-facing behavior.

If `SPEC.md` is paired with already-chunked `tokens/<name>.md` or `grammar/<name>.md` files (an optional pre-chunked layout — prepared manually or by an external tool, distinct from `extract-text-spec`'s single-`SPEC.md` output), pass those file paths verbatim to the relevant phase subagent — they're already chunked, so no slicing is needed. The phase-to-directory mapping mirrors the partition table: **tokenizer phase consumes `tokens/*.md`** (Lexical Elements); **parser and printer phases each consume `grammar/*.md`** (Structure (Grammar)). The scope-gate file count in step 5 of `## Before you start` follows the same mapping — count `tokens/*.md` for the tokenizer phase and `grammar/*.md` for the parser and printer phases.

Before launching subagents, grep the slices for `> **Ambiguity:**` callouts and surface them to the user.

## Context summary format

`_context_tokens.md` and `_context_ast.md` exist so the next phase's subagent can rely on a small, deterministic snapshot in place of re-reading the upstream source files. Treat them as machine-readable, not narrative — a later subagent must be able to scan the file top-to-bottom and pick out symbols without parsing prose.

**Strict format.** One symbol per line, signature only. No rationale, no examples, no commentary, no code bodies. The only structure permitted is the `## Section` headings shown below. Inside a struct or interface body, one field/method per line is still "one symbol per line"; that is fine. List items in the same order they appear in the source.

**Hard cap: 400 lines per file.** If the summary you would write exceeds 400 lines, the phase's work-unit was sized too large — that is the whole point of the cap. Do not write a longer summary, do not abbreviate to fit, and do not split the summary across files. Stop, tell the user the request needs to be chunked (e.g., "tokenize comments and identifiers first, then come back for literals"), and re-launch the phase with the smaller scope.

### `_context_tokens.md` shape

```
## TokenType
<every constant from the TokenType const block, one per line, in declaration order>

## Token
type Token struct {
    <one field per line, signature only>
}
```

### `_context_ast.md` shape

```
## Parse
func Parse(r io.Reader) (*File, error)

## File
type File struct {
    <one field per line>
}

## Type
type Type interface {
    <one method signature per line>
}

## AST nodes
<every concrete type that implements Type, in declaration order; one type per block; struct fields one per line>
```

## Phase order

Run phases in order. Do not skip ahead. Phase 1 writes `_context_tokens.md` (the `TokenType` constants and `Token` struct the next phases need); Phase 2 writes `_context_ast.md` (the `File` struct, `Type` interface, concrete AST nodes, and the `Parse()` signature). Phase 3 reads both and writes nothing forward.

**If you have an `Agent` / `Task` tool available, spawn a subagent per phase** — it keeps the orchestrator's context lean. **If you don't, run each phase inline yourself**, in the same order, with the same slicing and the same `_context_tokens.md` / `_context_ast.md` summaries between phases. **When the scope gate (step 5 of `## Before you start`) has partitioned a phase, run that phase's sub-units serially per `## Phase chunking` instead of as a single call** — the per-phase descriptions below describe the unpartitioned shape, and `## Phase chunking` explains how a sub-unit varies from it. The discipline (test-first, exact `Pos` values, inner action loop for complex types, round-trip tests) matters more than who executes the work.

### Phase 1 — tokenizer

Spawn a subagent with:
- The slice list: `<spec_path> offset=<line_start> limit=<line_end - line_start + 1>` for every tokenizer section above (and any peer `tokens/*.md` paths).
- Source paths: `<package>/tokenizer.go`, `<package>/tokenizer_test.go`.
- Inline pointers to the **Tokenizer** section of `references/architecture.md` and `references/testing.md`.
- A clear description of what token types and tokenizing rules to add or change.

Subagent must read its slices via `Read(path, offset, limit)`, write tokenizer tests first (table-driven, exact `Pos{Line, Column}` values, `collect` helper drains the `iter.Seq2`), confirm tests fail for the right reason, implement tokenizer changes following the closure pattern (capture state in returned action functions; never accumulate on the tokenizer struct), then confirm `(cd <package> && go test -race ./...)` passes.

When the subagent returns, run `(cd <package> && go test -race ./...)` yourself, then write `_context_tokens.md` in the strict format from the [Context summary format](#context-summary-format) section. Honor the 400-line cap; if the summary would exceed it, stop and ask the user to chunk the request before relaunching this phase.

### Phase 2 — parser

Spawn a subagent with:
- Path slices for the parser sections above.
- `_context_tokens.md`.
- Source paths: `<package>/parser.go`, `<package>/parser_test.go`.
- Pointers to the **Parser** section of `references/architecture.md` and `references/testing.md`.
- A description of what AST nodes and parser rules to add or change.

Subagent must write parser tests first using the public `Parse()` function with real source strings — never construct AST nodes by hand for expectations (the empty-`File` scaffold case is the only exception). Confirm tests fail for the right reason. Implement parser changes; for any complex type (nested members, repetition, alternation), use the **inner action loop pattern** with one `parserAction[*T]` per state — flat for-loops with switches accrete and become unmaintainable, so this is a hard rule. Use `p.expect(types...)` everywhere the grammar requires a specific token; never inline the type check.

When the subagent returns, run tests yourself, then write `_context_ast.md` in the strict format from the [Context summary format](#context-summary-format) section. Honor the 400-line cap; if the summary would exceed it, stop and ask the user to chunk the request before relaunching this phase.

### Phase 3 — printer

Spawn a subagent with:
- Path slices for the printer sections above.
- `_context_tokens.md` and `_context_ast.md`.
- Source paths: `<package>/printer.go`, `<package>/printer_test.go`.
- Pointers to the **Printer** section of `references/architecture.md` and `references/testing.md`.
- A description of what printer rules to add or change.

Subagent must write printer tests first — both **direct** tests (AST in, expected string out) **and a round-trip test** (`Parse → Print → Parse → require.Equal`) for every new printer method. Round-trip is the cheapest end-to-end correctness check; direct tests pin the formatting choices round-trip can't see. Confirm tests fail, implement printer changes (closure pattern for slice iteration; every write goes through `pr.write` so `pr.err` short-circuits), then run `(cd <package> && go test -race ./...)` yourself for final verification.

### Cleanup

Delete `_context_tokens.md` and `_context_ast.md`. Don't leave scratch files in the package.

## Phase chunking

When the scope gate (step 5 of `## Before you start`) has partitioned a phase into N sub-units, run those sub-units **serially** — they all `Edit` the same `tokenizer.go` / `parser.go` / `printer.go` file, and parallel sub-calls would race each other's edits. The 600-line / 8-chunk threshold is sized so each sub-unit's incremental output stays under the existing 400-line `_context_tokens.md` / `_context_ast.md` cap; partitioning is the up-front move that prevents the cap from being hit mid-phase.

Sub-call `i` in a partitioned phase is briefed exactly like the un-partitioned phase (per `### Phase N` above), with three differences:

1. **Narrower slice list.** Only the `(path, offset, limit)` rows for sections this sub-unit covers — plus `Overview` and `Examples` (always carried) so the high-level shape and user-facing behavior stay in view.
2. **Append, don't overwrite, the running summary.** When `i == 1`, `_context_tokens.md` / `_context_ast.md` does not yet exist; the sub-call writes it under the strict format from `## Context summary format`. When `i > 1`, the file already holds symbols added by sub-calls 1..i-1; the sub-call **reads** it (capped at 400 lines, so cheap) and **appends** its own new symbols at the end of the relevant `## Section`. Don't duplicate headings; don't reorder existing entries. To keep this append-only protocol consistent with source declaration order, **each sub-unit's new symbols must be added at the end of the relevant source blocks** in `tokenizer.go` / `parser.go` / `printer.go` — new `TokenType` constants go at the end of the existing const block, new AST node types go after the existing ones, new printer functions go after the existing ones. If a later sub-unit logically needs a symbol inserted earlier in source, the partitioning was wrong; re-scope the sub-units so each only adds at the end.
3. **No full-`Read` of the growing source file.** Sub-calls `i > 1` treat the running `_context_*.md` as the cross-reference of record for what symbols already exist in `tokenizer.go` / `parser.go` / `printer.go`. `Edit` adds new symbols without a fresh whole-file read; if a specific helper needs inspection (e.g. to mirror an existing pattern), `Read` it with `offset` / `limit`, never the whole file. This is the whole point of the gate — once a phase has appended hundreds of lines to its source, re-reading that source in the next sub-call would crowd out the spec slice the sub-call is here for.

After all sub-units complete, the merged `_context_tokens.md` / `_context_ast.md` is what the next phase consumes. The 400-line cap still applies to the merged file; **check the cap *before* appending each sub-unit's symbols — sum the existing merged-file line count plus the lines about to be added, and if the total would exceed 400, stop without appending**. The sub-unit cap was sized too generously; re-partition with smaller sub-units, or ask the user to chunk the request further. This pre-append check matches the hard-cap rule in `## Context summary format` — the merged summary never exists in a state over 400 lines.

## Why this shape

The `(path, offset, limit)` form keeps the spec authoritative — no copying, no rewriting, no scratch `_spec_*.md` files to drift out of sync. Each subagent loads only the bytes its phase needs, so a 50-page format spec costs the orchestrator one `grep` and a small table, not 50 pages of context.
