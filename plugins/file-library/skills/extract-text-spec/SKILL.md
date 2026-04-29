---
name: extract-text-spec
description: Extract text-based file format specifications (RFCs, ABNF/EBNF grammars, vendor HTML docs, PDFs) into a single progressive-disclosure SPEC.md suitable for Go tokenizer/parser/printer implementation. Use whenever the user wants to capture or document a text format — config languages, query languages, serialization formats, markup — even if they don't say "extract" explicitly. Trigger on phrases like "document this format", "I want to write a parser for X", "turn this RFC into a reference", or whenever the user names a specific text format (JSON, TOML, INI, YAML, CSS, GraphQL, ABNF, HCL) and wants its grammar written down.
---

# extract-text-spec

Read a text format specification from any source (RFC, PDF, vendor HTML, local file) and produce a single `SPEC.md` whose `##` section boundaries map directly onto what the `implement-go-text-file-library` agent partitions per pipeline stage (tokenizer / parser / printer). Context never grows with spec size: large specs are extracted by subagents that read line-bounded slices of a normalized scratch file and write to per-section scratch outputs the orchestrator concatenates without loading.

## Output layout

```
<format-name>/
└── SPEC.md     # Overview, Lexical Elements, Grammar, Semantics, Examples, Appendix
```

A single file by design — `implement-go-text-file-library` partitions it by `##` heading into per-stage scratch files. Read `references/output-format.md` for the exact section template and the rules each section must follow before writing any output.

## Workflow

### Phase 0 — capture intent

Ask the user for:
- Spec source: URL, local path, or description
- Output path (default: `./<format-name>/SPEC.md`)
- Sections, productions, or features to prioritize or skip (e.g. "section 4 only", "JSON Pointer is out of scope")

### Phase 1 — normalize the source to a local scratch file

Produce a single line-addressable local file (`_spec.txt` or `_spec.html`) so the rest of the pipeline is source-agnostic. **If a dedicated extraction skill (PDF, docx) is installed in the runtime environment, prefer it; this skill does not bundle one. Otherwise use the fallback tool — every row below has one.**

| Source type | Prefer (if installed) | Fall back to (tool) |
|---|---|---|
| PDF (URL or local) | a PDF-extraction skill if available (e.g. an `example-skills:pdf`-style skill bundled with some Claude Code installs) — extract text, save to `_spec.txt` | `pdftotext spec.pdf _spec.txt`; otherwise `Read` with `pages:` |
| `.docx` | a docx-extraction skill if available — extract text, save to `_spec.txt` | `pandoc -f docx -t plain spec.docx > _spec.txt`; if pandoc is unavailable, ask the user to convert to `.txt` or `.pdf` |
| RFC at `rfc-editor.org/rfc/rfcNNNN` | — | `curl -L https://www.rfc-editor.org/rfc/rfcNNNN.txt > _spec.txt` |
| Generic HTML doc page | — | `curl -L <url> > _spec.html` (optionally `pandoc -f html -t markdown` if available) |
| Local `.txt` / `.md` / `.html` | — | use as-is |

After Phase 1 every downstream step reads from `_spec.{txt,html}` by line range. No re-fetching.

### Phase 2 — build a TOC

Write `_toc.md` with one row per section: `(section_id, title, line_start, line_end)`. Source-aware extraction, but the resulting TOC is uniform.

- **Plain text (RFC):** `grep -n` for `^\d+(\.\d+)*\.\s+` headings
- **HTML:** `grep -n '<h[1-6][^>]*>'` (or `pandoc → markdown` then `grep -n '^#+ '`)
- **PDF text dump:** use the spec's TOC if present; else page boundaries

The TOC is the orchestrator's only persistent map of the source. Every subsequent line-range reference must come from it.

### Phase 3 — pick a tier

Choose by *scratch file size and grammar surface*, not source type. Counts below refer to distinct token classes plus distinct grammar productions in the spec:

