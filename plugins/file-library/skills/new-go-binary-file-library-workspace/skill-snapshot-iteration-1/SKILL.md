---
name: new-go-binary-file-library
description: Scaffold a new Go binary file library package with types, decoder, encoder, and tests
disable-model-invocation: true
argument-hint: "[package-name]"
---

Scaffold a new Go binary file library package at `./$ARGUMENTS[0]/` following the **types / decoder / encoder** pipeline pattern. The package name is `$ARGUMENTS[0]`. Read `references/architecture.md` for the patterns each generated file must implement and why.

## Before Scaffolding

1. Check the repo for any existing Go binary file library package (one with `types.go`, `decoder.go`, `encoder.go`). If one exists, read its source files and use them as the reference for structure, helpers, naming, and license-header style.
2. Read any `CLAUDE.md` files in the repo root or existing packages for project-specific conventions.
3. Check `git log --oneline -10` for commit message style.

If no existing binary file library exists, fall back to the canonical patterns in `references/architecture.md`.

## What to Generate

Create the following files. Adapt naming and style to match any existing binary library in the repo. Default the byte order to `binary.BigEndian` (most network and container formats); the implementer can change it.

### 1. `$ARGUMENTS[0]/doc.go`
Package doc file with the repo's license header and a one-line package comment.

### 2. `$ARGUMENTS[0]/types.go`
Scaffold with:
- A top-level `File` struct as the placeholder root type (one placeholder field is fine).
- An example enum type (e.g., `Kind uint8`) with a `const` block and a `String()` method, to demonstrate the pattern.
- An example bit field mask/shift constant pair, as a comment-documented placeholder.
- An exported `ErrInvalid` sentinel or an `InvalidFieldError` struct error type with `Error()` and `Unwrap()`, to demonstrate the error pattern.

### 3. `$ARGUMENTS[0]/types_test.go`
Scaffold with:
- One placeholder table-driven test (`TestKindString`) verifying the example enum's `String()` method.
- One placeholder size-check test using `binary.Size()` against a fixed-size struct, to demonstrate the pattern.
- `t.Parallel()` at both function and subtest level; assertions via `github.com/stretchr/testify/require`.

### 4. `$ARGUMENTS[0]/decoder.go`
Scaffold with:
- Internal `decoder` struct wrapping an `io.Reader` and storing a `binary.ByteOrder` field.
- Constructor `newDecoder(r io.Reader) *decoder` defaulting to `binary.BigEndian`.
- One stub method `func (d *decoder) readFile() (*File, error)` returning a zero `File` and an "unimplemented" error wrapped with `fmt.Errorf("decoding File: %w", err)`.
- Public `Decode(r io.Reader) (*File, error)` that constructs a decoder and calls `readFile()`.
- An `UnexpectedEOFError` or similar error type to demonstrate context-rich error wrapping.

### 5. `$ARGUMENTS[0]/decoder_test.go`
Scaffold with:
- One placeholder table-driven test (`TestDecode`) using a hex byte literal input (`[]byte{0x00}`) and `bytes.NewReader`, asserting the current "unimplemented" error path. The test exists so the implementer can flip it to a happy path once `readFile()` is real.
- `t.Parallel()` at both levels; `require` from testify.

### 6. `$ARGUMENTS[0]/encoder.go`
Scaffold with:
- Internal `encoder` struct wrapping an `io.Writer` and storing a `binary.ByteOrder` field.
- Constructor `newEncoder(w io.Writer) *encoder` defaulting to `binary.BigEndian`.
- One stub method `func (e *encoder) writeFile(f *File) error` returning an "unimplemented" error wrapped with `fmt.Errorf("encoding File: %w", err)`.
- Public `Encode(w io.Writer, f *File) error` that constructs an encoder and calls `writeFile(f)`.

### 7. `$ARGUMENTS[0]/encoder_test.go`
Scaffold with:
- One placeholder table-driven test (`TestEncode`) using a `bytes.Buffer` as the writer, asserting the current "unimplemented" error path.
- One placeholder round-trip test (`TestEncodeDecodeRoundTrip`) showing the `Encode → Decode → compare` pattern, currently expected to fail at the unimplemented step.
- `t.Parallel()` at both levels; `require` from testify.

### 8. `$ARGUMENTS[0]/CLAUDE.md`
Create a package-level CLAUDE.md documenting:
- The types / decoder / encoder pipeline and where each component lives.
- Byte order: which `binary.ByteOrder` the package uses and why.
- Error wrapping convention: `fmt.Errorf("decoding %s: %w", structName, err)` / `"encoding %s: %w"`.
- Testing style: hex byte literals, `bytes.NewReader` / `bytes.Buffer`, table-driven, `require`, `t.Parallel()` at both levels, round-trip tests for every encoder method.

Base the structure on any existing package-level `CLAUDE.md` in the repo, or write a fresh one if none exists. See `references/architecture.md` for the full rationale to summarize from.

## After Scaffolding

1. Run `go mod tidy` to update dependencies.
2. Run `go build ./$ARGUMENTS[0]/...` to verify compilation.
3. Run `go test ./$ARGUMENTS[0]/...` — placeholder tests should pass against the "unimplemented" error stubs.
4. Report what was created, the byte order chosen, and what the user should implement next (typically: define the real types from `SPEC.md`, then run the `implement-binary-file-library` agent).
