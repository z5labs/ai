# Audit report template

This is the exact output format. Use it verbatim in file mode and as the body of the top-level review comment in PR mode.

## File-mode template

```markdown
# Audit: <skill-name>

- Target: `<resolved-path>`
- Date: <YYYY-MM-DD>
- Findings: <total-count>  (idempotency: <n>, reproducibility: <n>, context-management: <n>, strict-definitions: <n>)

## Findings

### Idempotency

- `<file>:<line>` — <one-line description>. <optional one-sentence suggestion>
- `<file>:<line>` — ...

(If no findings under this objective, write a single line: `No findings.`)

### Reproducibility

- ...

### Context management

- ...

### Strict definitions

- ...

## Passing checks

A short bulleted list of objectives where the skill scored clean — useful so the author knows what NOT to lose in a revision.

- Idempotency declaration is present and specific (`SKILL.md:<line>`).
- Description names what / when / when-to-skip.
- ...

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration.
```

## PR-mode templates

### Inline review comments

One per finding whose `file:line` falls on a line modified in the PR. Body is a single paragraph:

```
**<Objective>** — <one-line description>.

<optional one-sentence suggestion>
```

The objective name (in bold) makes the category scannable in the PR thread. Keep the body short — the comment is anchored to the line, so the reader already has context.

### Top-level review summary

A single review comment posted via `gh pr review --comment`:

```
audit-skill: <total> findings across <N> objectives

idempotency: <n>, reproducibility: <n>, context-management: <n>, strict-definitions: <n>

<for each finding NOT posted inline (because its line wasn't in the diff):>
- `<file>:<line>` — **<objective>** — <description>

<if total == 0:>
audit clean — <total-checks-run> checks passed across all four objectives.
```

## Phrasing rules

- One finding per line. No multi-paragraph findings.
- Always lead with `<file>:<line>` so the author can jump straight to the spot.
- Name the exact phrase or pattern you matched when relevant ("`as appropriate`", "`rm -rf` on line 42").
- Suggest a fix when the fix is short and obvious. Don't suggest when the fix requires judgment the author should make themselves.
- Don't editorialize. "This is bad practice" / "this could be improved" / "consider whether this is really necessary" — all dead weight.
- Don't number findings globally; the per-objective grouping is the structure.

## What goes in "Passing checks"

Only call out checks that:
1. Took deliberate effort to satisfy (not just "the file exists"), AND
2. Are easy to inadvertently lose in a rewrite.

Examples worth listing: a clear idempotency declaration, a description with all three triggering elements, a reference TOC on a long file. Examples NOT worth listing: "SKILL.md is under 500 lines" by 200 lines (no risk of regression).
