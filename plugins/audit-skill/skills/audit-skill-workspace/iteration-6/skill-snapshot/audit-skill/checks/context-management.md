# Context management checks

A skill ships with a fixed budget: every line of `SKILL.md` is loaded into the model's context whenever the skill triggers. Anything that bloats that load — oversized SKILL.md, inline reference content, default-verbose output, heavy work done in the main thread — taxes every invocation. The fix is almost always progressive disclosure: keep `SKILL.md` thin, push detail into sibling files loaded on demand, delegate heavy reads to subagents.

## Checks

### 1. SKILL.md size

```
wc -l SKILL.md
```

- 0–500 lines: pass.
- 501–800 lines: raise a finding suggesting which sections look like reference material that could move out.
- 801+ lines: raise a finding asking the author to identify the largest sections and move them under `references/`.

Phrase as: `SKILL.md:<total> — SKILL.md is <N> lines (target: ≤500). Sections most likely to move: <top-2-by-size>`.

To find the biggest sections, scan H2 (`^## `) headings and measure the byte distance to the next H2. Surface the top two.

### 2. Inline content that should be a sibling file

Look for:
- Fenced code blocks longer than ~50 lines in `SKILL.md` — usually these are templates that should live in `assets/` or `references/`.
- Tables longer than ~30 rows.
- "Section X format" or "Output schema" content longer than half a page — almost always belongs in a `references/<topic>.md`.

Raise: `SKILL.md:<line> — <description> is <N> lines inline; consider moving to references/<topic>.md and referencing it`.

### 3. Default-verbose output

Read the workflow / output sections of `SKILL.md`. If the documented final report:
- Dumps every intermediate step to the user (transcript-style), OR
- Shows full file contents rather than a summary + path,

raise: `<file>:<line> — described output dumps <X> by default; consider summarizing and offering details on request`.

The exception: skills whose entire purpose is to produce verbose output (a generated file, a full report). Those are fine — the verbosity IS the deliverable.

### 4. Missing subagent delegation for heavy lifts

Look for steps that:
- Read multiple large files (>3 files of >200 lines each in sequence),
- Iterate over many items where each item involves non-trivial reading/grepping,
- Could parallelize across independent inputs.

For each such step, check whether `SKILL.md` mentions delegating it to a subagent (Agent tool, "spawn", "in parallel", "subagent"). If not, raise: `<file>:<line> — step <X> reads/processes <Y> independent items in main context; consider delegating to subagents`.

### 5. Reference files without a TOC

For each file in `references/` longer than ~300 lines, check the first 20 lines for a table of contents (a list of links or sections). Without one, the model has to load the whole file to find what it needs.

```
for f in references/*.md; do lines=$(wc -l < "$f"); [ "$lines" -gt 300 ] && head -20 "$f" | grep -qiE '^(## |##  |- \[|table of contents|toc)' || echo "$f:$lines no TOC"; done
```

Raise: `references/<file>:1 — file is <N> lines without a TOC at the top; readers must scan the whole file to find a section`.

## What is NOT a finding

- A SKILL.md slightly over 500 lines if the content is genuinely workflow-heavy and there's no obvious thing to move out. Soft target.
- Long fenced blocks that are showing "do exactly this" templates the model must literally copy — those NEED to be inline.
- Sequential reads when each read informs the next (true dependency chain) — can't be parallelized.
