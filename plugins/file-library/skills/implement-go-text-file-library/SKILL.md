---
name: implement-go-text-file-library
description: Implement features for Go text file library packages that follow a tokenizer/parser/printer pipeline. Use whenever the user wants to add token types, parser rules, AST nodes, or printer logic to a Go package built around `tokenizer.go`, `parser.go`, and `printer.go` — including phrases like "tokenize the X keyword", "parse the Y block", "format the Z node", "add support for comments", or "implement the spec section on records", even if the user doesn't say "text".
---

You are an orchestrator that adds features to an existing Go text file library package. You prepare context, then delegate each pipeline phase (tokenizer → parser → printer) to a focused subagent. You never read the full `SPEC.md` into your own context — large specs would crowd out orchestration. Instead, grep `SPEC.md` for section line ranges and hand each subagent a `(path, offset, limit)` slice it can read directly.

Read `references/architecture.md` for the tokenizer/parser/printer patterns each subagent must follow (especially the **inner action loop** rule for complex types — flat for-loops with switches do not scale and must be rejected in review).
Read `references/testing.md` for text-specific test conventions before launching any subagent.

## Before you start

1. Read the package's `CLAUDE.md` (if present) and the repo-root `CLAUDE.md` for project conventions and license-header style.
2. List the package: confirm `tokenizer.go`, `parser.go`, `printer.go`, and their `_test.go` siblings exist. If they don't, the user wants the `new-go-text-file-library` scaffold first — say so and stop.
3. Check for `<package>/SPEC.md`. If absent, the user's request and existing source files are the only context — pass them directly to each subagent and skip the partitioning step.
4. Identify scope: which token types, AST nodes, parser rules, and printer rules will change.
5. **Check the user prompt against the spec.** If the user's request contradicts something in `SPEC.md` (e.g. the spec rejects a syntax the user wants supported), the user's prompt is the active intent — flag the conflict so they can confirm, then implement what the user asked for.

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

If `SPEC.md` is paired with already-chunked `tokens/<name>.md` or `grammar/<name>.md` files (the layout some `extract-text-spec` runs produce), pass those file paths verbatim to the relevant phase subagent — they're already chunked, so no slicing is needed.

Before launching subagents, grep the slices for `> **Ambiguity:**` callouts and surface them to the user.

## Phase order

Run phases in order. Do not skip ahead. Each phase passes a small `_context_<phase>.md` summary forward.

**If you have an `Agent` / `Task` tool available, spawn a subagent per phase** — it keeps the orchestrator's context lean. **If you don't, run each phase inline yourself**, in the same order, with the same slicing and the same `_context_<phase>.md` summaries between phases. The discipline (test-first, exact `Pos` values, inner action loop for complex types, round-trip tests) matters more than who executes the work.

### Phase 1 — tokenizer

Spawn a subagent with:
- The slice list: `<spec_path> offset=<line_start> limit=<line_end - line_start + 1>` for every tokenizer section above (and any peer `tokens/*.md` paths).
- Source paths: `<package>/tokenizer.go`, `<package>/tokenizer_test.go`.
- Inline pointers to the **Tokenizer** section of `references/architecture.md` and `references/testing.md`.
- A clear description of what token types and tokenizing rules to add or change.

Subagent must read its slices via `Read(path, offset, limit)`, write tokenizer tests first (table-driven, exact `Pos{Line, Column}` values, `collect` helper drains the `iter.Seq2`), confirm tests fail for the right reason, implement tokenizer changes following the closure pattern (capture state in returned action functions; never accumulate on the tokenizer struct), then confirm `go test -race ./...` passes.

When the subagent returns, run `go test -race ./...` yourself and write `_context_tokens.md` listing the `TokenType` constants and `Token` struct definition the next phases will need.

### Phase 2 — parser

Spawn a subagent with:
- Path slices for the parser sections above.
- `_context_tokens.md`.
- Source paths: `<package>/parser.go`, `<package>/parser_test.go`.
- Pointers to the **Parser** section of `references/architecture.md` and `references/testing.md`.
- A description of what AST nodes and parser rules to add or change.

Subagent must write parser tests first using the public `Parse()` function with real source strings — never construct AST nodes by hand for expectations (the empty-`File` scaffold case is the only exception). Confirm tests fail for the right reason. Implement parser changes; for any complex type (nested members, repetition, alternation), use the **inner action loop pattern** with one `parserAction[*T]` per state — flat for-loops with switches accrete and become unmaintainable, so this is a hard rule. Use `p.expect(types...)` everywhere the grammar requires a specific token; never inline the type check.

When the subagent returns, run tests yourself and write `_context_ast.md` capturing the `File` struct, the `Type` interface, every concrete AST type that implements it, and the `Parse()` signature.

### Phase 3 — printer

Spawn a subagent with:
- Path slices for the printer sections above.
- `_context_tokens.md` and `_context_ast.md`.
- Source paths: `<package>/printer.go`, `<package>/printer_test.go`.
- Pointers to the **Printer** section of `references/architecture.md` and `references/testing.md`.
- A description of what printer rules to add or change.

Subagent must write printer tests first — both **direct** tests (AST in, expected string out) **and a round-trip test** (`Parse → Print → Parse → require.Equal`) for every new printer method. Round-trip is the cheapest end-to-end correctness check; direct tests pin the formatting choices round-trip can't see. Confirm tests fail, implement printer changes (closure pattern for slice iteration; every write goes through `pr.write` so `pr.err` short-circuits), then run `go test -race ./...` yourself for final verification.

### Cleanup

Delete `_context_tokens.md` and `_context_ast.md`. Don't leave scratch files in the package.

## Why this shape

The `(path, offset, limit)` form keeps the spec authoritative — no copying, no rewriting, no scratch `_spec_*.md` files to drift out of sync. Each subagent loads only the bytes its phase needs, so a 50-page format spec costs the orchestrator one `grep` and a small table, not 50 pages of context.
