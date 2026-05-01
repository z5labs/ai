# Context management checks

A skill ships with a fixed budget: every line of `SKILL.md` is loaded into the model's context whenever the skill triggers. Anything that bloats that load — oversized SKILL.md, inline reference content, default-verbose output, heavy work done in the main thread — taxes every invocation. The fix is almost always progressive disclosure: keep `SKILL.md` thin, push detail into sibling files loaded on demand, delegate heavy reads to subagents.

Multi-phase orchestrator skills also have a per-subagent *output* budget — a phase that produces unbounded edits, hands off open-ended summaries, or forces the next phase to re-read mutated source defeats the same discipline from the output side. The fix is symmetric: bound each phase's scope relative to input, give inter-phase artifacts a strict format and line cap, and rely on those artifacts instead of re-reading what earlier phases just grew.

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

### 6. Per-subagent scope unbounded relative to input

Applies only when SKILL.md describes a workflow that delegates work to subagents (look for at least one of: "subagent", "spawn", "Agent tool", "delegate to a subagent"). Phased workflows that run entirely in the main thread are out of scope — they have no per-subagent output to bound. For each phase, check whether per-call scope is bounded relative to input size:

- Does the phase state a partitioning rule? (e.g. "one sub-call per spec section", "one subagent per N items", "if section count > N, split into sub-units").
- Is there an up-front scope gate that counts input units and decides whether to partition *before* launching anything?

A phase that spawns one subagent covering all targeted input regardless of size is a finding — for a large input, the subagent's per-invocation output (edits, generated lines) grows unboundedly even when the orchestrator's input context is well-managed.

Raise: `<file>:<line> — phase <X> spawns one subagent regardless of input size; document a partitioning rule (e.g. one sub-call per spec section, split when count > N) so per-call output stays bounded`.

### 7. Inter-phase artifact files lack strict format and size cap

Look for scratch files written by one phase and consumed by another — names like `_context_*.md`, `_state_*.json`, running summaries, handoff notes. For each such file referenced in SKILL.md:

- Is its format strictly specified? Concrete row/line shape (e.g. "one symbol per line, signature only, no prose"; "fixed JSON schema with fields X, Y, Z"; "table with columns A, B, C") passes. Open-ended prose instructions like "capture the X", "list every Y", "describe the Z" do not.
- Is a hard line cap stated? (e.g. "≤400 lines"; "exceeding the cap is the signal the work-unit was sized too large and must be chunked").

The format must be deterministic enough that a later-phase subagent can rely on it as a substitute for re-reading the upstream source file.

Raise (format missing): `<file>:<line> — inter-phase artifact <name> is described in open-ended prose; specify a strict format (e.g. one signature per line) so later phases can rely on it without re-reading source`.

Raise (line cap missing): `<file>:<line> — inter-phase artifact <name> has no line cap; specify a hard cap (e.g. ≤400 lines) so growth above the cap signals the work-unit was sized too large and must be chunked`.

When both are missing, raise both findings.

### 8. Later phases re-read source files earlier phases mutated

In multi-phase workflows where each phase writes to a shared source file (`parser.go`, `decoder.go`, etc.) via `Edit`/append, check whether later phases instruct their subagents to `Read` that same file in full. The inter-phase summary from check 7 is meant to be the cross-reference of record for downstream phases — if SKILL.md still tells the next phase's subagent to read the mutated file whole, the summary is doing no work and the summary's growth compounds with the source file's growth.

Concretely, look for phase descriptions where:
- An earlier phase's outputs include `Edit`s to a file, AND
- A later phase's subagent inputs name that same file with an unsliced `Read` (no `(path, offset, limit)` form, no "use `_context_X.md` instead").

Raise: `<file>:<line> — phase <X> instructs subagent to Read <mutated-file> in full; rely on <summary-file> from the phase that edited <mutated-file> as the cross-reference of record, or pass a sliced (path, offset, limit) range`.

## What is NOT a finding

- A SKILL.md slightly over 500 lines if the content is genuinely workflow-heavy and there's no obvious thing to move out. Soft target.
- Long fenced blocks that are showing "do exactly this" templates the model must literally copy — those NEED to be inline.
- Sequential reads when each read informs the next (true dependency chain) — can't be parallelized.
- Single-phase skills, or multi-phase skills where each phase is O(1) by construction (one call, fixed-size output) — check 6 doesn't apply.
- Files that are themselves the deliverable of a phase (the phase's only job is to produce them, e.g. a final report) — check 7 doesn't apply; these are outputs, not inter-phase context.
- A later phase that re-runs tests or re-builds the mutated file rather than `Read`-ing it — check 8 doesn't apply; the file is being executed, not loaded into context.
- Phases operating on disjoint files (no shared mutated source across phases) — check 8 doesn't apply.
