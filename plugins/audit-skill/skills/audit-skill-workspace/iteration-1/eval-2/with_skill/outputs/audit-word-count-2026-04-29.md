# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-1/eval-2/with_skill/work/.claude/skills/word-count/`
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

- `SKILL.md:3` — description's "when to use" says "a specific file or set of files", but the Inputs section (`SKILL.md:15`) declares a single positional `path` and the workflow processes exactly one file. The "or set of files" phrasing will over-trigger the skill on multi-file requests it cannot satisfy. Either drop "or set of files" from the description, or extend the contract to accept multiple paths.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:11`: "This skill is **idempotent**: re-running on the same input file overwrites the report with current numbers").
- Output overwrite intent is documented at the write site (`SKILL.md:19`, `SKILL.md:25`).
- Description names what (count words/lines/bytes + write report), when (user asks for word/line/byte counts), and when-to-skip (sentiment/readability/frequency analysis) — all three triggering elements present.
- Input `path` is fully specified: source (positional CLI arg), required-ness, and validation rules (`[ -f "$path" ]` and `file --mime-type` text/* check) at `SKILL.md:15`.
- Output is fully specified: path pattern (`<path>.count.md`), format (three lines `words: N` / `lines: N` / `bytes: N`), and overwrite behavior at `SKILL.md:18-19`.
- `argument-hint` is present in frontmatter (`SKILL.md:4`).
- Explicit "When to skip" section enumerating near-miss tasks (`SKILL.md:27-29`).
- No vague directives, no implicit environment dependencies, no non-deterministic verbs.
- SKILL.md is 29 lines — well under the 500-line context budget; no inline content needs to move to references/.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration.
