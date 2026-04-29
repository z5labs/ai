---
name: word-count
description: Count words, lines, and bytes in a text file and write a small report. Use when the user asks for word/line/byte counts on a specific file or set of files. Skip when the user wants more sophisticated text analysis (sentiment, readability, frequency) — those need a different skill.
argument-hint: "<path-to-file>"
---

# word-count

Produce a three-number report (word count, line count, byte count) for a single text file and write the result to a sibling `.count.md` file.

This skill is **idempotent**: re-running on the same input file overwrites the report with current numbers. Re-running with a different input file produces a separate report.

## Inputs

- `path` (required, positional CLI arg): absolute or repo-relative path to a regular file. Validation: must exist (`[ -f "$path" ]`) and be a text file (test with `file --brief --mime-type "$path"` — `--brief` strips the leading `<path>:` prefix so the raw output is the mime type; reject if it doesn't start with `text/`).

## Output

- A file at `<path>.count.md` containing three lines: `words: N`, `lines: N`, `bytes: N`. Overwrites any existing file at that destination — re-running is the documented refresh path.

## Workflow

1. **Validate input.** If `path` doesn't exist or isn't a text file, exit with a one-line error to the user and do not write anything.
2. **Compute counts.** Run `wc -wlc "$path"` and capture the three numbers from the first whitespace-delimited fields.
3. **Write the report.** Overwrite `<path>.count.md` with the three lines above. Tell the user the report path.

## When to skip

If the user asks for things this skill doesn't produce — character count (vs byte count), Unicode-aware grapheme counts, frequency analysis, language detection — defer to a more capable text-analysis tool rather than misusing this one.
