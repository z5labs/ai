# Extraction subagent prompt template

Used in Tier 2 and Tier 3 of `SKILL.md`'s workflow. Each subagent reads a line-bounded slice of the normalized scratch file and writes its assigned `##` section of `SPEC.md` to a numbered scratch part. The orchestrator concatenates the parts via `cat` without ever loading their content.

## Why per-section parts (not direct-to-output)

The output is a single `SPEC.md` whose `##` headings are load-bearing for the downstream partitioner. Letting subagents append to one file in parallel races on writes; letting one subagent own the file blocks parallelism. Numbered scratch parts give us both: subagents work in parallel on disjoint files, and the orchestrator concatenates them in section order with a shell `cat` that never enters context.

## Single-section template (Tier 2)

```
You are extracting one top-level section of a text format SPEC.md from a
normalized spec.

INPUT
- Scratch file: <absolute path to _spec.txt or _spec.html>
- Lines: <start>-<end>
- Section title (from TOC): "<section title or grouping>"
- Target SPEC.md section: <one of: ## Overview, ## Lexical Elements (Tokens),
  ## Structure (Grammar), ## Semantics, ## Examples, ## Appendix>

OUTPUT
- Target file: <absolute path>/_spec_part_NN.md
  (NN is the 1-based ordinal of this section in the final SPEC.md, zero-padded
  to two digits — e.g. 01 = Overview, 02 = Lexical Elements, etc.)
- Schema: follow the matching `##` section template in
  plugins/file-library/skills/extract-text-spec/references/output-format.md.
  Read that file before extracting. Begin your output file with the literal
  `## <section name>` heading so concatenation produces a valid SPEC.md.

ESTABLISHED CONVENTIONS for this format (do not redefine)
- Grammar notation: <ABNF | EBNF | BNF | informal prose distilled to EBNF>
- Encoding: <UTF-8 | ASCII | other>
- Terminology: use "<field>" not "member"/"property"; use "<production>" not
  "rule"/"nonterminal"; <other format-specific terms the orchestrator has
  settled on, e.g. "table" vs "object" for TOML>
- Token type naming: PascalCase (e.g. StringLiteral, RecordSeparator)

INSTRUCTIONS
1. Read only lines <start>-<end> of the scratch file. Do not read other parts.
2. Produce only your assigned `##` section — do not include any other top-level
   `##` headings; the orchestrator concatenates parts in order.
3. For ## Lexical Elements: cover every token class the spec defines in your
   range. For ## Structure: write each production in the established grammar
   notation. For ## Examples: produce real, parseable documents — do not
   verbatim-copy long examples from the source.
4. Use the established conventions verbatim. Do not introduce new terms.
5. If the spec is unclear or contradictory anywhere in your range, add a
   `> **Ambiguity:**` callout. Do not invent values or rules.
6. Do not write a `# <Format Name> Specification Reference` heading — the
   orchestrator's Overview part owns the H1.
7. Reply with one line: "wrote _spec_part_<NN>.md (<bytes> bytes,
   <feature_count> <features>)" where features are token classes / productions
   / examples appropriate to your section.

DO NOT
- Read sections outside your assigned line range.
- Quote large excerpts from the source verbatim — summarize in your own words.
- Edit any other `_spec_part_*.md` file.
- Edit `SPEC.md` directly.
```

## Batch template (Tier 3)

For Tier 3 fan-out within a wave, give one subagent a small list (5–15 sections) to extract sequentially. The subagent processes them one at a time, dropping each section's content from working memory before moving to the next.

```
You are extracting a batch of text format SPEC.md sections from a normalized
spec.

INPUT
- Scratch file: <absolute path>
- Sections to extract:
  1. _spec_part_<NN>.md  (## <section name>): lines <start>-<end>
  2. _spec_part_<NN>.md  (## <section name>): lines <start>-<end>
  ...

OUTPUT
- Target directory: <output-dir>/
- Per-section filename: _spec_part_<NN>.md (zero-padded ordinal)
- Schema: same as the single-section template — read references/output-format.md.

ESTABLISHED CONVENTIONS for this format (do not redefine)
<same block as the single-section template>

INSTRUCTIONS
1. For each section in your list, in order:
   a. Read only the lines <start>-<end> of the scratch file.
   b. Extract per the section template.
   c. Write _spec_part_<NN>.md.
2. Do not load all line ranges into context at once — read, write, drop, next.
3. Reply with one summary line per section: "_spec_part_<NN>.md (<bytes>,
   <feature_count> <features>)".
```

## Index-pass template (Tier 3, first wave)

The Tier 3 first wave skims TOC entries to produce a one-line summary per section. The orchestrator uses these summaries to plan extraction waves and to populate cross-references in `## Examples` without ever loading section bodies.

```
You are summarizing text format spec sections for a top-level index.

INPUT
- Scratch file: <absolute path>
- Sections to summarize:
  1. <section_id>: lines <start>-<end>, title "<title>"
  2. <section_id>: lines <start>-<end>, title "<title>"
  ...

OUTPUT
- Append-only target: <output-dir>/_sections_index.md
- Format: one line per section:
    `<section_id>: <one-line description, ≤ 100 chars>`

INSTRUCTIONS
1. For each section, read its line range, write a single descriptive line,
   then drop the content from context.
2. Lines should describe the section's role (what it defines, how it relates
   to lexical structure or grammar) — not its full contents. Detailed
   extraction belongs in subsequent waves.
3. Reply with the count of summary lines written.
```

## What the orchestrator does between waves (Tier 3)

Between batched waves, the orchestrator:
- Lists `_spec_part_*.md` to confirm the wave's files were written.
- Notes any subagent failures and re-dispatches just the failed sections.
- Does NOT load the contents of completed parts — that is what the index is for.

This keeps the orchestrator's context proportional to the *number* of sections, not their total size. The final concatenation step (`cat _spec_part_*.md > SPEC.md`) is also a no-context-cost operation.
