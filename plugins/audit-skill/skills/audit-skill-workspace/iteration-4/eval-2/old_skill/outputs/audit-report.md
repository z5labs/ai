# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-2/old_skill/work/.claude/skills/word-count/SKILL.md`
- Date: 2026-04-30
- Findings: 1  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 1)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description claims "a specific file or set of files" (plural) but `argument-hint` (line 4: `<path-to-file>`), the `path` input (line 15), and the workflow (lines 21–25) only handle a single file; either narrow the description to a single file or extend the workflow to iterate over multiple paths.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`) — explicitly states re-running overwrites the report and re-running on a different input produces a separate report.
- Description names what (count words/lines/bytes and write a report), when to use (word/line/byte counts on a file), and when to skip (sentiment, readability, frequency analysis) (`SKILL.md:3`).
- Single declared input with name, source, required-ness, and validation rules (existence + text mime-type via `file --brief --mime-type`) (`SKILL.md:15`).
- Output path, format, and overwrite behavior all declared (`SKILL.md:19`).
- `wc -lwc` field-order ambiguity called out explicitly with the correct mapping (`SKILL.md:24`) — closes a real reproducibility hazard.
- `argument-hint` present in frontmatter (`SKILL.md:4`).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-2/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
