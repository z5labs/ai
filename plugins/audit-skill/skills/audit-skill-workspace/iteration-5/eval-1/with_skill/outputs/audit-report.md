# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-5/eval-1/with_skill/work/skills/word-count/SKILL.md`
- Date: 2026-04-30
- Findings: 0  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 0, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

No findings.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`) — names the stance ("idempotent") and the mechanism ("overwrites the report with current numbers"); skip condition for the idempotency objective fires cleanly.
- Output overwrite behavior is restated at the point of mutation (`SKILL.md:19`, `SKILL.md:25`) — "Overwrites any existing file at that destination — re-running is the documented refresh path."
- Description names all three triggering elements: what (count words/lines/bytes, write report), when (user asks for word/line/byte counts on a file), and when-to-skip (sentiment, readability, frequency analysis) (`SKILL.md:3`).
- Inputs are fully grounded (`SKILL.md:15`): name (`path`), source (positional CLI arg), required-ness, and validation (existence test plus `file --brief --mime-type` mime-type check with the `text/` prefix rule).
- Outputs are fully grounded (`SKILL.md:17-19`): path pattern (`<path>.count.md`), format (three named lines), and pre-existing-file behavior (overwrite).
- `wc -lwc` field-order note (`SKILL.md:24`) — explicitly tells the model that `wc` emits `lines words bytes` regardless of flag order, and to map by position rather than by flag order. Defensive against a real reproducibility footgun.
- `--brief` flag rationale (`SKILL.md:15`) — explains *why* the flag is used (strips the leading `<path>:` prefix so the prefix-match against `text/` works), so a future edit can't drop it without understanding the consequence.
- "When to skip" section (`SKILL.md:27-29`) is concrete — names the specific near-miss capabilities (character count vs byte count, grapheme counts, frequency, language detection) rather than a generic "use a different tool" hedge.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration. (In this case there are no findings — the report is primarily a list of passing checks the author should preserve in any future revision.)
