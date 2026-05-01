# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-1/old_skill/work/skills/word-count/`
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

- `SKILL.md:3` — description's "or set of files" promises plural inputs, but `argument-hint: "<path-to-file>"` (line 4), the Inputs section ("`path` ... a regular file", line 15), and the Output section ("for a single text file", line 9) all handle exactly one file. Either drop "or set of files" from the description or extend the workflow to loop over a glob.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11` — "re-running on the same input file overwrites the report with current numbers").
- Description names what (count words/lines/bytes, write a report), when (user asks for word/line/byte counts), and when-to-skip (sentiment, readability, frequency).
- Input `path` is grounded: source (positional CLI arg), required, with concrete validation rules — `[ -f "$path" ]` and `file --brief --mime-type` must start with `text/` (`SKILL.md:15`).
- Output is grounded: path (`<path>.count.md`), format (three named lines), overwrite behavior stated (`SKILL.md:19`).
- `wc -lwc` field-order pitfall is called out explicitly with the correct mapping (`SKILL.md:24`).
- `argument-hint` present in frontmatter (`SKILL.md:4`).
- "When to skip" section enumerates concrete near-misses (character vs byte, grapheme counts, frequency, language detection) — `SKILL.md:28-29`.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-1/old_skill/work/audit-word-count-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
