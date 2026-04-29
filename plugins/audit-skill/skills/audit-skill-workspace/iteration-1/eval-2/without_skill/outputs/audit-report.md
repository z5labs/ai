# Audit Report: word-count skill

**Target:** `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-1/eval-2/without_skill/work/.claude/skills/word-count/SKILL.md`

**Date:** 2026-04-29

**Auditor:** baseline judgment (no checklist)

## Overview

The `word-count` skill is a small, well-scoped utility that counts words, lines, and bytes in a text file and writes a sibling `.count.md` report. The skill is documented clearly, declares idempotency, and provides positive and negative triggering signals in its `description`. Below is a list of issues found, ranked by severity.

## Findings

### Bug — `file --mime-type` parse will misbehave without `-b`

**Severity:** High

The validation step says:

> test with `file --mime-type`; reject if it doesn't start with `text/`

By default, `file --mime-type "$path"` prints `<path>: <mime>` — the output begins with the filename, not the mime type. A literal "starts with `text/`" check will reject every non-empty path. The fix is to use the brief flag:

```bash
file --mime-type -b "$path"
```

…which prints just the mime string. Alternatively, parse after the trailing `: `. As written, the skill's validation logic is broken for the common case.

### Bug risk — `wc -wlc` output order is fixed, not flag-driven

**Severity:** Medium

The workflow says:

> Run `wc -wlc "$path"` and capture the three numbers from the first whitespace-delimited fields.

`wc` always emits its columns in the fixed order **lines, words, bytes** regardless of the flag order on the command line. A reader who follows the flag order `-wlc` may incorrectly map field 1 → words, field 2 → lines, field 3 → bytes, yielding swapped `words` and `lines` values. The skill should either:

- Reorder the flags to `wc -lwc` so the flag order matches the output order, or
- Explicitly state the output order ("`wc` prints lines, then words, then bytes — assign accordingly"), or
- Use three separate invocations (`wc -w`, `wc -l`, `wc -c`) to remove ambiguity.

### Inconsistency — argument hint vs. description scope

**Severity:** Low

- `argument-hint: "<path-to-file>"` is singular.
- `description` says "a specific file **or set of files**".
- The workflow only handles one file (`$path`).

Either tighten the description to a single file, or extend the workflow to iterate over multiple paths. Today, a user invoking with several paths would silently get a one-file behavior.

### Missing — `allowed-tools` declaration

**Severity:** Low

No `allowed-tools` field in the frontmatter. The skill only needs `Bash` (and arguably `Write` for the report, though the workflow uses redirection). Restricting the tool surface reduces blast radius and clarifies intent. Recommended:

```yaml
allowed-tools: Bash, Write
```

### Missing — error format and write-failure handling

**Severity:** Low

Step 1 says "exit with a one-line error … and do not write anything." That is fine for input validation, but step 3 silently assumes the sibling write succeeds. If the input lives in a read-only directory (e.g. `/usr/share/...`), the skill will fail at write time without prescribed handling. Suggest documenting the failure mode (e.g., "if the sibling write fails, report the error path and abort").

### Minor — output filename convention

**Severity:** Informational

`<path>.count.md` produces names like `notes.txt.count.md` (double extension). It is unambiguous and idempotent, but readers may expect `notes.count.md` (extension replaced). The current behavior is documented and acceptable; flagging only because it is mildly unusual.

### Minor — no example output

**Severity:** Informational

The skill specifies the format (`words: N`, `lines: N`, `bytes: N`) but provides no concrete example. A 3-line example block would remove ambiguity (e.g., is there a trailing newline? lowercase keys? exact spacing?) and prevent drift across runs.

## Strengths

- Clear positive and negative triggering language in the `description`.
- Explicit idempotency contract (re-runs overwrite).
- Honest "When to skip" section that defers Unicode/grapheme/character work and frequency/sentiment analysis to other tools — avoids scope creep.
- Input validation called out as a first step before any side effect.
- Quoting of `"$path"` in the documented commands is correct.
- Compact: one screenful, easy to follow.

## Recommended fixes (priority order)

1. Add `-b` to the `file --mime-type` invocation (or change the parse), so validation actually works.
2. Make the `wc` output-order assumption explicit (preferred: change flags to `-lwc` to mirror output order, then list the fields as lines/words/bytes).
3. Reconcile `argument-hint` with the "set of files" wording in the description.
4. Add `allowed-tools: Bash, Write` to the frontmatter.
5. Document the sibling-write failure mode and add a short example of the report contents.

## Verdict

The skill's scope, structure, and tone are good. Two real bugs (`file --mime-type` parsing and `wc` field-order ambiguity) would cause the skill to fail or produce wrong numbers in normal use; both are easy fixes. After those, the remaining items are polish.