| Tier | When | How to extract |
|---|---|---|
| 1 | ≤ ~15K words and ≤ ~30 productions | Read the whole scratch file inline, write `SPEC.md` directly. No subagents. |
| 2 | medium spec, ≤ ~100 productions | One subagent per top-level `##` section of `SPEC.md` — all six: Overview, Lexical Elements, Structure, Semantics, Examples, Appendix. Each writes a numbered scratch file (`_spec_part_NN.md`, ordinals 01–06); the orchestrator separately writes `_spec_part_00.md` containing the H1 and concatenates without reading. Use `references/extraction-subagent.md`. |
| 3 | large or sprawling spec (HTML5, SQL, full LaTeX) | Two-pass. First pass: index subagents skim each TOC entry into a one-line summary in `_sections_index.md`. Second pass: extraction subagents in waves, each owning a slice of the TOC, writing their assigned `_spec_part_NN.md` files. Orchestrator concatenates between waves and never reads section bodies. |

### Phase 4 — extract

**Tier 1.** Read `_spec.{txt,html}` and `references/output-format.md`, write `SPEC.md` directly.

**Tier 2.** First, the orchestrator writes `_spec_part_00.md` containing exactly the H1 line followed by a blank line:

```
# <Format Name> Specification Reference

```

No subagent ever writes the H1 — that ordinal (00) is reserved for the orchestrator. Then dispatch one subagent per top-level section using `references/extraction-subagent.md`. Each subagent:
- Reads only its assigned line range from the scratch file
- Reads `references/output-format.md` to learn the section template
- Writes its output to `_spec_part_NN.md`, where NN is the section's ordinal: `01=Overview, 02=Lexical Elements, 03=Structure, 04=Semantics, 05=Examples, 06=Appendix`

After all subagents complete, the orchestrator concatenates without loading content (lexicographic sort puts `00` first):

```bash
cat _spec_part_*.md > SPEC.md
```

Subagents are responsible for following the established conventions you pass them (terminology, grammar notation). The orchestrator does **not** read the parts to reconcile — that would defeat the context budget. If the subagent prompt is right, the parts are already consistent.

**Tier 3.** Same H1 step first: the orchestrator writes `_spec_part_00.md` with the H1 before any extraction wave begins. Then wave-dispatch (parallel within a wave, sequential between waves). The first wave is index-only — each subagent skims its TOC slice and appends one-line summaries to `_sections_index.md`. The orchestrator uses that index to drive subsequent extraction waves and to populate the `## Examples` cross-reference list, never loading the actual section bodies. Between waves, list the `_spec_part_*.md` files to verify each wave's slice landed; re-dispatch only the missing parts.

### Phase 5 — verify and clean up

1. Confirm `SPEC.md` contains every required `##` section from `references/output-format.md`. Missing sections almost always mean a subagent failed silently — re-dispatch its slice rather than papering over the gap.
2. Confirm at least three worked examples appear under `## Examples` (Minimal / Typical / Complex). The implementer agent uses these as test fixtures, so absence of any one breaks the downstream pipeline.
3. Delete `_spec.{txt,html}`, `_toc.md`, all `_spec_part_*.md`, and (Tier 3) `_sections_index.md`.
4. Report to the user: output path, tier used, token-class count, production count, example count, and any `> **Ambiguity:**` callouts encountered during extraction.

## Important conventions

- **Capture every token class.** A tokenizer must be able to classify every character sequence the format defines. If the spec mentions a syntax element, it gets a row under `## Lexical Elements`.
- **Write grammar in EBNF (or ABNF if the spec uses it).** Each grammar production becomes a parser action — paraphrased grammar rots into ambiguity at implementation time, so use the formal notation the spec uses and keep production names verbatim where reasonable.
- **State whitespace and comment significance explicitly.** "Whitespace ignored" is not enough — say where it's significant (string literals, line-oriented formats), where it's not, and whether comments nest.
- **Use stable terminology across sections.** Pick "field" or "member" — not both. When you dispatch subagents, list the chosen terms so each section uses them. Inconsistent terminology forces the implementer agent to guess at type names.
- **Don't quote large excerpts verbatim.** Summarize in your own words; generate original example documents that exercise the same rules. Quote only minimal grammar fragments with attribution (section/page).
- **Don't invent.** If the spec is silent, mark it unspecified. If it contradicts itself, add a `> **Ambiguity:**` callout so the implementer can decide.

## Next step for the user

Once extraction is complete, the typical next step is to scaffold the package with `/new-go-text-file-library <name>` and then invoke the `implement-go-text-file-library` agent. That agent partitions `SPEC.md` by its `##` headings into `_spec_tokens.md` / `_spec_grammar.md` / `_spec_examples.md` and dispatches tokenizer / parser / printer subagents.
