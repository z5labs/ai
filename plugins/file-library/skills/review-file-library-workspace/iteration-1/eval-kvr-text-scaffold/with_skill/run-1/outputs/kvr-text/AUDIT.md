# Audit: kvr-text (text file library)

**Date:** 2026-05-03
**Spec:** SPEC.md (158 lines)
**Tests:** PASS

## Summary

- 28 findings across 3 phases / 7 categories
- Phases: tokenizer (10), parser (11), printer (7)
- Severity: blockers (18), warnings (7), info (3)

## Tokenizer findings

### Missing token types

- **[blocker]** SPEC.md § Lexical Elements / Identifier (lines 23-31) — `tokenizer.go` declares `TokenIdentifier` but the dispatch in `tokenize` (lines 117-128) never produces it; no specialised action for ASCII letter/underscore-starting runs exists.
- **[blocker]** SPEC.md § Lexical Elements / Symbol (lines 33-44) — `TokenSymbol` constant exists but `tokenize` (lines 117-128) has no branch that recognises any of `=`, `{`, `}`, `;` and yields a `TokenSymbol`.
- **[blocker]** SPEC.md § Lexical Elements / String (lines 46-56) — `TokenString` constant exists but `tokenizer.go` has no string-literal action: nothing scans for the opening `"`, decodes `\\`, `\"`, `\n`, `\t`, or rejects literal newlines with `UnterminatedStringError`. `UnterminatedStringError` itself is not declared in `tokenizer.go`.
- **[blocker]** SPEC.md § Lexical Elements / Number (lines 58-66) — `TokenNumber` constant exists but `tokenize` (lines 117-128) has no digit-run action.
- **[blocker]** SPEC.md § Lexical Elements / Comment (lines 68-75) — `TokenComment` constant exists but no action in `tokenizer.go` recognises `#`, strips leading horizontal whitespace, or stops at newline/EOF.
- **[info]** SPEC.md § Lexical Elements / Invalid (lines 77-79) — `TokenInvalid` is declared as the zero value in `tokenizer.go` line 20 and intentionally never produced; matches the spec's "sentinel only" contract.

### Drift

- **[blocker]** `tokenizer.go tokenize` (lines 117-128) — top-level dispatch is a stub that consumes a single rune and returns `nil` regardless of input, dropping every non-EOF rune on the floor and emitting zero tokens. SPEC.md § Lexical Elements (line 21) requires six token types to be produced; current behavior produces none, so every spec-defined token class is unreachable. (see also: every category above)
- **[warning]** `tokenizer_test.go TestTokenizer` (lines 26-51) — the only test case is `empty_input_yields_no_tokens`. None of the spec-defined token classes (Identifier, Symbol, String, Number, Comment) has a test that exercises it; per checklist § Tokenizer / Drift, untested behavior is unverifiable drift.
- **[warning]** `tokenizer_test.go` — no test pins the `Pos{Line, Column}` 1-based column tracking that SPEC.md § Lexical Elements (line 21) requires every token to carry; the position-tracking helpers in `tokenizer.go` (`next`/`backup`, lines 83-109) are therefore unverified.
- **[warning]** `tokenizer.go` — no `UnterminatedStringError` type is declared, even though SPEC.md § Lexical Elements / String (line 56) names it as the typed error a literal-newline-in-string condition must produce. CLAUDE.md (lines 60-63) mandates typed errors with `errors.As` callers, so this gap blocks spec-compliant string scanning.

## Parser findings

### Grammar gaps

