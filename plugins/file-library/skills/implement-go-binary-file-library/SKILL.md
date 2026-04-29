---
name: implement-go-binary-file-library
description: Implement features for Go binary file library packages that follow a types/decoder/encoder pipeline. Use whenever the user wants to add struct types, decoder methods, encoder methods, bit-field handling, or checksum/integrity logic to a Go package built around `types.go`, `decoder.go`, and `encoder.go` — including phrases like "implement the X header", "add support for Y in the gzip package", "decode the Z field per SPEC.md", or "wire up the FLG bit field", even if the user doesn't say "binary".
---

You are an orchestrator that adds features to an existing Go binary file library package. You prepare context, then delegate each pipeline phase (types → decoder → encoder) to a focused subagent. You never read the full `SPEC.md` into your own context — large specs would crowd out orchestration. Instead, grep `SPEC.md` for section line ranges and hand each subagent a `(path, offset, limit)` slice it can read directly.

Read `references/architecture.md` for the types/decoder/encoder patterns each subagent must follow (especially the `FieldError → OffsetError → leaf` chain, which all error sites funnel through).
Read `references/testing.md` for binary-specific test conventions before launching any subagent.

## Before you start

1. Read the package's `CLAUDE.md` (if present) and the repo-root `CLAUDE.md` for project conventions and license-header style.
2. List the package: confirm `types.go`, `decoder.go`, `encoder.go`, and their `_test.go` siblings exist. If they don't, the user wants the `new-go-binary-file-library` scaffold first — say so and stop.
3. Check for `<package>/SPEC.md`. If absent, the user's request and existing source files are the only context — pass them directly to each subagent and skip the partitioning step.
4. Identify scope: which struct types, decoder methods, and encoder methods will change.
5. **Note any scaffold-only stub tests** (e.g. `TestDecodeStubReturnsErrUnimplemented`, `TestEncodeStubReturnsErrUnimplemented`). These pin the unimplemented chain and will start failing the moment you wire up the real public API. The decoder phase deletes/replaces the `Decode` stub test; the encoder phase deletes/replaces the `Encode` stub test. Don't leave them green by short-circuiting `Decode`/`Encode` to keep returning `errUnimplemented`.
6. **Check the user prompt against the spec.** If the user's request contradicts something in `SPEC.md` (e.g. the spec says reject a flag that the user wants supported), the user's prompt is the active intent — flag the conflict so they can confirm, then implement what the user asked for.

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

If `SPEC.md` is paired with `structures/<name>.md` or `encoding-tables/<name>.md` files (the layout produced by `extract-binary-spec`), pass those file paths verbatim to the relevant phase subagent — they are already chunked, so no slicing is needed.

Before launching subagents, grep the slices for `> **Ambiguity:**` callouts and surface them to the user.

## Phase order

Run phases in order. Do not skip ahead. Each phase passes a small `_context_<phase>.md` summary forward.

**If you have an `Agent` / `Task` tool available, spawn a subagent per phase** — it keeps the orchestrator's context lean. **If you don't, run each phase inline yourself**, in the same order, with the same slicing and the same `_context_<phase>.md` summaries between phases. The discipline (test-first, errors via `wrapErr`, only the slices the phase needs) matters more than who executes the work.

### Phase 1 — types

Spawn a subagent with:
- The slice list: `<spec_path> offset=<line_start> limit=<line_end - line_start + 1>` for every types section above (and any peer `structures/*.md` / `encoding-tables/*.md` paths).
- Source paths: `<package>/types.go`, `<package>/types_test.go`.
- Inline pointers to the **Types** section of `references/architecture.md` and `references/testing.md`.
- A clear description of what struct types, enums, and constants to add or change.

Subagent must read its slices via `Read(path, offset, limit)`, write tests first (size checks via `binary.Size()`, `String()` round-trip for enums, error-chain assertions), confirm tests fail, implement types (every enum gets a `String()` method — non-negotiable; the first hex-dump test failure pays for it), then confirm `go test -race ./...` passes.

When the subagent returns, run `go test -race ./...` yourself and write `_context_types.md` listing every exported type, enum value, and bit-field constant the next phases will need.

### Phase 2 — decoder

Spawn a subagent with:
- Path slices for the decoder sections above.
- `_context_types.md`.
- Source paths: `<package>/decoder.go`, `<package>/decoder_test.go`.
- Pointers to the **Decoder** section of `references/architecture.md` and `references/testing.md`.
- A description of what `readX` methods to add or change.

Subagent must write decode tests first using hex byte literals + `bytes.NewReader`, including failure-path tests that assert the `FieldError → OffsetError → leaf` chain via `errors.Is`/`errors.As`. Every error site funnels through `d.wrapErr`. Each new structure gets its own `readX` method — don't inline record/header reading inside `readFile` even when it would compile, since it makes per-structure failure tests harder to target.

When the subagent returns, run tests yourself and write `_context_decoder.md` capturing the public `Decode` signature, the byte order, and the names of any new exported decode helpers.

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

## Why this shape

The `(path, offset, limit)` form keeps the spec authoritative — no copying, no rewriting, no scratch `_spec_*.md` files to drift out of sync. Each subagent loads only the bytes its phase needs, so a 50-page format spec costs the orchestrator one `grep` and a small table, not 50 pages of context.
