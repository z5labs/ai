# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-1/with_skill/work/skills/word-count/`
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

- `SKILL.md:3` — description claims "a specific file or set of files" but `argument-hint` at `SKILL.md:4` is singular (`<path-to-file>`) and the Inputs section at `SKILL.md:15` declares a single `path`; either narrow the description to "a single file" or extend the workflow to iterate over multiple paths.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`): re-running on the same input overwrites the report with current numbers.
- Description names what (count words/lines/bytes and write a report), when (user asks for word/line/byte counts), and when-to-skip (sentiment, readability, frequency analysis) — all three triggering elements present (`SKILL.md:3`).
- Input is fully grounded: name (`path`), source (positional CLI arg), required-ness, and validation rules (must exist, must be `text/*` mime type) (`SKILL.md:15`).
- Output is fully declared: path pattern (`<path>.count.md`), format (three named lines), and overwrite behavior (`SKILL.md:19`).
- Workflow step 2 calls out the `wc -lwc` field-order trap explicitly and forbids inferring the mapping from flag order (`SKILL.md:24`) — a deliberate reproducibility anchor.
- Security: skill takes only a file path; no credential-shaped inputs, prompts, or generated files.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-1/with_skill/work/audit-word-count-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
