---
name: extract-binary-spec
description: Extract binary file format specifications (RFCs, vendor HTML docs, PDFs) into a chunked, per-structure markdown reference for Go encoder/decoder implementation. Use whenever the user wants to capture or document a binary format — packets, file containers, wire protocols, mainframe records — even if they don't say "extract" explicitly. Trigger on phrases like "document this format", "I want to implement a parser for X", "turn this RFC into a reference", or whenever the user mentions a specific binary format (PNG, gzip, DNS, BMP, RIFF, COBOL copybooks, SMF records) and wants its layout written down.
---

# extract-binary-spec

Read a binary format specification from any source (RFC, PDF, vendor HTML, local file) and produce a chunked, progressive-disclosure markdown reference suitable for the `implement-go-binary-file-library` agent. The output is per-structure so consumers load only what they need — important for mainframe-scale formats with hundreds of record types.

## Output layout

```
<format-name>/
├── SPEC.md                    # overview, conventions, top-level structure, indexes
├── structures/<name>.md       # one file per structure: byte diagram, field table, bit fields
├── encoding-tables/<name>.md  # one file per lookup table
└── examples/<name>.md         # worked examples (minimal, typical, complex)
```

Read `references/output-format.md` for the exact templates and required fields in each file before writing any output.

## Workflow

### Phase 0 — capture intent

Ask the user for:
- Spec source: URL, local path, or description
- Output directory (default: `./<format-name>/`)
- Versions, message types, or features to prioritize or skip

### Phase 1 — normalize the source to a local scratch file

Produce a single line-addressable local file (`_spec.txt` or `_spec.html`) so the rest of the pipeline is source-agnostic. **Prefer a dedicated skill if available; only fall back to raw tools when no skill exists.**

| Source type | Prefer (skill) | Fall back to (tool) |
|---|---|---|
| PDF (URL or local) | `pdf` skill — extract text, save to `_spec.txt` | `pdftotext spec.pdf _spec.txt`; otherwise `Read` with `pages:` |
| `.docx` | `docx` skill — extract text, save to `_spec.txt` | — |
| RFC at `rfc-editor.org/rfc/rfcNNNN` | — | `curl -L https://www.rfc-editor.org/rfc/rfcNNNN.txt > _spec.txt` |
| Generic HTML doc page | — | `curl -L <url> > _spec.html` (optionally `pandoc -f html -t markdown` if available) |
| Local `.txt` / `.md` / `.html` | — | use as-is |

After Phase 1 every downstream step reads from `_spec.{txt,html}` by line range. No re-fetching.

### Phase 2 — build a TOC

Write `_toc.md` with one row per section: `(section_id, title, line_start, line_end)`. Source-aware extraction, but the resulting TOC is uniform.

- **Plain text (RFC):** grep `^\d+(\.\d+)*\.\s+` for headings; line numbers come from `grep -n`
- **HTML:** grep `<h[1-6][^>]*>` (or `pandoc → markdown`, then grep `^#+`)
- **PDF text dump:** use the spec's TOC if present; else page boundaries

### Phase 3 — pick a tier

Choose by *scratch file size and structure count*, not source type:

| Tier | When | How to extract |
|---|---|---|
| 1 | ≤~15K words and ≤~10 structures | Read whole scratch file, write all output files directly. No subagents. |
| 2 | ≤~50 structures | One subagent per structure (or small group), addressed by `(file, line_range)`. Use the template in `references/extraction-subagent.md`. |
| 3 | 100s of structures (mainframe-scale) | Two-pass: first pass dispatches index subagents that skim each TOC entry and write a one-line summary to `_structures_index.md`. Second pass dispatches extraction subagents in batches of 10–20 structures, each writing its assigned `structures/<name>.md` files. Run batches in waves so the orchestrator's context stays clean. |

### Phase 4 — extract

Tier 1: do it inline.
Tiers 2 and 3: dispatch subagents using the prompt template in `references/extraction-subagent.md`. Each subagent writes directly to its assigned file — no scratch consolidation step.

For Tier 3, wave-dispatch (parallel within a wave, sequential between waves) so the orchestrator's context doesn't fill with subagent results. Between waves, list `structures/` to verify the wave's files exist; re-dispatch only the failed structures.

### Phase 5 — verify and clean up

1. Cross-check `SPEC.md`'s structure index against `structures/*.md` — every indexed structure has a file, every file is indexed.
2. Resolve cross-references: structures referencing each other and field-table → encoding-table lookups should use stable relative paths (`../encoding-tables/opcodes.md`) as anchors, not section IDs.
3. Delete `_spec.{txt,html}`, `_toc.md`, and (Tier 3) `_structures_index.md`.
4. Report to the user: output path, structure count, encoding-table count, example count, byte-order summary, and any `> **Ambiguity:**` callouts encountered during extraction.

## Important conventions

- **Capture every byte.** A decoder must consume every field in the wire format. If the spec mentions a field, it gets a row in some `structures/<name>.md` field table.
- **Always state byte order explicitly.** Globally in `SPEC.md#Conventions`, and per-structure when it differs.
- **Use Go type names, not paraphrases.** Field-table Type columns must hold real Go types: `uint8`/`uint16`/`uint32`/`uint64`, `[N]byte`, `[]byte`, or the PascalCase name of another structure file (e.g. `DomainName`). Never paraphrase ("a 32-bit IPv4 address", "a domain name reference"). For structures that aren't byte-offset-driven (bit-only registers, recursive encodings like DNS labels, single-field primitive payloads like RDATA-A), see the "Structure variants" section in `references/output-format.md` — pick the right presentation instead of forcing a byte-offset table.
- **Distinguish bytes from bits.** Column headers `Offset (bytes)` and `Bit(s)`. Never mix without labeling.
- **Don't invent.** If the spec is silent, mark it unspecified. If contradictory, add a `> **Ambiguity:**` callout so the implementer can decide.
- **Don't quote large excerpts verbatim.** Summarize in your own words; generate original hex examples that exercise the same rules.

## Next step for the user

Once extraction is complete, the typical next step is to scaffold the package with `/new-go-binary-file-library <name>` and then invoke the `implement-go-binary-file-library` agent. That agent partitions `SPEC.md` and the `structures/` / `encoding-tables/` / `examples/` directories across types/decoder/encoder subagents.
