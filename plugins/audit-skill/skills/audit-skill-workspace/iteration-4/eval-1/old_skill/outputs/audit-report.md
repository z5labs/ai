# Audit: word-count

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-1/old_skill/work/skills/word-count/`
- Date: 2026-04-30
- Findings: 1  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 1)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description says "a specific file or set of files" (plural) but `argument-hint` (line 4), the body sentence on line 9 ("a single text file"), the `path` input (line 15), and the output (line 19) all handle one file. Either drop "or set of files" from the description or extend the workflow to accept multiple paths.

## Passing checks

- Idempotency declaration is present and specific: line 11 states re-running overwrites the prior report, and line 19 restates the overwrite behavior on the output path.
- Description names all three trigger elements — what (count words/lines/bytes and write a report), when (user asks for word/line/byte counts), and when-to-skip (sentiment/readability/frequency analysis).
- A dedicated `## When to skip` section (line 27) reinforces the negative case beyond the description.
- Inputs section grounds `path` with name, source (positional CLI arg), required-ness, and two concrete validation rules (`[ -f "$path" ]` and mime-type prefix `text/`).
- Output section gives the exact path pattern (`<path>.count.md`), the three-line format, and the overwrite behavior.
- Step 2 calls out the `wc -lwc` field-order gotcha explicitly ("Do NOT infer the mapping from the flag order"), removing a reproducibility trap a careful author could easily miss.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-1/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
