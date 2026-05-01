# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-2/old_skill/work/.claude/skills/word-count/SKILL.md`
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

- `SKILL.md:3` — description claims "a specific file or set of files" (plural) but `argument-hint` at `SKILL.md:4` (`"<path-to-file>"`) and the workflow at `SKILL.md:9,15,23-25` accept a single path; either narrow the description to one file or extend the workflow to iterate over a set.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11` — "re-running on the same input file overwrites the report"; "different input file produces a separate report").
- Description names what (word/line/byte counts + report), when to use (user asks for word/line/byte counts on a file), and when to skip (sentiment/readability/frequency) (`SKILL.md:3`).
- Input `path` declared with source (positional CLI arg), required-ness, and concrete validation rules (`[ -f "$path" ]` and `file --brief --mime-type` must start with `text/`) (`SKILL.md:15`).
- Output path, format, and overwrite behavior are all stated (`SKILL.md:19`).
- `wc -lwc` field-mapping caveat is documented explicitly so the model doesn't infer the order from the flags (`SKILL.md:24`).
- A "When to skip" section reinforces the negative cases from the description (`SKILL.md:27-29`).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-2/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
