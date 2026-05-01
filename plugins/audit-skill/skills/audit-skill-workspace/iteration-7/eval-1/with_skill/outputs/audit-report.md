# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-1/with_skill/work/skills/word-count/`
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

- `SKILL.md:3` — description claims "a specific file or set of files" but `argument-hint` (`SKILL.md:4`), Inputs (`SKILL.md:15`), and Workflow (`SKILL.md:23`–`25`) only deliver single-file handling. Either narrow the description to "a text file" or extend the workflow to loop over multiple paths.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11` — "re-running on the same input file overwrites the report with current numbers"); re-run behavior on a different input is also called out.
- Description names what (count words/lines/bytes + write report), when to use (asks for word/line/byte counts), and when to skip (sentiment / readability / frequency need a different skill) (`SKILL.md:3`).
- Inputs section grounds the single input with name, source (positional CLI arg), required-ness, and validation (existence + mime-type) (`SKILL.md:15`).
- Output section documents path (`<path>.count.md`), format (three named lines), and overwrite behavior (`SKILL.md:19`).
- Workflow step 2 anchors `wc` field mapping to the tool's documented fixed-order behavior rather than flag order, eliminating a common reproducibility trap (`SKILL.md:24`).
- Security: no secret-shaped inputs, no prompted credentials, no URL-form connection strings, no generated files with credential-shaped values — the skill stays entirely out of the credential path.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-1/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
