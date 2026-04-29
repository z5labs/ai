# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-3/eval-1/with_skill/work/skills/word-count/`
- Date: 2026-04-29
- Findings: 2  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 2)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description claims "a specific file or set of files" but `argument-hint` at line 4 and the Inputs section at line 15 only accept a single `path`; either narrow the description to one file or extend the workflow to iterate over a set.
- `SKILL.md:24` — step 2 says to run `wc -wlc "$path"` and "capture the three numbers from the first whitespace-delimited fields", but `wc` always emits fields in the order `lines words bytes` regardless of flag order, while the documented output at line 19 labels them `words: N`, `lines: N`, `bytes: N`. A model that maps fields by reading the flag order `-wlc` will mislabel lines as words and words as lines. State the field-to-label mapping explicitly (e.g. "field 1 = lines, field 2 = words, field 3 = bytes") or use `wc -w`, `wc -l`, `wc -c` separately.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`) — names the refresh path (overwrite) and notes that a different input file produces a separate report.
- Description names what (count words/lines/bytes and write a report), when to use (asks for word/line/byte counts), and when to skip (sentiment, readability, frequency).
- Inputs section (`SKILL.md:15`) declares name, source (positional CLI arg), required-ness, and validation (`[ -f "$path" ]` plus `file --mime-type` text-prefix test).
- Output contract (`SKILL.md:19`) names path pattern (`<path>.count.md`), format (three labeled lines), and overwrite behavior.
- A "When to skip" section (`SKILL.md:27`) reinforces the negative-case triggering signal beyond the frontmatter description.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-3/eval-1/with_skill/work/audit-word-count-2026-04-29.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
