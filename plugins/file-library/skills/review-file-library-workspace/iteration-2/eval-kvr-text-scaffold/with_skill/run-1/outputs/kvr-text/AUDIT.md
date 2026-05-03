# Audit: kvr-text (text file library)

**Date:** 2026-05-03
**Spec:** SPEC.md (158 lines)
**Tests:** PASS

## Summary

- 38 findings across 7 categories
- Phases: tokenizer (13), parser (13), printer (12)
- Severity: blockers (34), warnings (0), info (4)

## Tokenizer findings

### Missing token types

- **[blocker]** SPEC.md § Lexical Elements / Identifier (lines 23-31) — `TokenIdentifier` constant is declared in `tokenizer.go` (line 21) but no tokenizer action produces it: `tokenize` (line 117) is a stub that consumes one rune and returns `nil`, so identifiers are never yielded.
- **[blocker]** SPEC.md § Lexical Elements / Symbol (lines 33-44) — `TokenSymbol` constant is declared in `tokenizer.go` (line 22) but no action recognises `=`, `{`, `}`, or `;`; the `tokenize` dispatch in `tokenizer.go` (lines 117-128) has no symbol branch.
- **[blocker]** SPEC.md § Lexical Elements / String (lines 46-56) — `TokenString` constant is declared in `tokenizer.go` (line 23) but no action recognises double-quoted strings; backslash-escape decoding for `\\`, `\"`, `\n`, `\t` is not implemented anywhere; the `UnterminatedStringError` typed error mandated for literal-newline-in-string is missing from `tokenizer.go`.
- **[blocker]** SPEC.md § Lexical Elements / Number (lines 58-66) — `TokenNumber` constant is declared in `tokenizer.go` (line 24) but no action recognises digit runs; `tokenize` (lines 117-128) does not branch on digit characters.
- **[blocker]** SPEC.md § Lexical Elements / Comment (lines 68-75) — `TokenComment` constant is declared in `tokenizer.go` (line 25) but no action recognises `#`-prefixed line comments; the leading-horizontal-whitespace stripping rule from the spec (line 70) has no implementation.
- **[info]** SPEC.md § Lexical Elements / Invalid (lines 77-79) — `TokenInvalid` is correctly declared as the iota zero value in `tokenizer.go` (line 20) and intentionally never produced; this matches the spec's "sentinel for uninitialised" requirement.

### Drift

- **[blocker]** SPEC.md § Lexical Elements / Identifier (lines 23-31) — `tokenizer_test.go` has no test that exercises `TokenIdentifier` (only `empty_input_yields_no_tokens`); identifier tokenization is unverified, so per the checklist's "untested behavior is unimplemented" rule this is a blocker.
- **[blocker]** SPEC.md § Lexical Elements / Symbol (lines 33-44) — `tokenizer_test.go` has no test that exercises any of the four symbols `= { } ;`; symbol tokenization is unverified.
- **[blocker]** SPEC.md § Lexical Elements / String (lines 46-56) — `tokenizer_test.go` has no test for string literals: undecorated content, the `\\` / `\"` / `\n` / `\t` escapes (lines 50-54), or the literal-newline `UnterminatedStringError` ambiguity rule (line 56) are all untested.
- **[blocker]** SPEC.md § Lexical Elements / Number (lines 58-66) — `tokenizer_test.go` has no test for digit runs; numeric tokenization is unverified.
- **[blocker]** SPEC.md § Lexical Elements / Comment (lines 68-75) — `tokenizer_test.go` has no test for `#`-prefixed comments; the leading-whitespace-stripping rule (line 70) and trailing-newline-not-in-Value rule are unverified.
- **[blocker]** SPEC.md § Overview (line 21) — every token must carry a 1-based `Pos{Line, Column}` of its first rune; `tokenizer_test.go` has no test that asserts token `Pos` for any non-trivial input, so position-tracking correctness across newlines (the `next()` / `backup()` interaction in `tokenizer.go` lines 83-109) is unverified.
- **[blocker]** SPEC.md § Examples (lines 120-158) — none of the spec's four example inputs (minimal, typical, complex block, round-trip) are exercised by `tokenizer_test.go`; the spec's most direct sanity check on user-facing behaviour is unused.

## Parser findings

### Grammar gaps

- **[blocker]** SPEC.md § Structure (Grammar) / `File = { Statement } .` (lines 86-87) — `parser.go` defines `File` as an empty struct (`type File struct{}` line 11); the spec mandates `File` hold a sequence of statements (records and blocks), so the type lacks a `Statements` / `Records` / `Blocks` field entirely. `parseFile` (lines 71-75) is a stub returning `(nil, nil)` with no dispatch on token types.
- **[blocker]** SPEC.md § Structure (Grammar) / `Statement = Comment* ( Record | Block ) .` (line 87) and § Comments are statements with attachment (lines 108-110) — no `Statement` AST node or interface exists in `parser.go`; the spec's comment-attachment model (a run of `Comment*` attached to the following `Record` or `Block`) has no representation.
- **[blocker]** SPEC.md § Structure (Grammar) / `Record = "record" Type Identifier "=" Value .` (line 89) and § Records (lines 95-97) — no `Record` AST type exists in `parser.go`; the spec also requires a `LeadingComments []string` field (line 110) which therefore does not exist either. No `parseRecord` action exists.
- **[blocker]** SPEC.md § Structure (Grammar) / `Block = "block" Identifier "{" { Statement ";" } "}" .` (line 90) and § Blocks (lines 99-106) — no `Block` AST type exists in `parser.go`; no `Records []Record` field, no `LeadingComments` field. No `parseBlock` action exists. The "trailing `;` required before closing `}`" rule (line 101 + illegal example line 105) has no implementation.
- **[blocker]** SPEC.md § Structure (Grammar) / `Type = "string" | "number" .` (line 91) — no `Type` enum / constants for the two valid type names exist in `parser.go`; the file declares a marker interface `type Type interface { isType() }` (lines 13-17) but it has no implementations and is unrelated to record-type-name handling.
- **[blocker]** SPEC.md § Structure (Grammar) / `Value = TokenString | TokenNumber .` (line 92) — no `Value` AST node or union exists in `parser.go`; the type/value-agreement check from § Semantics (line 118) cannot be performed without it.
- **[blocker]** SPEC.md § Semantics / Type/value agreement (line 118) — the spec mandates a typed `TypeMismatchError{Type, Got}` carrying the value-token position; this typed error is not declared in `parser.go`.
- **[info]** `parser.go` declares a `Type` marker interface (lines 13-17) that does not appear in the spec's grammar — likely a scaffolding placeholder; once real AST nodes are added this name will collide with the spec's `Type = "string" | "number"` and should be renamed or removed.

