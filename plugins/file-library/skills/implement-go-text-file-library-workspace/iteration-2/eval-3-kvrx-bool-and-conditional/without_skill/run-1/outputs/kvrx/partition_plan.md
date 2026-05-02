# KVRX bool + conditional implementation partition plan

This task has parser/printer surface that exceeds the 600-line single-phase
budget for KVRX. The parser must handle:

- record-bool body (`record bool KEY = true|false`)
- conditional statement (`if (Expr) { stmt; }` + `elif`/`else`)
- a tiny expression evaluator (`&NAME` lookup + `==` against bool literal)

To keep each phase below the 600-line gate the parser and printer phases are
each split into two sub-units. The implementation runs sequentially in this
agent (no subagents), but the chunking still stands as the editing plan.

## Phase 1 — Tokenizer

Single sub-unit. Token types added: `TokenNewline`. Recognise:

- whitespace (skip)
- newline (yield `TokenNewline`)
- line comment `# ...` (yield `TokenComment`)
- identifier (`[A-Za-z_][A-Za-z0-9_]*`) — including the keywords `record`,
  `block`, `import`, `type`, `if`, `elif`, `else`, `true`, `false`, `null`
- symbol — single-rune `=`, `{`, `}`, `(`, `)`, `;`, `&` and the two-rune
  `==` greedy form
- string `"..."` with no escape processing for this slice (test inputs do not
  exercise escapes)

Tokens carry exact `Pos{Line, Column}` (1-based) of their first rune.

## Phase 2 — Parser

### Sub-unit 2a — AST + dispatch + record-bool

- AST nodes: `BoolLiteral`, `Record`, `Conditional`, `ConditionalBranch`,
  `Reference`, `BinaryExpr` (just `==`).
- Dispatch on first identifier value (`record` / `if`).
- `record bool KEY = true|false` end-to-end.
- Scope tracking: the parser keeps a flat `[]*Record` slice of previously
  declared records so the conditional resolver can look them up.

### Sub-unit 2b — Conditional + tiny expression resolver

- `if "(" Expr ")" "{" body "}" { elif (...) { ... } } [ else { ... } ]`.
- Body grammar: zero or more `Statement ";"` (here only record statements
  are accepted, matching the SPEC's restricted body grammar for this slice).
- Expression resolver: enough to evaluate either a bare `BoolLiteral`, a
  `&NAME` reference (resolved by walking the parser's record list), or a
  comparison `Reference == BoolLiteral` / `BoolLiteral == BoolLiteral`.
  Anything else in the expression is parsed but evaluation falls back to
  rejecting the conditional with `NonStaticConditionalError` per spec.
- Per spec, the AST stores all branches verbatim (so they round-trip);
  evaluation only decides which branch is "active" — a flag the AST does
  not need to expose for this slice.

## Phase 3 — Printer

### Sub-unit 3a — record-bool

- `printFile` walks `f.Statements` and dispatches on concrete type.
- `Record` with `Bool` value prints `record bool KEY = true|false\n`.

### Sub-unit 3b — conditional

- Conditional prints `if (Expr) {\n    Stmt;\n    ...}` then each elif/else
  branch on the same line as the closing `}` of the prior branch, matching
  the structural form in the spec example.
- Inner statements print with one tab of indent and a trailing `;`.

## Tests

- `tokenizer_test.go` — exact `Pos` checks for keywords + symbols + bool
  literals + an `if (... == ...) { ... }` body.
- `parser_test.go` — drives `Parse()` only:
  - empty file
  - single `record bool ENABLED = true`
  - if/elif/else with `&MODE == true` body
- `printer_test.go` — direct prints + round-trip for both shapes.
