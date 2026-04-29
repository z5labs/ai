# Audit Report: `word-count` skill

**Target:** `./.claude/skills/word-count/SKILL.md` (29 lines, 1760 bytes)
**Date:** 2026-04-29
**Auditor:** Claude Code

## Summary

The `word-count` skill is a small, single-purpose skill that wraps `wc -wlc` and writes a sibling `.count.md` report. The SKILL.md is concise, well-scoped, and frontmatter is structurally valid. The skill correctly documents its inputs, outputs, idempotency, and "when to skip" boundary conditions. No supporting scripts/resources exist — the workflow is small enough that a bundled script isn't strictly required, but as written the workflow depends on the model running shell commands correctly.

Overall assessment: **acceptable for its declared scope, with a handful of correctness and clarity issues** that should be addressed before relying on it in production.

## Findings

### 1. Frontmatter

| Field | Value | Status |
|---|---|---|
| `name` | `word-count` | OK — matches directory name. |
| `description` | Two sentences, includes use-when and skip-when guidance. | OK — concrete and discriminating. The skip-when clause helps the harness avoid mis-triggering on sentiment/readability requests. |
| `argument-hint` | `<path-to-file>` | OK — positional, single argument, matches the workflow. |

No `allowed-tools` field is present. This is not strictly required, but for a skill that only needs Bash + Write, declaring `allowed-tools` would be a useful least-privilege improvement.

### 2. Scope and triggering

- The description is well-bounded: it names the deliverable (three-number report), the trigger condition (word/line/byte counts on a specific file), and the explicit skip cases (sentiment, readability, frequency).
- "a specific file or set of files" in the description is **inconsistent** with the actual workflow, which only supports a single file (the `path` input is described as a single positional arg, and the output is a single `<path>.count.md`). Either drop "or set of files" from the description, or document a multi-file iteration in the workflow.

### 3. Inputs / validation

- `path` is required and must be a regular file. The validation rule `[ -f "$path" ]` is correct for "regular file."
- The text-file check uses `file --mime-type` and rejects anything not starting with `text/`. This is reasonable, but **MIME detection via `file(1)` has known false negatives**: files such as JSON, JavaScript, YAML, and many source files are sometimes returned as `application/json`, `application/javascript`, etc. The skill would refuse to count them despite being plain text. Consider one of:
  - Broadening the allow-list to `text/*` plus `application/json`, `application/xml`, `application/javascript`, etc.
  - Falling back to checking that the file is non-binary (no NUL bytes in the first N KB) instead of strict MIME-prefix matching.
- No size cap. A multi-GB file would be processed without warning. Documenting (or enforcing) a size ceiling would be safer.
- No handling for symlinks. `[ -f ]` follows symlinks — usually fine, but should be acknowledged.

### 4. Output

- Output path is `<path>.count.md` — sibling to the input with `.count.md` appended to the **full** filename, e.g. `notes.txt` → `notes.txt.count.md`. This is unambiguous but slightly unusual; users might expect `notes.count.md`. Document the chosen convention explicitly with an example to remove ambiguity.
- The output is described as a Markdown file but contains only three plain `key: N` lines with no Markdown structure (no heading, no code fence). The `.md` extension is therefore decorative. Either:
  - Use a `.txt` extension, or
  - Add a real Markdown structure (heading + fenced block or table) so the extension is meaningful.
- Idempotency is documented (overwrite on re-run). Good.
- Output path collision: if the input is itself named `something.count.md`, re-running would clobber the input. Worth a one-line guard.

### 5. Workflow correctness (highest-impact finding)

Step 2 says: *"Run `wc -wlc "$path"` and capture the three numbers from the first whitespace-delimited fields."*

- `wc -wlc` is correct (words, lines, bytes), but `wc`'s **output column order is fixed as `lines words bytes` regardless of flag order**. The instruction "the three numbers from the first whitespace-delimited fields" is therefore correct only if the model also knows that the first field is *lines*, not *words*. A naive read of the flag order (`-wlc` → words, lines, chars) maps the fields wrong and silently swaps the `words` and `lines` numbers in the report.
- Recommend rewriting step 2 to either:
  - Run `wc -l`, `wc -w`, `wc -c` separately and assign each result to the labelled output line, or
  - Explicitly note that `wc` always prints `lines words bytes` in that order regardless of flags, and parse accordingly.
- `-c` counts bytes, not characters. The skill correctly labels the field `bytes`, and the "When to skip" section correctly defers character counts elsewhere. Good.

### 6. Error handling

- "exit with a one-line error to the user and do not write anything" — clear, but doesn't say what the error wording or exit shape should be. Since there's no script, this is an instruction to the model; a tiny example string would reduce variance.
- No coverage for: unreadable file (exists but no read permission), `wc` returning non-zero, write failure on the output path.

### 7. "When to skip" section

Strong. It names concrete adjacent capabilities (character vs byte, grapheme counts, frequency, language detection) and tells the model to defer rather than approximate. This is exactly the boundary-setting that helps a skill stay narrow.

### 8. Bundled assets

There are no scripts, references, or assets — only `SKILL.md`. For a skill this small that's defensible, but the correctness risk in step 2 (finding 5) would be eliminated by a 5-line bundled `scripts/count.sh` that does the parsing deterministically and writes the report. Recommend adding one.

## Recommendations (priority order)

1. **High — fix the `wc` parsing instruction** so the output order (`lines words bytes`) cannot be confused with the flag order (`-wlc`). Either rewrite step 2 or add a `scripts/count.sh`.
2. **Medium — reconcile "single file" vs "set of files"** in the description, or add a multi-file workflow.
3. **Medium — broaden or replace the MIME check** so common plain-text formats (JSON, YAML, JS, etc.) aren't rejected.
4. **Low — clarify the output filename convention** with an example, and add a guard against clobbering an input named `*.count.md`.
5. **Low — make the `.md` extension meaningful** (real Markdown structure) or switch to `.txt`.
6. **Low — declare `allowed-tools`** (Bash, Write) for least privilege.
7. **Low — document failure modes** (unreadable, write-fail, oversized) and suggest a size cap.

## Verdict

The skill is small, well-scoped, and the description does a good job at triggering and skip behavior. The primary substantive risk is the `wc` field-ordering ambiguity in step 2, which can produce a silently wrong report. Addressing finding #1 would move this from "acceptable" to "solid."
