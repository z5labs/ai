# Extraction subagent prompt template

Used in Tier 2 and Tier 3 of `SKILL.md`'s workflow. Each subagent extracts one structure (or a small batch) from a line range of the normalized scratch file and writes directly to `structures/<name>.md` (and `encoding-tables/<name>.md` if its section defines lookup tables).

## Why direct-to-output (no scratch consolidation)

The earlier `extract-binary-spec` agent had subagents write to numbered scratch files (`_spec_part_NN.md`) which the orchestrator then concatenated. That step exists only because the output was a single file. With per-structure output, subagents write directly to their final destinations — no consolidation context cost, and the orchestrator can verify completeness by listing `structures/` instead of reading the contents of every file.

## Single-structure template (Tier 2)

```
You are extracting one binary format structure from a normalized spec.

INPUT
- Scratch file: <absolute path to _spec.txt or _spec.html>
- Lines: <start>-<end>
- Section title (from TOC): "<section title>"

OUTPUT
- Target file: <absolute path>/structures/<name>.md
- Schema: follow the `structures/<name>.md` template in
  plugins/file-library/skills/extract-binary-spec/references/output-format.md.
  Read that file before extracting.

ESTABLISHED CONVENTIONS for this format (do not redefine)
- Byte order: <big-endian | little-endian | mixed (and per-field rules)>
- Bit numbering: <MSB-0 | LSB-0>
- Size units: <bytes | octets | words (size)>
- Terminology: use "<field>" not "member"/"attribute"; use "<record>" not
  "row"/"entry"; <other format-specific terms the orchestrator has settled on>

INSTRUCTIONS
1. Read only lines <start>-<end> of the scratch file. Do not read other parts.
2. Extract every field defined in this section. Every byte the wire format
   carries needs a row in the field table.
3. Use Go-friendly types in the field table (uint8, uint16, [N]byte, []byte).
4. State byte order at the top of the file only if it differs from the
   format-wide convention listed above.
5. If this section defines bit-packed fields, fill in the Bit fields
   subsection using the established bit-numbering convention.
6. If this section defines lookup tables (opcodes, record types, error codes),
   write a separate file at <output-dir>/encoding-tables/<table-name>.md using
   the encoding-table template, and link to it from the field table's
   Description column.
7. If the spec is unclear or contradictory anywhere in your range, add a
   `> **Ambiguity:**` callout. Do not invent values.
8. Reply with one line: "wrote structures/<name>.md (N fields, M bit fields,
   K encoding tables)".

DO NOT
- Read sections outside your assigned line range.
- Quote large excerpts from the source verbatim — summarize in your own words.
- Edit SPEC.md or any structure file other than your target.
```

## Batch template (Tier 3)

For Tier 3 fan-out, give one subagent a small list (10–20) of structures to extract in one run. The subagent processes them sequentially, writing one file per structure. Sequential processing matters: the subagent reads its assigned range, writes the file, drops the content from working memory, then moves to the next.

```
You are extracting a batch of binary format structures from a normalized spec.

INPUT
- Scratch file: <absolute path>
- Structures to extract:
  1. <name-1>: lines <start>-<end>, section "<title>"
  2. <name-2>: lines <start>-<end>, section "<title>"
  ...

OUTPUT
- Target directory: <output-dir>/structures/
- Per-structure filename: <name>.md (kebab-case)
- Schema: same as the single-structure template — read
  references/output-format.md.

ESTABLISHED CONVENTIONS for this format (do not redefine)
<same block as the single-structure template>

INSTRUCTIONS
1. For each structure in your list, in order:
   a. Read only the lines <start>-<end> of the scratch file.
   b. Extract per the structure template.
   c. Write structures/<name>.md.
2. Do not load all line ranges into context at once — read, write, drop, next.
3. If multiple structures reference the same lookup table, write the
   encoding-table file once on the first occurrence and link to it from each
   field table afterward.
4. Reply with one summary line per structure: "structures/<name>.md (N fields)".
```

## Index-pass template (Tier 3, first wave)

The Tier 3 first wave skims TOC entries to produce a one-line summary per structure. The orchestrator uses these summaries to populate `SPEC.md`'s structure index without ever loading the full structure files.

```
You are summarizing binary format structures for a top-level index.

INPUT
- Scratch file: <absolute path>
- Structures to summarize:
  1. <name-1>: lines <start>-<end>, section "<title>"
  2. <name-2>: lines <start>-<end>, section "<title>"
  ...

OUTPUT
- Append-only target: <output-dir>/_structures_index.md
- Format: one line per structure:
    `<name>: <one-line description, ≤ 100 chars>`

INSTRUCTIONS
1. For each structure, read its line range, write a single descriptive line,
   then drop the content from context.
2. Lines should describe the structure's role (what it carries, when it
   appears) — not its field-by-field layout. Layout belongs in the
   per-structure extraction wave.
3. Reply with the count of summary lines written.
```

## What the orchestrator does between waves (Tier 3)

Between batched waves, the orchestrator:
- Lists `structures/` to confirm the wave's files were written
- Notes any subagent failures and re-dispatches just the failed structures
- Updates `SPEC.md`'s structure index with one-liner summaries from `_structures_index.md`
- Does NOT load the contents of completed structure files — that is what the index is for

This keeps the orchestrator's context proportional to the *number* of structures, not their total size.
