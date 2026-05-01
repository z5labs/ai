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

- `SKILL.md:3` — description claims "a specific file or set of files" (plural) but `argument-hint: "<path-to-file>"` (line 4) and the `path` input (line 15) accept only a single file; either narrow the description to a single file or extend the workflow to loop over multiple paths.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`) — "re-running on the same input file overwrites the report with current numbers" plus the explicit refresh-path note at `SKILL.md:19`.
- Description names what (count words/lines/bytes and write a report), when to use (user asks for word/line/byte counts), and when to skip (sentiment, readability, frequency analysis) — `SKILL.md:3`.
- Inputs are fully grounded: name, source (positional CLI arg), required, and validation (`[ -f "$path" ]` + mime-type check) all stated — `SKILL.md:15`.
- Output contract is explicit: path pattern (`<path>.count.md`), format (three named fields), and overwrite behavior all named — `SKILL.md:19`.
- Workflow embeds a deliberate `wc` field-order note (`SKILL.md:24`) telling the model not to infer mapping from flag order — preempts a class of reproducibility bug.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-2/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
