---
name: new-go-text-file-library
description: Scaffold a new Go text file library package with tokenizer, parser, printer, and tests
disable-model-invocation: true
argument-hint: "[package-name]"
---

Scaffold a new file library package at `./$ARGUMENTS[0]/` following the tokenizer/parser/printer pipeline pattern. The package name is `$ARGUMENTS[0]`.

## Before Scaffolding

1. Check the repo for any existing file library packages that follow the tokenizer/parser/printer pattern. If one exists, read its source files and use them as the reference for structure, helpers, and naming conventions.
2. Read any `CLAUDE.md` files in the repo root or existing packages for project-specific conventions (license headers, import style, error patterns, etc.).
3. Check `git log --oneline -10` for commit message conventions.

If no existing file library package exists, use the canonical patterns described below.

## What to Generate

Create the following files, adapting naming and style to match any existing file library in the repo.

### 1. `$ARGUMENTS[0]/doc.go`

Package doc file with appropriate license header (match repo convention) and a package comment.

### 2. `$ARGUMENTS[0]/tokenizer.go`

Scaffold with:
- `Pos` struct (Line, Column int)
- `Token` struct (Pos, Type, Value) with `String()` method
- `TokenType` enum with at minimum: `TokenComment`, `TokenIdentifier`, `TokenSymbol`, `TokenString`, `TokenNumber`
- `tokenizer` struct wrapping a `*bufio.Reader` with position tracking and `next()`, `backup()` methods
- `tokenizerAction` type: `func(t *tokenizer, yield func(Token, error) bool) tokenizerAction`
- Helper functions for error propagation, token yielding, and whitespace skipping
- `Tokenize` public function returning `iter.Seq2[Token, error]`
- Stub entry point action that reads one rune and returns nil
- Error types (e.g., `UnexpectedCharacterError`)

### 3. `$ARGUMENTS[0]/tokenizer_test.go`

Scaffold with:
- `collect` helper function for gathering tokens from `iter.Seq2`
- One example table-driven test (`TestTokenizer`) with a single placeholder test case
- `t.Parallel()` at both levels
- `require` from testify

### 4. `$ARGUMENTS[0]/parser.go`

Scaffold with:
- `File` struct as the top-level AST node (with at minimum a placeholder field)
- `Type` interface with marker method
- `parser` struct wrapping `next func() (Token, error, bool)` with `expect()` method
- `parserAction[T]` type: `func(p *parser, t T) (parserAction[T], error)`
- `Parse` public function: creates parser via `iter.Pull2(Tokenize(r))`, runs action loop, returns `*File`
- Stub entry action that returns `(nil, nil)`
- Error types (e.g., `UnexpectedEndOfTokensError`, `UnexpectedTokenError`)

### 5. `$ARGUMENTS[0]/parser_test.go`

Scaffold with:
- One example table-driven test (`TestParser`) with a single placeholder test case
- Tests MUST call `Parse()` with real source strings, not construct AST manually
- `t.Parallel()` at both levels
- `require` from testify

### 6. `$ARGUMENTS[0]/printer.go`

Scaffold with:
- `printer` struct wrapping `io.Writer` with `err` field, `write()`, `writef()` methods
- `printerAction` type: `func(pr *printer, f *File) printerAction`
- Helper for writing a string then continuing to next action
- `Print` public function: runs action loop checking `pr.err` each iteration
- Stub entry action that returns nil

### 7. `$ARGUMENTS[0]/printer_test.go`

Scaffold with:
- One example table-driven test (`TestPrinter`) with direct print test structure
- One example round-trip test (`TestPrinterRoundTrip`) structure
- `t.Parallel()` at both levels
- `require` from testify

### 8. `$ARGUMENTS[0]/CLAUDE.md`

Create a package-specific CLAUDE.md documenting:
- The state machine pattern with the package's action types
- Helper function signatures and usage
- Testing style and conventions
- Error types

Base the structure on any existing package-level `CLAUDE.md` in the repo, or create a fresh one if none exists.

## After Scaffolding

1. Run `go mod tidy` to update dependencies
2. Run `go build ./$ARGUMENTS[0]/...` to verify compilation
3. Run `go test ./$ARGUMENTS[0]/...` to verify tests pass
4. Report what was created and what the user should implement next