- **[blocker]** SPEC.md § Structure (Grammar) / `File = { Statement } .` (line 86) — `parser.go File` (line 11) is `struct{}`; it has no field for the `Statement` repetition (e.g. no `Statements []Statement` or pair of `Records []Record` + `Blocks []Block`). The grammar's top-level repetition is therefore unrepresentable in the AST.
- **[blocker]** SPEC.md § Structure (Grammar) / `Statement = Comment* ( Record | Block ) .` (line 87) — no `Statement` AST node and no `parseStatement` action exist in `parser.go`. The dispatch hub for "comments-then-record-or-block" is missing.
- **[blocker]** SPEC.md § Structure (Grammar) / `Record = "record" Type Identifier "=" Value .` (line 89) — no `Record` Go type in `parser.go` (no fields for `Type`, `Key`, `Value`, or the `LeadingComments []string` mandated by SPEC.md § Comments are statements with attachment, line 110); no `parseRecord` action chain.
- **[blocker]** SPEC.md § Structure (Grammar) / `Block = "block" Identifier "{" { Statement ";" } "}" .` (line 90) — no `Block` Go type in `parser.go` and no `parseBlock` action chain. Per CLAUDE.md § The inner action loop rule (lines 32-46), `Block` is exactly the kind of complex/nested type that requires an inner action loop; there is no scaffold for one.
- **[blocker]** SPEC.md § Structure (Grammar) / `Type = "string" | "number" .` (line 91) — no Go representation (e.g. `type RecordType int` with `TypeString` / `TypeNumber` constants, or a string field validated at parse time). The `Type` interface declared in `parser.go` (lines 13-17) is a marker for AST nodes — unrelated to the grammar's `Type` non-terminal — so the grammar element is entirely unimplemented.
- **[blocker]** SPEC.md § Structure (Grammar) / `Value = TokenString | TokenNumber .` (line 92) — no `Value` AST representation in `parser.go`; no action consumes a `TokenString` or `TokenNumber` as a record value.
- **[blocker]** SPEC.md § Structure (Grammar) / `Comment = TokenComment .` (line 88) and § Comments are statements with attachment (line 110) — no parser logic accumulates a run of `TokenComment`s into `LeadingComments` for the next non-comment statement, and there is no AST field to hold them. (see also: Drift / Semantics)
- **[blocker]** `parser.go parseFile` (lines 71-75) — the top-level action is a stub returning `(nil, nil)` immediately; it never reads a token, dispatches, or builds anything, so every grammar production above is unreachable.
- **[blocker]** SPEC.md § Semantics / Type/value agreement (line 118) — no `TypeMismatchError{Type, Got}` typed error is declared in `parser.go`, despite the spec naming it as the parse-time rejection for `record string K = 42`.

### Drift

- **[warning]** `parser_test.go TestParser` (lines 10-35) — the only test is `empty_input_yields_zero_file`. Every grammar production listed above is untested, so any future implementation drift between grammar and code is invisible. CLAUDE.md § Testing (line 55) marks the empty-input case as the *only* allowed exception to the "drive the public `Parse()`" rule, and that single allowed exception is the entire test surface today.
- **[info]** `parser.go Type` interface (lines 13-17) — name collides with the grammar non-terminal `Type` (`Type = "string" | "number"`, SPEC.md line 91). Once the grammar's `Type` lands as a Go type, the AST-node marker interface will need renaming (e.g. `Node`) to avoid a confusing two-meaning identifier; flag now so the rename happens before consumers depend on the marker.

## Printer findings

### Missing/incomplete printer rules

- **[blocker]** SPEC.md § Structure (Grammar) / `Record` (line 89) and § Examples / Typical (lines 126-134) — no `printRecord` action in `printer.go`. `printFile` (lines 38-41) is a stub that returns `nil` immediately; the package cannot reproduce any non-empty input. (Caveat: there is no `Record` AST type yet either — see parser phase — but the printer rule is independently missing.)
- **[blocker]** SPEC.md § Structure (Grammar) / `Block` (line 90) and § Examples / Complex (lines 136-146) — no `printBlock` action in `printer.go`. The spec requires `;` between every inner statement *and* after the last (line 101), and the printer must emit it; nothing in `printer.go` does.
- **[blocker]** SPEC.md § Comments are statements with attachment (line 110) and § Examples / Round-trip (lines 148-158) — no printer logic emits `LeadingComments` before a record or block. Round-trip preservation of leading comments is a load-bearing requirement (the spec carves out a dedicated example for it) and the current printer cannot satisfy it.
- **[info]** `printer.go printFile` (lines 38-41) — the stub correctly handles the empty-file case (SPEC.md § Examples / Minimal, lines 122-124) by emitting nothing; this is the one rule the printer does satisfy today.

### Round-trip test coverage

- **[warning]** `printer_test.go TestPrinterRoundTrip` (lines 39-68) — only the `empty_source_round_trips` case is exercised. Per CLAUDE.md § Testing (line 56), every printer rule must have a round-trip test; once `printRecord` exists it will need one, once `printBlock` exists it will need one, and SPEC.md § Examples / Round-trip (lines 148-158) explicitly demands round-trip coverage for leading comments on top-level records. None of these three round-trip cases exist.
- **[warning]** `printer_test.go TestPrinter` (lines 11-37) — only the `empty_file_prints_empty_string` direct test case exists. CLAUDE.md § Testing (line 56) requires both a direct test *and* a round-trip test per printer rule; the direct-test surface for `Record` and `Block` is empty.
- **[warning]** `printer_test.go` — no test exercises the block-internal `;` separator behavior pinned by SPEC.md § Blocks (lines 99-106), including the "trailing `;` required before `}`" rule. Once `printBlock` exists this is the cheapest place a printer/parser asymmetry could hide.

### Drift

- (none) — with `printFile` stubbed to no-op, there is no formatted output to compare against SPEC.md § Examples; drift cannot be assessed until a real printer rule exists. Any deviation surfaces under "Missing/incomplete printer rules" above.
