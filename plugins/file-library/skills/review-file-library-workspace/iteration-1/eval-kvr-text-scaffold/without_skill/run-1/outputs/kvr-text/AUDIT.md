# kvr-text Audit Report

Audit of the freshly scaffolded `kvr` package against `SPEC.md`.

## Summary

The package is a scaffold: the token-type enum and the public surface
(`Tokenize`, `Parse`, `Print`, `File`) exist, but every dispatch function is a
stub. None of the lexical recognisers, no grammar productions, and no print
rules are wired up. The AST has no record/block/comment node types at all, and
the test suite covers only the empty-input case.

This audit treats the scaffold as the starting point — every gap below is an
implementation task that must be completed before the package can satisfy the
spec.

| Category                              | Findings |
|---------------------------------------|----------|
| Missing token types / lexical rules   | 7        |
| Grammar gaps (AST + parser)           | 9        |
| Missing or incomplete printer rules   | 6        |
| Round-trip / test coverage gaps       | 6        |
| Drift between spec and implementation | 3        |
| **Total**                             | **31**   |

---

## 1. Missing token types / lexical rules

The `TokenType` enum (`tokenizer.go:19-26`) declares all six spec token types,
but the dispatch function `tokenize` (`tokenizer.go:117-128`) is a stub: it
reads one rune and returns `nil`. None of the per-kind recognisers exist.

### 1.1 Identifier recogniser missing
- **Spec**: § Identifier (`SPEC.md:23-31`) — letter/underscore start, then
  letters/digits/underscores; `record` and `block` are still `TokenIdentifier`.
- **Impl**: `tokenizer.go:117-128` (`tokenize`) — no branch for letter/`_`; no
  `lexIdentifier` action.
- **Impact**: nothing parses — `record`, `block`, `string`, `number` all
  arrive as the keywords the grammar checks for, and they cannot be produced.

### 1.2 Symbol recogniser missing
- **Spec**: § Symbol (`SPEC.md:33-44`) — `=`, `{`, `}`, `;` each yield a
  one-rune `TokenSymbol`.
- **Impl**: `tokenizer.go:117-128` — no branch for any of the four symbols.
- **Impact**: every grammar production (record `=`, block `{`/`}`, separator
  `;`) is unreachable.

### 1.3 String recogniser missing
- **Spec**: § String (`SPEC.md:46-56`) — `"`-delimited; escapes `\\ \" \n \t`;
  decoded value (no quotes/escapes); literal newline rejected with
  `UnterminatedStringError` carrying the opening-quote position.
- **Impl**: `tokenizer.go:117-128` — no `"` branch; no `lexString` action; no
  `UnterminatedStringError` type defined anywhere in `tokenizer.go`.
- **Impact**: string values cannot be tokenised; required typed error is
  missing for `errors.As` callers.

### 1.4 Number recogniser missing
- **Spec**: § Number (`SPEC.md:58-66`) — one or more ASCII digits; value is
  the digit text, parser converts later.
- **Impl**: `tokenizer.go:117-128` — no digit branch; no `lexNumber` action.
- **Impact**: `record number ANSWER = 42` cannot be tokenised.

### 1.5 Comment recogniser missing
- **Spec**: § Comment (`SPEC.md:68-75`) — starts at `#`, runs to newline or
  EOF; leading horizontal whitespace and the `#` are stripped from `Value`;
  trailing newline is stripped.
- **Impl**: `tokenizer.go:117-128` — no `#` branch; no `lexComment` action.
- **Impact**: comment attachment (the only round-trippable annotation) is
  impossible; `LeadingComments` will always be empty.

### 1.6 Whitespace skipping missing
- **Spec**: Overview (`SPEC.md:7`) — "Whitespace between tokens is
  insignificant except inside quoted strings."
- **Impl**: `tokenize` (`tokenizer.go:117-128`) consumes one rune, does
  nothing with it, and returns. Spaces, tabs, and newlines have no skip
  branch.
- **Impact**: even after recognisers exist, real input cannot stream through
  unless whitespace is absorbed between tokens.

### 1.7 `UnexpectedCharacterError` is declared but never returned
- **Spec**: § Invalid (`SPEC.md:77-79`) — implies any rune the dispatch does
  not want is invalid.
