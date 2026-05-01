# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-1/old_skill/work/skills/word-count/SKILL.md`
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

- `SKILL.md:3` — description claims "a specific file or set of files" (plural) but `argument-hint` at `SKILL.md:4` (`"<path-to-file>"`) and the `path` input at `SKILL.md:15` accept only a single file; the workflow at `SKILL.md:22-25` operates on one path. Either narrow the description to a single file or extend the workflow (and `argument-hint`) to accept multiple paths.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11` — "re-running on the same input file overwrites the report with current numbers"). The output-overwrite behavior is also restated in the Output section (`SKILL.md:19`).
- Description names what (count words/lines/bytes, write a report), when to use (user asks for word/line/byte counts), and when to skip (sentiment / readability / frequency / other text analysis) — `SKILL.md:3`.
- `Inputs` section documents `path` with source (positional CLI arg), required-ness, and validation rules (`[ -f "$path" ]` plus `file --brief --mime-type` ⇒ `text/`) at `SKILL.md:15`.
- `Output` section declares the destination path pattern (`<path>.count.md`), the exact three-line format, and the overwrite-on-existing behavior at `SKILL.md:19`.
- Reproducibility: `wc -lwc` field-order pitfall is called out explicitly (`SKILL.md:24` — "Do NOT infer the mapping from the flag order") so two runs converge on the same field assignments.
- Security: skill takes only a path argument, runs read-only `wc` / `file` commands, writes a sibling report; no secrets are prompted for, accepted as arguments, or written to disk.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-1/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
