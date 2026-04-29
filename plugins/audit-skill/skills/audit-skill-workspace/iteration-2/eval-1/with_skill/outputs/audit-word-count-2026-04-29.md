# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-1/with_skill/work/skills/word-count/`
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

- `SKILL.md:3` — description says "a specific file or set of files" (plural) but `argument-hint` (line 4), `path` input (line 15), and the workflow are single-file only. Either narrow the description to "a single text file" or extend the workflow to loop over inputs.
- `SKILL.md:24` — step 2 says "capture the three numbers from the first whitespace-delimited fields" of `wc -wlc`, but does not state the column-to-label mapping. `wc` always emits columns in the order `lines words bytes` regardless of flag order, while the output template (line 19) lists `words / lines / bytes`. State the mapping explicitly so the model does not transcribe columns in flag order.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`, reinforced at `SKILL.md:19`): re-running overwrites the report.
- Description names what (count words/lines/bytes, write report), when to use (word/line/byte counts on a file), and when to skip (sentiment, readability, frequency).
- Inputs section grounds `path` with source (positional CLI arg), required-ness, and validation rules (`[ -f "$path" ]`, `file --mime-type` starts with `text/`).
- Outputs section grounds the artifact with path (`<path>.count.md`), format (three labelled lines), and overwrite behavior.
- `argument-hint` is present (`SKILL.md:4`), so the calling syntax is visible without reading the body.
- A dedicated "When to skip" section (line 27) reinforces the negative case from the description.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-2/eval-1/with_skill/work/audit-word-count-2026-04-29.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
