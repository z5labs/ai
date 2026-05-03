# Text package audit checklist

Each phase subagent emits findings to `<package>/_audit_<phase>.md` using the format and categories below. The orchestrator concatenates these into `AUDIT.md` without further editing, so the headings and finding-line shape here are the durable contract.

## Finding-line format (all phases)

Every finding is one bullet. The bullet starts with a **severity prefix** so the orchestrator can grep counts:

- `- **[blocker]**` — spec mandates X, implementation does not have X. Without this, the package fails to handle inputs the spec defines as valid.
- `- **[warning]**` — implementation handles X but in a way that differs from the spec, OR a test category the package needs is missing (e.g., no round-trip test for a printer rule).
- `- **[info]**` — observation worth recording but not necessarily a defect (e.g., implementation supports more than the spec requires; consider documenting the extension).

After the prefix, cite both sides of the comparison so the reader can jump straight to either:

```
- **[blocker]** SPEC.md § Lexical Elements / String literals (lines 45–80) — no `TokenString` constant in `tokenizer.go`
- **[warning]** `printer.go printRecord` (line 120) — direct test exists, no round-trip test in `printer_test.go`
- **[info]** `tokenizer.go` defines `TokenComment` but SPEC.md `## Lexical Elements` does not mention comments — likely an undocumented extension
```

If a finding spans multiple categories (e.g., a missing token type also breaks a parser rule), cite it under the first category and add `(see also: <category>)` rather than duplicating.

## Per-phase output skeleton

Each `_audit_<phase>.md` file uses this exact skeleton — empty categories must still appear with a single bullet `- (none)` so the orchestrator's per-category grep is reliable:

```
## <Phase> findings

### <Category 1>
- ...

### <Category 2>
- ...
```

The orchestrator does not re-order or edit these — the headings here are what the reader sees in `AUDIT.md`.

---

## Tokenizer phase

**Source files to read:** `tokenizer.go`, `tokenizer_test.go`.
**Spec sections received:** Overview, Lexical Elements (Tokens) and all subsections, Examples.

### Categories

#### Missing token types
Cross-reference every distinct token class named in the Lexical Elements section against the `TokenType` constant block in `tokenizer.go`. Categories named in the spec but not represented as a constant are blockers. A spec sub-section like "String literals" is a token class even if the spec doesn't explicitly say "TokenString" — match by concept, not name.

Also flag token types that exist in `tokenizer.go` but are never produced by any action — they're either drift (token class was retired in the spec) or dead code.

#### Drift
Token types whose semantics in `tokenizer.go` (the action that produces them) differ from the spec — wrong character class, wrong escape rules, wrong position handling. Read the relevant tokenizer action and compare to the spec rule. Tests that pin the wrong behavior count as evidence of drift, not against it — a passing test for behavior the spec rejects is the clearest drift signal.

If `tokenizer_test.go` lacks a test that exercises a spec-defined token class, list a `[warning]` in this category — untested behavior is unverifiable drift.

---

## Parser phase

**Source files to read:** `parser.go`, `parser_test.go`.
**Spec sections received:** Overview, Structure (Grammar), Semantics, Examples.

### Categories

#### Grammar gaps
For every production in the Structure (Grammar) section, confirm `parser.go` has a corresponding `parserAction[*T]` (or equivalent action chain). Missing productions are blockers. Productions implemented as flat for-with-switch instead of an inner action loop are warnings — the implement skill mandates the inner action loop pattern for any complex/nested type, so divergence from that pattern is real drift even if the package compiles.

Also check that AST node types named (or implied) in the grammar exist as Go types in `parser.go`. A grammar rule like `Block ::= "{" Record* "}"` implies a `Block` AST node holding `Records []Record`; if `parser.go` omits the type, that's a blocker.

#### Drift
Productions implemented in `parser.go` whose accepted/rejected inputs differ from the spec — extra optional terminals, missing terminals, wrong associativity, wrong precedence. The Semantics section of the spec is the source of truth for semantic-side drift (e.g., "comments attach to the next node, not the previous one").

If `parser_test.go` doesn't drive `Parse()` for its expected values (constructs AST literals directly), flag as `[warning]` — the implement skill calls this out as a hard rule because hand-built AST expectations let drift hide.

---

## Printer phase

**Source files to read:** `printer.go`, `printer_test.go`.
**Spec sections received:** Overview, Structure (Grammar), Semantics, Examples.

### Categories

#### Missing/incomplete printer rules
For every AST node type defined in `parser.go`, confirm `printer.go` has a printer action (or branch) that emits it. Missing printer actions are blockers — the package can parse inputs it cannot reproduce, breaking round-trip.

Incomplete rules (printer action exists but omits a field documented in the spec — e.g., a `Record` printer that drops `LeadingComments`) are blockers too: a non-round-trippable printer is a defect, not a stylistic choice.

#### Round-trip test coverage
For every AST node type, confirm `printer_test.go` has a `Parse → Print → Parse → require.Equal` test that exercises that node. Missing round-trip tests are warnings. The implement skill mandates round-trip tests for every printer method because direct tests pin formatting choices but cannot detect printer/parser asymmetry; this category is where that mandate is enforced after the fact.

If only direct tests exist (no round-trip), flag the specific AST node as `[warning] no round-trip test`. If the round-trip exists but only for a trivial case (empty input), flag as `[warning] round-trip exists but does not exercise <field>`.

#### Drift
Printer rules whose output formatting differs from what the spec's Examples section shows — wrong separator, wrong indentation rule, wrong comment placement. The Examples section is the cheapest reference for end-to-end formatting expectations.

Rules that emit valid output the spec doesn't show (extra trailing newline, stylized separator) are `[info]` not `[warning]` — round-trip will catch real defects, and stylistic extensions are often deliberate.

---

## Test-status integration

The orchestrator passes the result of `go test -race ./...` to every phase subagent. If tests are failing:

- The first ~10 lines of failure output is in the test-status header at the top of `AUDIT.md`.
- Each phase subagent must scan failing test names for tests that belong to its phase (e.g., the tokenizer phase scans for `TestTokenizer*`) and add a `[blocker]` finding under the relevant category referencing the failing test by name. A failing test is direct, runtime-verified drift evidence — the cheapest finding the audit produces.

If tests pass, no test-failure findings are added; the test-status header in `AUDIT.md` is the only mention.

## What the audit does not do

- Does not propose fixes — findings cite the gap, not the patch.
- Does not edit the spec, source, or tests.
- Does not run benchmarks or coverage tools.
- Does not score the package — every finding stands on its own; no aggregate "grade".
