# Audit: word-count

- Target: `./.claude/skills/word-count/` (resolved by name from `./.claude/skills/<name>/`)
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

- `SKILL.md:3` — description claims "a specific file or set of files" (plural) but `argument-hint` at `SKILL.md:4` is `<path-to-file>` (singular) and the `path` input at `SKILL.md:15` is a single positional arg; the workflow processes one file. Either narrow the description to "a single file" or extend the workflow (and `argument-hint`) to accept a glob / multiple paths.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`) — names the refresh path ("overwrites the report with current numbers") and the disjoint-input behavior ("a different input file produces a separate report"). Easy to lose in a rewrite that compresses the preamble.
- Description names what / when / when-to-skip (`SKILL.md:3`) — all three trigger elements are present, and the when-to-skip clause names concrete near-misses (sentiment, readability, frequency).
- Inputs section grounds the `path` input with name, source (positional CLI arg), required-ness, and a two-part validation rule (`SKILL.md:15`).
- Outputs section names the destination path, the three-line format, and overwrite behavior (`SKILL.md:19`).
- Workflow step 2 calls out the `wc -lwc` field-order trap explicitly (`SKILL.md:24`) — this is the kind of land-mine note that's easy to drop in a rewrite and expensive to re-discover.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration.
