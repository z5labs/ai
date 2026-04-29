# Audit report template

This file defines two output formats — one for file mode, one for PR mode. Use the format that matches the routing decision in `SKILL.md`. The two are not interchangeable: file mode produces a standalone markdown report; PR mode produces a set of inline review comments plus a summary review whose body starts with a deduplication marker.

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

A single review posted via `gh api -X POST /pulls/{n}/reviews -f event=COMMENT` (see `references/pr-mode.md` for the full invocation). The first line is a deduplication marker: re-running the audit on the same head commit looks for this marker and short-circuits, so the format is load-bearing — keep the angle-bracket comment exactly as shown:

```
<!-- audit-skill: $HEAD_SHA -->
audit-skill: <total> findings across <N> objectives

idempotency: <n>, reproducibility: <n>, context-management: <n>, strict-definitions: <n>

<for each finding NOT posted inline (because its line wasn't in the diff):>
- `<file>:<line>` — **<objective>** — <description>

<if total == 0:>
audit clean — <total-checks-run> checks passed across all four objectives.
```

`$HEAD_SHA` is the PR's `headRefOid` at the moment the audit ran. The marker is the only way to detect a prior audit on the same commit — don't omit it, and don't reformat it.

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
