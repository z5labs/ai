---
name: new-go-text-file-library
description: Scaffold a new Go text file library package with tokenizer, parser, printer, and tests. Use when the user asks to "start a new <format> library" or "scaffold a new <format> parser". Skip when the user wants to add features to an existing text package (use `implement-go-text-file-library` instead) or when the target format is binary, e.g. `types.go`/`decoder.go`/`encoder.go` (use `new-go-binary-file-library` instead).
disable-model-invocation: true
argument-hint: "[package-name]"
---

Scaffold a new Go text file library package at `./$ARGUMENTS[0]/` following the **tokenizer / parser / printer** pipeline. The package name is `$ARGUMENTS[0]`. Read `references/architecture.md` for the patterns each generated file must implement and why — especially the action-loop state machine that the tokenizer, parser, and printer all share.

## Inputs

- **`$ARGUMENTS[0]`** (required) — the package name, supplied as the slash-command argument (e.g. `/new-go-text-file-library kvr`). Used both as the directory name (`./$ARGUMENTS[0]/`) and the Go `package` identifier. Validate by listing `./$ARGUMENTS[0]/`; if the directory already exists with any of the files in `## Outputs` present, stop and direct the user to `implement-go-text-file-library` so prior work is not clobbered. An invalid Go identifier (hyphens, leading digit, etc.) will surface as a `go build` failure in `## After Scaffolding`.

## Outputs

