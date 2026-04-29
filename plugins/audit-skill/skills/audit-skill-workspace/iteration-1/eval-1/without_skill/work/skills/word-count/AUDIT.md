# Audit Report — `word-count` skill

**Target:** `./skills/word-count/SKILL.md`
**Auditor:** Claude (baseline, no audit checklist)
**Date:** 2026-04-29
**Verdict:** Mostly solid for a small skill. **One real bug** in the wc parsing instructions, and **one description/inputs mismatch** that will confuse triggering. A handful of smaller robustness gaps worth fixing before the skill sees real use.

---

## Findings

Findings are tagged by severity:

- **bug** — likely to produce wrong output
- **inconsistency** — internal contradiction in the skill itself
- **robustness** — edge case that will cause user-visible failure
- **nit** — minor / stylistic

### 1. (bug) `wc -wlc` output order is wrong in the workflow description

> "Run `wc -wlc "$path"` and capture the three numbers from the first whitespace-delimited fields."

`wc`'s output column order is **fixed**: `newlines words bytes` (POSIX), regardless of the order the flags are passed. The skill flags `-wlc` (word, line, byte) read left-to-right as if the columns will appear in that order. They will not. A naive implementer reading this skill will produce a report where `words:` and `lines:` are swapped.

**Fix:** Either

- spell out the column order explicitly: "wc emits `lines words bytes` in that fixed order — assign accordingly," or
- run three separate calls (`wc -w`, `wc -l`, `wc -c`) so position is unambiguous, or
- show the exact extraction snippet (e.g. `read lines words bytes _ < <(wc -wlc "$path")`).

This is the most consequential issue in the skill.

### 2. (inconsistency) Description says "set of files," everything else says one file

The description's trigger clause:

> "Use when the user asks for word/line/byte counts on a specific file **or set of files**."

But:

- `argument-hint: "<path-to-file>"` — singular.
- `Inputs` says `path` is a single positional arg, with `[ -f "$path" ]` validation — singular.
- `Workflow` walks through one file end-to-end.
- `Output` writes one sibling `.count.md` — implies one input.

**Effect:** the skill will be triggered on multi-file requests it cannot fulfill, and it has no documented fallback for that case. Either remove "or set of files" from the description, or extend the inputs/workflow to loop over multiple paths and write one report per file. Removing the phrase is the smaller change and is consistent with the rest of the skill.

### 3. (robustness) Empty files are rejected by the text-file check

`file --mime-type` reports empty files as `inode/x-empty`, which does not start with `text/`. The validation step in the workflow will therefore exit with an error on a perfectly legitimate empty text file — even though `wc` would happily report `0 0 0`.

**Fix:** either special-case zero-byte files (counts are trivially zero) before the mime-type test, or accept `inode/x-empty` alongside `text/*`. Worth at least documenting if the rejection is intentional.

### 4. (robustness) Destination writability is never checked

Workflow assumes `<path>.count.md` is writable. If `path` lives in a read-only directory (e.g. a mounted snapshot, a system path the user is reading from), the validation passes but the write step fails partway through. The skill should either pre-check writability of the parent dir, or document that the failure mode is "best-effort write, surface the OS error to the user."

### 5. (robustness) No handling of `path` ending in `.count.md`

Running the skill on `notes.count.md` produces `notes.count.md.count.md` — possibly intentional, but a user re-running by mistake will start accumulating nested reports. Cheap defense: if `path` ends in `.count.md`, refuse with a one-line error ("path looks like a count report; pass the original file instead").

### 6. (nit) Output is named `.md` but is not Markdown

The three-line body (`words: N`, `lines: N`, `bytes: N`) is plain key-value text. Renaming to `.count.txt`, or formatting as a real Markdown snippet (a table or a fenced block), would match the extension. Low priority; current shape is fine if the `.md` extension is just to signal "human-readable text companion."

### 7. (nit) "Idempotent" claim is correct but slightly under-specified

The skill calls itself idempotent because it overwrites. Strictly, idempotence here is "same input file + unchanged contents → same output." If the input changes between runs the output changes too — that is the *point*, but the wording could be tightened to "re-runnable / overwrites" to avoid suggesting the output is invariant.

### 8. (nit) `file --mime-type` portability

`file --mime-type` is GNU-style. BSD `file` (macOS default before recent versions) accepts `-I` instead. If the skill is ever invoked on a stock macOS shell, the validation step silently fails. Either pin to `file -b --mime-type` (works on both modern GNU and recent BSD), or document the dependency.

---

## What the skill gets right

- **Description has both a trigger clause and a skip clause**, with concrete examples of out-of-scope tasks (sentiment, readability, frequency). This is the right shape for skill descriptions and is the single biggest factor in correct triggering.
- **Inputs are validated before any side effect**, and the failure mode is documented ("exit with a one-line error to the user and do not write anything").
- **Output location is deterministic** (sibling file with a fixed suffix), which is what makes the idempotency claim work.
- **Scope is tightly drawn.** The skill resists the temptation to grow into a general text-analysis tool, and the "When to skip" section names the categories it punts on.
- **Argument shape (`argument-hint`) is present and matches the single-input workflow** (modulo finding #2).

---

## Recommended priority

1. Fix finding #1 (wc column order) — this is the only finding that breaks correctness.
2. Fix finding #2 (description vs. inputs) — affects when the skill fires.
3. Address #3–#5 if the skill is going to be run unattended.
4. #6–#8 are polish.

Nothing in the skill warrants a rewrite. With #1 and #2 resolved, the skill is ready to ship as-is.
