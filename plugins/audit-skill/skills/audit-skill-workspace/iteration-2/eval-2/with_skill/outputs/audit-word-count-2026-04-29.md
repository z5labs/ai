# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-2/with_skill/work/.claude/skills/word-count/`
- Date: 2026-04-29
- Findings: 1  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 1)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description's "when to use" claims "a specific file or set of files" (plural) but `argument-hint` on line 4 is `"<path-to-file>"` and the workflow accepts a single `path`. Either narrow the description to a single file or extend the workflow to iterate over multiple paths.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11` states "This skill is **idempotent**" and `SKILL.md:19` documents overwrite intent).
- Description names what (count words/lines/bytes and write a report), when to use (user asks for word/line/byte counts), and when to skip (sentiment, readability, frequency analysis) — all three triggering elements (`SKILL.md:3`).
- Inputs section declares name, source, required-ness, and validation rule for `path` (`SKILL.md:13-15`).
- Output section declares path pattern (`<path>.count.md`), format (three lines, named keys), and pre-existing-file handling (overwrite) (`SKILL.md:17-19`).
- Workflow steps 1-3 are linear with explicit preconditions (validate before compute, compute before write) (`SKILL.md:23-25`).
- `When to skip` section lists concrete near-miss cases by name (`SKILL.md:27-29`).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-2/with_skill/work/audit-word-count-2026-04-29.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