### Drift

- **[blocker]** SPEC.md § Examples / Minimal (lines 122-124) — `parser_test.go` only covers `empty_input_yields_zero_file`; per the checklist, a parser test that *only* covers the empty-input scaffold case for grammar productions the spec defines is effectively no test. Every production in lines 86-92 (`File`, `Statement`, `Record`, `Block`, `Type`, `Value`, `Comment`) is therefore unverified.
- **[blocker]** SPEC.md § Examples / Typical (lines 126-134) — no parser test exercises the two-records-with-leading-comment example; comment attachment to the immediately-following record (the rule from line 110) is untested.
- **[blocker]** SPEC.md § Examples / Complex block (lines 136-146) — no parser test exercises a block containing records with an inner comment; block parsing (including the `;` separator) is untested.
- **[blocker]** SPEC.md § Semantics / Whitespace fidelity (line 117) — no parser test asserts that blank lines and column positions are *not* preserved while comments and structural content are; the semantic guarantee is unverified.
- **[blocker]** SPEC.md § Semantics / Identifier case sensitivity (line 114) and Record key uniqueness (line 115) — no parser test exercises duplicate keys or differing-case identifiers; the documented semantics are unverified.

## Printer findings

### Missing/incomplete printer rules

- **[blocker]** SPEC.md § Structure (Grammar) / `File = { Statement } .` (lines 86-87) — `printer.go` `printFile` (lines 38-41) is a stub returning `nil` immediately; no iteration over file statements exists. Even once `parser.go` gains the missing `Statements`/`Records`/`Blocks` fields, no printer action will visit them.
- **[blocker]** SPEC.md § Structure (Grammar) / `Record` (line 89) and § Examples / Typical (lines 126-134) — `printer.go` has no `printRecord` action; the `record <type> <ident> = <value>` form has no emission rule.
- **[blocker]** SPEC.md § Structure (Grammar) / `Block` (line 90) and § Examples / Complex (lines 136-146) — `printer.go` has no `printBlock` action; block opening, inner-record indentation (the spec example uses 4-space indent at line 140), the `;` separator after each inner statement, and `}` closing are all unimplemented.
- **[blocker]** SPEC.md § Comments are statements with attachment (lines 108-110) — `printer.go` has no rule for emitting `LeadingComments` (the field itself is also missing from `parser.go`); a non-round-trippable printer for attached comments is a defect, not a stylistic choice.
- **[blocker]** SPEC.md § Examples / Complex (lines 136-146) — no rule emits the `;` separator between inner block statements or the trailing `;` before `}` (illegal-without-it per line 105); printer cannot reproduce a legal block at all.

### Round-trip test coverage

- **[blocker]** SPEC.md § Examples / Round-trip (lines 148-156) — `printer_test.go` `TestPrinterRoundTrip` (lines 39-68) only exercises the empty-source case (`source: ""`); the spec's explicit round-trip example (two records, each with a leading comment) has no test. Per the checklist, a round-trip test that exercises only a trivial case is treated the same as no round-trip test for the AST nodes it omits — `Record`, `Block`, and `LeadingComments` are all in the same untested state.
- **[blocker]** SPEC.md § Structure (Grammar) / `Record` (line 89) — no round-trip test covers `Record`; the AST node also does not yet exist in `parser.go`, but per the checklist's "untested behavior is unimplemented" rule the gap is a blocker even before the type is introduced.
- **[blocker]** SPEC.md § Structure (Grammar) / `Block` (line 90) — no round-trip test covers `Block`; trailing-`;` requirement and inner-statement indentation (visible in spec example lines 139-143) are unverified.
- **[blocker]** SPEC.md § Comments are statements with attachment (lines 108-110) — no round-trip test asserts that a `LeadingComments` run survives `Parse → Print → Parse`; this is the cheapest test of the spec's most subtle requirement and is missing.
- **[blocker]** SPEC.md § Semantics / Whitespace fidelity (line 117) — no round-trip test verifies that blank lines and column positions are dropped while comments and structural content survive; the semantic guarantee is unverified.

### Drift

- **[info]** `printer.go` (lines 38-41) — `printFile` returning `nil` for an empty file is consistent with SPEC.md § Examples / Minimal (line 124: "`Print(&File{})` writes nothing"). The only behaviour the printer currently implements is correct; all subsequent drift findings will become observable once the missing rules are added.
- **[info]** SPEC.md § Examples / Complex (lines 136-146) shows 4-space indentation inside blocks; once `printBlock` is implemented, divergence from this indentation would be drift to flag here. Currently no rule exists to drift, so no drift finding is recorded.
