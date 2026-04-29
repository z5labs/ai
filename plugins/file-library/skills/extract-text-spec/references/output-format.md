# Output format

The `extract-text-spec` skill produces a single `SPEC.md` whose `##` boundaries are load-bearing: the `implement-go-text-file-library` agent partitions on them to feed tokenizer, parser, and printer subagents in isolation. Stick to the section names and order below — renaming `## Lexical Elements (Tokens)` to `## Tokens` will break that partitioning.

## File layout

```
<format-name>/
└── SPEC.md
```

## SPEC.md template

````markdown
# <Format Name> Specification Reference

## Overview

One paragraph: what the format is for, the version covered, and the governing
standard (RFC number, ISO number, vendor doc URL).

## Lexical Elements (Tokens)

Everything the tokenizer needs. The tokenizer classifies a stream of characters
into tokens, so this section must be exhaustive — every byte the format
recognizes belongs to some token class described here.

For each token class, give:
- **Name** — what the implementer will call it (PascalCase token type, e.g.
  `StringLiteral`, `RecordSeparator`).
- **Pattern** — exact syntax. Use a regex or grammar fragment when feasible;
  call out delimiters, escape sequences, and required vs. optional surroundings.
- **Examples** — at least one concrete example per token class. Examples cost
  almost nothing and save the implementer from guessing edge cases.
- **Edge cases** — empty matches, max length, encoding, line continuations, any
  surprising whitespace or newline behavior.

### Comments

Delimiters, nesting rules, placement rules. State whether comments may appear
inside string literals, inside multi-line constructs, etc. If the format has
no comments, say so explicitly — silence is read as "I forgot to ask".

### Whitespace and Delimiters

Significant vs. ignorable whitespace. Field separators, statement terminators,
record terminators, indentation rules, line-continuation behavior. State the
exact characters (`U+0020`, `U+0009`, `\r\n` vs. `\n`).

### Literals

String, number, boolean, null, date, regex — whatever the format supports. Per
literal type:
- Quoting / delimiter rules (single, double, backtick, raw, triple-quoted)
- Escape sequences with their interpretations
- Allowed numeric formats (decimal, hex, octal, binary, underscores, exponent)
- Whether literals can span lines

### Keywords and Reserved Words

The fixed identifiers that cannot be used as user-defined names. Include the
case-sensitivity rule (e.g. SQL is case-insensitive for keywords; JSON is
case-sensitive for `true`/`false`/`null`).

### Symbols and Operators

Structural punctuation (braces, brackets, colons, commas, arrows) plus any
operators. State the role of each symbol — `:` in JSON is a key-value
separator; `:` in Python type hints is something else entirely.

## Structure (Grammar)

Everything the parser needs. Describe the grammar in terms of the tokens above.

### Top-Level Structure

What a complete valid document looks like at the outermost level. State the
file's start and end conditions and any document-level constraints (single
root, optional BOM, trailing newline rules).

### Grammar Productions

For each production in the grammar, give:
- **Name** — PascalCase, matching the spec's name where reasonable.
- **Production** — EBNF / ABNF / BNF. Use the notation the source spec uses;
  if the spec is informal prose, distill it into EBNF. Keep production names
  verbatim from the spec — the implementer agent matches them to AST types.
- **Members / Fields** — name, type (another production or a token class),
  multiplicity (`?`, `*`, `+`), ordering rules, separator rules.
- **Nesting rules** — what can appear inside what. Recursion is normal but
  must be explicit.
- **Constraints** — value restrictions, required fields, uniqueness rules,
  ordering rules that the grammar can't express by itself.

### Ordering and Optionality

Cross-cutting rules: which constructs must appear before others (e.g.
TOML's `[[table]]` rules, JSON's strict member ordering=none), which are
required vs. optional, and what defaults apply when omitted.

## Semantics

Meaning and interpretation rules that affect AST shape and printer output —
not just whether something parses, but what it means.

- Type coercion or default values
- Inheritance, composition, or include rules
- Cross-references between elements (anchors, IDs, $ref)
- Validation rules beyond syntax (uniqueness, exclusivity)
- Equivalence rules — are `1.0` and `1` the same number? are duplicate keys an
  error or a last-wins overwrite?

## Examples

At least three complete, valid documents in the format. These become test
fixtures for tokenizer / parser / printer round-trips, so they need to
**actually parse** under the grammar above.

### Minimal Valid File

The smallest valid document. Often a single token or empty document if the
format permits.

### Typical File

A realistic document with the common features a real user would write.

### Complex File

A document that exercises edge cases, nested structures, escapes, comments,
and whatever else makes implementing this format hard. This is the example
that catches bugs.

## Appendix

- Character encoding requirements (UTF-8, ASCII, must-not-have-BOM, etc.)
- Size limits or implementation notes
- Version differences if the spec covers multiple versions
- Related standards / RFCs / registries
````

## Important rules (apply to every section)

- **Every `##` heading must appear in the order shown.** The downstream
  partitioner finds them by exact string match. Add subsections inside each
  `##` freely — but don't rename or reorder the top level.
- **Be precise about syntax.** "Strings use double quotes" is not enough —
  specify whether they span lines, what escapes are supported, whether empty
  strings are allowed, whether there's a maximum length.
- **Capture every token class.** The tokenizer must classify every character
  sequence. If the spec mentions a syntax element, it needs a row in
  `## Lexical Elements`.
- **Use formal grammar.** EBNF or ABNF — whichever the spec uses. Productions
  drive parser actions; paraphrase rots into ambiguity at implementation time.
- **Stable terminology.** Pick one word per concept ("field" or "member", not
  both; "production" or "rule", not both) and stick to it across sections.
- **Don't quote large excerpts verbatim.** Summarize in your own words and
  generate original example documents that exercise the same syntactic rules.
  Quote minimal grammar fragments with attribution (section/page) only.
- **Flag ambiguities.** When the spec is unclear or contradictory, write a
  `> **Ambiguity:**` blockquote so the implementer can make a documented
  decision rather than re-discovering the gap during coding.
- **Don't invent.** If the spec is silent, mark it unspecified. The
  implementer agent reads "unspecified" as "ask the user"; absence is read as
  "this didn't exist", which is worse.