- **Impl**: `UnexpectedCharacterError` is defined (`tokenizer.go:60-67`) but
  never constructed; the stub `tokenize` only forwards `bufio` errors.
- **Impact**: the typed error is dead code until dispatch is wired and falls
  through to a default `yield(&UnexpectedCharacterError{...}, ...)`.

---

## 2. Grammar gaps (AST + parser)

The parser has no AST shapes beyond `File`, and `parseFile` is a stub.

### 2.1 `File` has no children
- **Spec**: § Structure / `File = { Statement }` (`SPEC.md:86`) and § Examples
  (`SPEC.md:122-156`) — a file holds a sequence of records and blocks.
- **Impl**: `parser.go:11` — `type File struct{}` is empty.
- **Impact**: even a successful parse cannot record what it parsed.

### 2.2 No `Record` AST node
- **Spec**: § Records (`SPEC.md:95-97`) — needs Type, Key (Identifier), Value
  fields; § Comments are statements with attachment (`SPEC.md:108-110`) —
  needs `LeadingComments []string`.
- **Impl**: no `Record` type anywhere in `parser.go`.
- **Impact**: records cannot be represented. Round-trip impossible.

### 2.3 No `Block` AST node
- **Spec**: § Blocks (`SPEC.md:99-106`) and § Comments are statements
  (`SPEC.md:108-110`) — needs Name, inner Records, and `LeadingComments`.
- **Impl**: no `Block` type in `parser.go`.
- **Impact**: blocks cannot be represented.

### 2.4 No record-vs-block dispatch in `parseFile`
- **Spec**: § Structure (`SPEC.md:86-93`) — `Statement = Comment* (Record |
  Block)`.
- **Impl**: `parseFile` (`parser.go:71-75`) returns `nil, nil` immediately;
  no switch on `tok.Value` for `record` / `block`.
- **Impact**: every non-empty input parses to `&File{}` silently. The
  existing parser test passes only because it is the empty-input case.

### 2.5 `Record` production not implemented
- **Spec**: `Record = "record" Type Identifier "=" Value` (`SPEC.md:89`) and
  § Records (`SPEC.md:95-97`).
- **Impl**: no `parseRecord` action exists.
- **Impact**: records cannot be parsed.

### 2.6 `Block` production not implemented (incl. trailing `;` rule)
- **Spec**: § Blocks (`SPEC.md:99-106`) — explicitly: "The `;` is required
  between statements inside a block and is also required after the last inner
  statement (i.e. before the closing `}`)."