- **Generated files** in `./$ARGUMENTS[0]/`, written via `Write`: `doc.go`, `tokenizer.go`, `tokenizer_test.go`, `parser.go`, `parser_test.go`, `printer.go`, `printer_test.go`, `CLAUDE.md`. Each is a Go source file (or, for `CLAUDE.md`, package-level guidance markdown) — see `## What to Generate` for per-file content. `Write` will overwrite an existing file at the same path, which is why the input-validation step above stops if any of those output paths already exist in the target directory.
- **Side effects** (run from `./$ARGUMENTS[0]/` after files are written; this repo has no root `go.mod`, so each new package's tests must be run from inside the package):
  - `(cd ./$ARGUMENTS[0] && go mod tidy)` — refreshes module dependencies.
  - `(cd ./$ARGUMENTS[0] && go build ./...)` — verifies compilation.
  - `(cd ./$ARGUMENTS[0] && go test -race ./...)` — placeholder tests must pass against the empty-input stubs before reporting success.

## Before Scaffolding

1. Check the repo for any existing Go text file library (a package with `tokenizer.go`, `parser.go`, `printer.go`). If one exists, read its source files and use them as the reference for structure, helpers, naming, and license-header style.
2. Read any `CLAUDE.md` files in the repo root or existing packages for project-specific conventions.
3. Check `git log --oneline -10` for commit message style.

If no existing text file library exists, fall back to the canonical patterns in `references/architecture.md`.

## What to Generate

Create the following files. Adapt naming and style to match any existing text library in the repo. The stubs must compile and the placeholder tests must pass — they exercise the pipeline shape (action loop, error propagation, iter.Seq2 streaming) even though no real tokens or AST nodes exist yet.

### 1. `$ARGUMENTS[0]/doc.go`
Package doc file with the repo's license header and a one-line package comment.

### 2. `$ARGUMENTS[0]/tokenizer.go`
Scaffold with:
- `Pos` struct `{Line, Column int}`.
- `TokenType` typed integer with at minimum: `TokenComment`, `TokenIdentifier`, `TokenSymbol`, `TokenString`, `TokenNumber`. Include a `String()` method (named values pay for themselves the first time a test fails).
- `Token` struct `{Pos Pos; Type TokenType; Value string}` with `String()` method.
- Unexported `tokenizer` struct wrapping a `*bufio.Reader` with `pos Pos`, `next() (rune, error)`, and `backup()` methods.
- `tokenizerAction` type: `func(t *tokenizer, yield func(Token, error) bool) tokenizerAction`. Returning `nil` ends iteration.
- Helpers for the common patterns: yield-then-continue, yield-error-and-stop, skip-whitespace.
- Public `Tokenize(r io.Reader) iter.Seq2[Token, error]` that constructs the tokenizer and runs the action loop.
- Stub entry-point action that reads one rune via `t.next()`, returns `nil` on `io.EOF`, otherwise returns `nil` (the implementer wires up the dispatch switch).
- Exported error type `UnexpectedCharacterError{Pos Pos; Char rune}` with `Error()`.

### 3. `$ARGUMENTS[0]/tokenizer_test.go`
Scaffold with:
- A `collect` helper that drains an `iter.Seq2[Token, error]` into `([]Token, error)`.
- One placeholder table-driven test (`TestTokenizer`) with a single empty-input case asserting zero tokens and no error — proves the iterator and action loop work end-to-end.
- `t.Parallel()` at both function and subtest level; assertions via `github.com/stretchr/testify/require`.

### 4. `$ARGUMENTS[0]/parser.go`
Scaffold with:
- `File` struct as the top-level AST node (one placeholder field is fine).
- `Type` interface with an unexported marker method (e.g., `isType()`) so concrete AST nodes can satisfy it.
- Unexported `parser` struct wrapping `next func() (Token, error, bool)` (the result of `iter.Pull2`) with an `expect(types ...TokenType) (Token, error)` method.
- Generic `parserAction[T any]` type: `func(p *parser, t T) (parserAction[T], error)`. Returning `(nil, nil)` completes successfully; `(nil, err)` terminates with error.
- Public `Parse(r io.Reader) (*File, error)` that calls `iter.Pull2(Tokenize(r))`, runs the top-level action loop against a `*File`, and returns it.
- Stub entry action that returns `(nil, nil)` so the empty-input test passes.
- Exported error types `UnexpectedEndOfTokensError` and `UnexpectedTokenError{Got Token; Want []TokenType}` with `Error()`.

### 5. `$ARGUMENTS[0]/parser_test.go`
Scaffold with:
- One placeholder table-driven test (`TestParser`) with a single empty-input case calling `Parse(strings.NewReader(""))` and asserting equality against `&File{}` (the zero-value `*File`) and no error.
- Tests must call `Parse()` with real source strings to produce non-trivial expected values — never hand-construct AST nodes for those expectations. The zero-value `&File{}` used in the empty-input scaffold case is the only exception. Document this rule in the package `CLAUDE.md` so the implementer doesn't drift.
- `t.Parallel()` at both levels; `require` from testify.

### 6. `$ARGUMENTS[0]/printer.go`
Scaffold with:
- Unexported `printer` struct wrapping `io.Writer` with an `err error` field, plus `write(s string)` and `writef(format string, args ...any)` helpers that short-circuit when `pr.err != nil`.
- `printerAction` type: `func(pr *printer, f *File) printerAction`. Returning `nil` ends.
- Helper for the write-then-continue pattern (e.g., `writeThen(s string, next printerAction) printerAction`).
- Public `Print(w io.Writer, f *File) error` that runs the action loop, checking `pr.err` each iteration so a write error stops the loop and surfaces.
- Stub entry action that returns `nil` (empty input prints nothing).

### 7. `$ARGUMENTS[0]/printer_test.go`
Scaffold with:
- One placeholder table-driven test (`TestPrinter`) with an empty-`File` case asserting empty output and no error.
- One placeholder round-trip skeleton (`TestPrinterRoundTrip`) showing the `Parse → Print → Parse → require.Equal` shape, even if the body is a single empty-string case. Round-trip is the cheapest end-to-end correctness check available; every printer method should have one once the implementer fills things in.
- `t.Parallel()` at both levels; `require` from testify.

### 8. `$ARGUMENTS[0]/CLAUDE.md`
Write a **self-contained** package-level guide. Inline the relevant patterns directly; do not point readers at the skill's `references/architecture.md` — that file is the skill's own scratchpad and does not exist in the user's repo, so any link to it will dangle. Cover:
- The tokenizer / parser / printer pipeline and where each component lives.
- The action-loop state machine pattern with the package's three action types and what `nil` means for each.
- Helper signatures and when to use them (yield-then-continue, write-then-continue, expect).
- The "for complex/nested types, use the inner action loop pattern — no inline for-loops" rule. This is the single rule most likely to be violated by a fast implementer; call it out.
- Testing style: parser tests must drive the public `Parse()`, printer tests must include round-trips, `t.Parallel()` at both levels, `require` from testify, table-driven.

If a package-level `CLAUDE.md` already exists elsewhere in the repo, mirror its structure and tone; otherwise write fresh.

## After Scaffolding

1. `(cd ./$ARGUMENTS[0] && go mod tidy)`.
2. `(cd ./$ARGUMENTS[0] && go build ./...)` to verify compilation.
3. `(cd ./$ARGUMENTS[0] && go test -race ./...)` — placeholder tests should pass against the empty-input stubs.
4. Report what was created and what the user should implement next (typically: extract a `SPEC.md` with `extract-text-spec`, then run the `implement-go-text-file-library` agent).
