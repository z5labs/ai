# Audit: word-count

- Target: `./.claude/skills/word-count/`
- Date: 2026-04-30
- Findings: 1  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 1, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description claims `"a specific file or set of files"` but `argument-hint` at `SKILL.md:4` and the `Inputs` section at `SKILL.md:15` only accept a single `path`; either narrow the description to one file or extend the workflow to iterate over multiple paths.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`: "This skill is **idempotent**: re-running on the same input file overwrites the report with current numbers").
- Output overwrite behavior is stated explicitly (`SKILL.md:19`: "Overwrites any existing file at that destination — re-running is the documented refresh path").
- Description names what / when / when-to-skip — verb + artifact ("Count words, lines, and bytes ... write a small report"), trigger phrasings ("when the user asks for word/line/byte counts"), and a negative case ("Skip when the user wants more sophisticated text analysis (sentiment, readability, frequency)"); the when-to-skip element is the most-missed and is present here (`SKILL.md:3`).
- Input `path` is fully grounded: source (positional CLI arg), required, and validation rule (`[ -f "$path" ]` plus `file --brief --mime-type` text-prefix check) all stated (`SKILL.md:15`).
- Output is fully grounded: path pattern (`<path>.count.md`), format (three lines `words: N`, `lines: N`, `bytes: N`), and pre-existing-file behavior (overwrite) all stated (`SKILL.md:19`).
- `argument-hint` is present for the slash-style invocation (`SKILL.md:4`).
- Step 2 explicitly disambiguates `wc -lwc` field order ("Map field 1 → `lines`, field 2 → `words`, field 3 → `bytes`. Do NOT infer the mapping from the flag order"), preventing a reproducibility footgun (`SKILL.md:24`).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration.