- **Impl**: no `parseBlock` action exists. CLAUDE.md (`CLAUDE.md`, "inner
  action loop rule") is explicit that this MUST use an inner action loop, not
  a flat for/switch — there is nothing to satisfy that rule against yet.
- **Impact**: blocks cannot be parsed; the trailing-`;` requirement is at
  high risk of being missed when the implementer first writes `parseBlock`.

### 2.7 Comment attachment not implemented
- **Spec**: § Comments are statements with attachment (`SPEC.md:108-110`) —
  "the parser populates [`LeadingComments`] and the printer emits before the
  node's own output"; `Statement = Comment* (Record | Block)` (`SPEC.md:87`);
  example § Round-trip (`SPEC.md:148-156`).
- **Impl**: no comment-buffering state in `parser`; no `LeadingComments`
  field on any node (since neither node exists).
- **Impact**: comments either disappear or become free-floating, breaking the
  round-trip example at `SPEC.md:148-156`.

### 2.8 `TypeMismatchError` not defined or enforced
- **Spec**: § Semantics (`SPEC.md:118`) — "`record string K = 42` ... is
  rejected at parse time with a typed `TypeMismatchError{Type, Got}` carrying
  the value-token position."
- **Impl**: no `TypeMismatchError` type anywhere; no type/value cross-check.
- **Impact**: malformed records parse silently as well-formed (or fail with
  the wrong error type), breaking `errors.As(..., &TypeMismatchError{})`.

### 2.9 Type keyword set not enforced (`string` | `number`)
- **Spec**: `Type = "string" | "number"` (`SPEC.md:91`).
- **Impl**: no enforcement (no parser code yet).
- **Impact**: when `parseRecord` is added, it must reject any other
  identifier in the type slot via `UnexpectedTokenError` (or a typed variant).
  Easy to forget.

---

## 3. Missing or incomplete printer rules

`printFile` (`printer.go:38-41`) is a stub returning `nil`.

### 3.1 No record print rule
- **Spec**: § Records (`SPEC.md:95-97`) and example § Typical
  (`SPEC.md:128-132`) — `record <type> <KEY> = <value>` per line at top
  level.
- **Impl**: `printer.go` has no `printRecord` action.
- **Impact**: records cannot round-trip.

### 3.2 No block print rule
- **Spec**: § Blocks (`SPEC.md:99-106`) and example § Complex
  (`SPEC.md:138-144`) — `block NAME { <stmt>; <stmt>; }`, trailing `;`
  required before closing brace.
- **Impl**: no `printBlock` action.
- **Impact**: blocks cannot round-trip; the trailing-`;` rule is silently
  unenforced on output.

### 3.3 No leading-comment emission
- **Spec**: § Comments are statements with attachment (`SPEC.md:108-110`) —
  printer emits attached comments before the node's own output.
- **Impl**: no comment-emit action; no `LeadingComments` consumer.
- **Impact**: round-trip example § Round-trip (`SPEC.md:148-156`) fails.

### 3.4 No string-value escaping on output
- **Spec**: § String (`SPEC.md:46-56`) — recognised escapes are `\\ \" \n
  \t`. The printer must round-trip the decoded `Value` back into a
  source-legal quoted string (re-escape `"`, `\`, newline, tab).
- **Impl**: no `printValue`; no escape table.
- **Impact**: any string containing `"`, `\`, `\n`, or `\t` corrupts on the
  print side and either fails to re-parse or parses to a different value,
  breaking round-trip.

### 3.5 No statement separator at top level
- **Spec**: Overview (`SPEC.md:7`) — "End-of-line is just whitespace —
  statements end implicitly at the next valid statement opener." Examples
  (`SPEC.md:128-132`, `SPEC.md:148-156`) put each top-level statement on its
  own line.
- **Impl**: no newline emission anywhere in `printer.go`.
- **Impact**: even if record/block printing is added, two consecutive
  records concatenate to `record string A = "1"record number B = 2`, which
  the tokenizer will correctly fail to re-parse (no whitespace between the
  closing quote and `record`).

### 3.6 No empty-block handling defined
- **Spec**: § Blocks (`SPEC.md:99-106`) — does not forbid empty blocks
  syntactically (`{ }` would satisfy `{ Statement ";" }` with zero
  iterations), but does not show one. Spec is ambiguous.
- **Impl**: nothing to look at yet.
- **Impact**: when `printBlock` is added, the implementer needs a decision
  (allow or reject `block X { }`). Flag for spec clarification.

---

## 4. Round-trip / test coverage gaps

All three test files exist, each with exactly one case (empty input). The
non-empty surface is untested.

### 4.1 Tokenizer table covers only the empty input
- **Spec**: § Lexical Elements (`SPEC.md:19-79`) — six token types, each with
  positive examples in the spec.
- **Impl**: `tokenizer_test.go:34-38` — only `empty_input_yields_no_tokens`.
- **Gap**: no case per token type (identifier, symbol, string, number,
  comment), no escape-sequence cases, no `UnterminatedStringError` case, no
  `UnexpectedCharacterError` case, no position-tracking case (multi-line, to
  catch the `prevPos` bug `tokenizer.go:69-79` warns about).

### 4.2 Parser table covers only the empty input
- **Spec**: § Examples (`SPEC.md:122-156`) — minimal, typical, complex,
  round-trip.
- **Impl**: `parser_test.go:18-23` — only `empty_input_yields_zero_file`.
- **Gap**: no record case, no block case, no leading-comment case, no
  `TypeMismatchError` case, no `UnexpectedTokenError` case, no missing-`;`
  case for blocks. CLAUDE.md says parser tests must drive `Parse()` (they
  do — good), but the table is empty of real fixtures.

### 4.3 Printer table covers only the empty file
- **Spec**: § Examples (`SPEC.md:122-156`).
- **Impl**: `printer_test.go:19-23` — only `empty_file_prints_empty_string`.
- **Gap**: no per-rule "direct" test for record, block, leading comment, or
  string-escape formatting. CLAUDE.md is explicit: "Every printer rule gets
  both a direct test and a round-trip test." Currently zero rules ⇒ zero
  tests, but as soon as a rule lands the matching direct test must land too.

### 4.4 Round-trip table covers only the empty source
- **Spec**: § Round-trip (`SPEC.md:148-156`) is given as an explicit
  fixture.
- **Impl**: `printer_test.go:46-49` — only `empty_source_round_trips`.
- **Gap**: the literal round-trip example from the spec is not exercised; no
  round-trip case for blocks, for in-block comments (§ Complex,
  `SPEC.md:138-144`), or for strings containing escape-worthy characters.

### 4.5 No test for `LeadingComments` survival
- **Spec**: § Typical (`SPEC.md:128-134`) and § Round-trip
  (`SPEC.md:148-156`) — both call out that the leading comment must end up on
  the right node.
- **Impl**: no test exercises `LeadingComments` (the field does not exist
  yet, but the test gap remains).
- **Gap**: when comment attachment is implemented, a direct parse-equality
  test should pin which node receives the comment.

### 4.6 No `errors.As` / typed-error tests
- **Spec**: § String (`SPEC.md:56`) requires `UnterminatedStringError` with
  position; § Semantics (`SPEC.md:118`) requires `TypeMismatchError`;
  `tokenizer.go:60-67` declares `UnexpectedCharacterError`; `parser.go:19-36`
  declares `UnexpectedEndOfTokensError` and `UnexpectedTokenError`.
- **Impl**: no test asserts against any of these via `errors.As`.
- **Gap**: CLAUDE.md says "Tests assert via `errors.As` and `errors.Is`" —
  the current suite has no error-shape coverage at all.

---

## 5. Drift between spec and implementation

These are places where what the scaffold *does* declare diverges from, or
under-commits to, what the spec promises.

### 5.1 `UnterminatedStringError` is required by spec but not declared
- **Spec**: § String (`SPEC.md:56`) — "rejected with `UnterminatedStringError`
  carrying the opening-quote position."
- **Impl**: `tokenizer.go` declares only `UnexpectedCharacterError`. There is
  no `UnterminatedStringError` type.
- **Drift**: the named typed error from the spec is absent. This is a hard
  contract — callers will write `var ute *UnterminatedStringError;
  errors.As(err, &ute)` and get a compile error.

### 5.2 `TypeMismatchError` is required by spec but not declared
- **Spec**: § Semantics (`SPEC.md:118`) — `TypeMismatchError{Type, Got}` with
  value-token position.
- **Impl**: not declared in `parser.go`.
- **Drift**: same as 5.1, on the parser side.

### 5.3 Spec promises six token types; enum already enumerates all six but
none are produced
- **Spec**: § Lexical Elements (`SPEC.md:21`) — "KVR has six token types."
- **Impl**: `tokenizer.go:19-26` enumerates `TokenInvalid, TokenIdentifier,
  TokenSymbol, TokenString, TokenNumber, TokenComment` (six counting the
  sentinel; five "real"), and `String()` (`tokenizer.go:30-45`) names all
  five. The dispatch in `tokenize` (`tokenizer.go:117-128`) emits none of
  them.
- **Drift**: the public surface advertises a complete tokenizer, the
  implementation is empty. A consumer importing the package will see all
  the constants and assume they work.

---

## Suggested implementation order

The scaffold's CLAUDE.md prescribes a test-first order; this audit's
findings line up with it:

1. Tokenizer tests (Findings 4.1, 4.6) → tokenizer code (1.1–1.7, 5.1).
2. Parser tests (Findings 4.2, 4.5, 4.6) → AST shapes + parser
   (2.1–2.9, 5.2).
3. Printer direct tests (Finding 4.3) + round-trip tests (Finding 4.4) →
   printer code (3.1–3.6).

Every finding above cites both the spec section and the source file location
so the implementer can fix in place without re-reading the whole spec.
