---
name: code-review-checklist
description: Walk a single PR diff through a four-phase review checklist (correctness, tests, docs, style) and produce a markdown review report. Use when the user asks for a structured review on a specific diff or patch file. Skip when the user wants a security audit, a performance review, or feedback on more than one PR at a time.
argument-hint: "<path-to-diff>"
---

# code-review-checklist

A structured review pass for a single PR diff. The skill reads the diff once into main context, then walks four review phases (correctness, tests, docs, style) over that same content. Each phase appends one section to the report file. The whole workflow runs in the main thread — no parallel work, no extra processes.

## Idempotency

This skill is idempotent — re-running on the same diff overwrites the previous review report. The diff itself is never modified.

## Inputs

- `diff_path` (required, positional CLI arg) — absolute or repo-relative path to a unified-diff file (the output of `git diff` or `git format-patch`). Validation: must exist and be a regular file (`[ -f "$diff_path" ]`); reject if missing.
- `report_path` (optional, second positional arg) — destination path for the review report. Default: `./code-review-<basename-of-diff>.md` in the current working directory.

## Outputs

- Review report written to `<report_path>` — overwrites any existing file at that destination. The skill writes the report atomically: each phase appends its section to the file in order, and the Final assembly step prepends the one-line summary at the top after all four phases finish.

## Citation format

Findings cite `<file>:<line>` where `<file>` is taken from the `+++ b/<file>` line of the hunk and `<line>` is computed from the hunk header `@@ -a,b +c,d @@` as `c + (offset of the cited line within the hunk)`. The `+++` line alone never yields a line number; always pair it with the hunk header.

## Workflow

Read `diff_path` once into main context. Reuse that content across all four phases — do not re-read between phases.

### Phase 1: Correctness

For each hunk in the diff, raise a finding only when one of these mechanical triggers is present:
- A new conditional branch is added (`if`, `else if`, `case`) and no matching handler appears in the surrounding code (e.g. an `if` without an `else` for a returned-error path).
- A loop or array index uses `<=` against `length` or `len(...)` (off-by-one trigger).
- A pointer or optional value is dereferenced without a preceding nil/None/null guard in the same hunk.

Write findings to `## Correctness` in the report.

### Phase 2: Tests

For each hunk that touches non-test code (path does not match `*_test.*` or `*test_*`), check whether the diff also adds or modifies a test file in the same package or directory. If not, raise a finding. Phases 1's mechanical triggers carry over: if Phase 1 raised a finding on a hunk and no test was added for that hunk, raise a paired Tests finding.

Write findings to `## Tests` in the report.

### Phase 3: Docs

For each hunk that adds or modifies a symbol whose name starts with an uppercase letter (Go-style export marker; for other languages: any symbol annotated `public`, `export`, or written without a leading `_`), check whether the diff also updates the doc comment immediately preceding that symbol. If the doc comment is unchanged but the symbol's signature or behavior changed, raise a finding.

Write findings to `## Docs` in the report.

### Phase 4: Style

Mechanical checks only — no subjective style judgments. For each hunk, raise a finding when:
- The hunk introduces a line with trailing whitespace.
- The hunk adds an `import` line that is not referenced anywhere else in the same file (use grep against the file's identifiers in the diff context).

Write findings to `## Style` in the report.

### Final assembly

Read the file at `<report_path>` once, count hunk headers (`@@ -`) and `+++ b/` entries from the in-memory diff to compute `<N>` (hunks) and `<M>` (files), and count bullet entries in the report file for `<X>` (total findings). Prepend the line `Reviewed <N> hunks across <M> files. <X> total findings.` to the report file. Print the report path to stdout.
