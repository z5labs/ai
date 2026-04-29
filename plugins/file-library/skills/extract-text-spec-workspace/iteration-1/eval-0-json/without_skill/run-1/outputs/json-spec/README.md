# JSON Specification Reference (RFC 8259)

Source: RFC 8259, "The JavaScript Object Notation (JSON) Data Interchange
Format", T. Bray (Ed.), December 2017.
<https://www.rfc-editor.org/rfc/rfc8259.txt>

This directory is a focused extract for implementing a Go tokenizer/parser/
printer for JSON. It covers:

- Literal forms: strings (with escapes), numbers, `true`/`false`/`null`
- Structural tokens: `{`, `}`, `[`, `]`, `:`, `,`
- Grammar productions for objects, arrays, and members

The I-JSON profile and security considerations were intentionally excluded
per request. Other RFC sections not directly relevant to the
tokenize/parse/print pipeline (IANA registration, examples, references,
change history) are also omitted.

ABNF grammar productions (per RFC 5234) below are reproduced verbatim from
RFC 8259. Surrounding prose is summarized.

## Files

- `grammar.abnf` — All ABNF productions in a single file, suitable as a
  test fixture or for feeding an ABNF-aware tool.
- `01-grammar-overview.md` — Top-level grammar (`JSON-text`), whitespace,
  and the six structural-character productions (Section 2).
- `02-values.md` — Literal names `false` / `null` / `true` and the
  `value` production (Section 3).
- `03-objects.md` — Object and member productions (Section 4).
- `04-arrays.md` — Array production (Section 5).
- `05-numbers.md` — Number production and rules (Section 6).
- `06-strings.md` — String production, escapes, and surrogate pairs
  (Section 7).
- `07-character-encoding.md` — UTF-8 requirement, BOM rules, unpaired
  surrogates, and string comparison (Sections 8.1-8.3).
- `08-parsers-and-generators.md` — Parser and generator conformance
  requirements (Sections 9-10).
