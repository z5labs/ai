---
name: new-go-binary-file-library
description: Scaffold a new Go binary file library package with types, decoder, encoder, and tests. Use when the user asks to "scaffold a new binary decoder/encoder package" or "start a new binary format library". Skip when the user wants to add features to an existing binary package (use `implement-go-binary-file-library` instead) or when the target format is text, e.g. `tokenizer.go`/`parser.go`/`printer.go` (use `new-go-text-file-library` instead).
argument-hint: "[package-name]"
---

Scaffold a new Go binary file library package at `./$ARGUMENTS[0]/` following the **types / decoder / encoder** pipeline. The package name is `$ARGUMENTS[0]`. Read `references/architecture.md` for the patterns each generated file must implement and why — especially the section on the decode/encode error chain, which the scaffold pre-wires.

## Inputs

- **`$ARGUMENTS[0]`** (required) — the package name, supplied as the slash-command argument (e.g. `/new-go-binary-file-library gzip`). Used both as the directory name (`./$ARGUMENTS[0]/`) and the Go `package` identifier, so it must satisfy both constraints at once. Run all three checks below before writing any files; on any failure, stop with a message that names the failing check and the offending value. Failing fast here avoids half-written packages and prevents `Write` from silently overwriting existing source.
  - **Path-safe** — must not contain `/`, `\`, or `.`, and must not equal `..`. This is a single directory name, not a path.
  - **Valid Go package identifier** — must match `^[a-z][a-z0-9]*$` and must not be a Go keyword (`break`, `case`, `chan`, `const`, `continue`, `default`, `defer`, `else`, `fallthrough`, `for`, `func`, `go`, `goto`, `if`, `import`, `interface`, `map`, `package`, `range`, `return`, `select`, `struct`, `switch`, `type`, `var`). Lowercase-only matches Go's "short, concise, lowercase, no underscores" package-name convention; trailing digits are fine (`md5`, `oauth2`). Common rejects: hyphens (`my-format`), leading digits (`2html`), camelCase (`myFormat`).
  - **No filesystem collision** — if any entry exists at `./$ARGUMENTS[0]` (directory, regular file, or symlink), refuse before writing anything. For an existing directory, direct the user to `implement-go-binary-file-library` if they meant to add features to an in-progress package, or to remove it manually if it's stale and they want to re-scaffold. For a file or symlink, direct the user to remove the entry before re-running — there's no in-progress work to preserve. The skill never overwrites or merges because `Write` clobbers individual files silently — one check up front is the safe contract.

## Outputs

- **Generated files** in `./$ARGUMENTS[0]/`, written via `Write`: `doc.go`, `types.go`, `types_test.go`, `decoder.go`, `decoder_test.go`, `encoder.go`, `encoder_test.go`, `CLAUDE.md`. Each is a Go source file (or, for `CLAUDE.md`, package-level guidance markdown) — see `## What to Generate` for per-file content. `Write` would overwrite an existing file at the same path silently, which is why the input-validation step above refuses outright if any entry exists at `./$ARGUMENTS[0]`.
- **Side effects** (run from `./$ARGUMENTS[0]/` after files are written; this repo has no root `go.mod`, so each new package's tests must be run from inside the package):
  - `(cd ./$ARGUMENTS[0] && go mod tidy)` — refreshes module dependencies.
  - `(cd ./$ARGUMENTS[0] && go build ./...)` — verifies compilation.
  - `(cd ./$ARGUMENTS[0] && go test -race ./...)` — placeholder tests must pass against the unimplemented stubs (the `FieldError → OffsetError → errUnimplemented` chain is real even though the bytes aren't) before reporting success.

## Before Scaffolding

1. Find existing Go binary file libraries in the repo — directories that contain all three of `types.go`, `decoder.go`, and `encoder.go` (use `Glob`/`Grep`; package layouts are shallow). If multiple candidates exist, pick one deterministically using this tiebreaker chain — evaluate each rule against the current candidate set in order; if a rule matches one or more current candidates, replace the set with only those matches, and if a rule matches zero current candidates, leave the set unchanged. Stop when one candidate remains:
   1. **Sibling first** — prefer candidates whose parent directory equals the parent of `./$ARGUMENTS[0]/` (i.e. sibling packages). If one or more siblings exist in the current set, keep only those siblings; otherwise keep the set unchanged. Sibling packages almost always share license headers, helper conventions, and `CLAUDE.md` style, so the reference is most informative.
   2. **Most recent commit** — among the current candidates, compute `git log -1 --format=%ct -- <path>` for each package directory. If `git log` returns an empty result for a path or errors (for example, the path is untracked or `.git/` metadata is unavailable), treat that path's timestamp as `0`. Keep only the candidates with the highest timestamp. Recency is a proxy for "current style"; older packages may predate convention changes.
   3. **Lexicographic** — if multiple candidates still remain after the earlier rules (e.g. a bulk-import commit, or every candidate fell back to `0`), sort the surviving paths and take the first. This is the deterministic last-resort tiebreaker so two runs never disagree.

   Read the chosen package's source files and use them as the reference for structure, helpers, naming, and license-header style. If no candidate exists, fall back to the canonical patterns in `references/architecture.md`.
2. Read any `CLAUDE.md` files in the repo root or existing packages for project-specific conventions.
3. Check `git log --oneline -10` for commit message style.

## What to Generate

Create the following files. Adapt naming and style to match any existing binary library in the repo. Default the byte order to `binary.BigEndian` (most network and container formats); the implementer can change it.

### 1. `$ARGUMENTS[0]/doc.go`
Package doc file with the repo's license header and a one-line package comment.

### 2. `$ARGUMENTS[0]/types.go`
Scaffold with:
- A top-level `File` struct as the placeholder root type (one placeholder field is fine).
- An example enum type (e.g., `Kind uint8`) with a `const` block and a `String()` method, to demonstrate the pattern.
- An example bit field mask/shift constant pair (commented as a placeholder; OK to be unused until the implementer adds real bit fields).
- An exported `ErrInvalid` sentinel.
- An unexported `errUnimplemented = errors.New("unimplemented")` sentinel — the decoder/encoder stubs return this so tests can assert the error chain via `errors.Is`.
- An `OffsetError` struct: `{Offset int64; Err error}` with `Error()` (formats as `"at byte N: <err>"`) and `Unwrap() error`.
- A `FieldError` struct: `{Field string; Err error}` with `Error()` (formats as `"decoding <Field>: <err>"`) and `Unwrap() error`. Field is a dotted path, e.g. `"Header.Length"`.

### 3. `$ARGUMENTS[0]/types_test.go`
Scaffold with:
- One placeholder table-driven test (`TestKindString`) for the example enum.
- One placeholder size-check test using `binary.Size()` against a fixed-size struct.
- A `TestErrorChain` test that constructs `&FieldError{Field: "Header", Err: &OffsetError{Offset: 4, Err: errUnimplemented}}` and asserts `errors.Is(err, errUnimplemented)`, `errors.As(err, &(*FieldError)(nil))`, and `errors.As(err, &(*OffsetError)(nil))` — this nails down the chain shape so the implementer doesn't accidentally break `errors.Is`/`errors.As` later.
- `t.Parallel()` at both function and subtest level; assertions via `github.com/stretchr/testify/require`.

### 4. `$ARGUMENTS[0]/decoder.go`
Scaffold with:
- An unexported `countingReader` wrapping `io.Reader` with field `n int64`; `Read` delegates and increments `n` by the number of bytes read.
- Internal `decoder` struct holding `*countingReader` and a `binary.ByteOrder` field.
- Constructor `newDecoder(r io.Reader) *decoder` that wraps `r` in a `countingReader` and defaults to `binary.BigEndian`.
- A helper method `func (d *decoder) wrapErr(field string, err error) error` that returns `&FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}` — used at every error site so the chain is uniform.
- One stub method `func (d *decoder) readFile() (*File, error)` returning `nil, d.wrapErr("File", errUnimplemented)`.
- Public `Decode(r io.Reader) (*File, error)` that constructs a decoder via `newDecoder` and calls `readFile()`.

### 5. `$ARGUMENTS[0]/decoder_test.go`
Scaffold with:
- One placeholder table-driven test (`TestDecode`) using `bytes.NewReader([]byte{0x00})` and asserting the full chain via `require.ErrorIs(t, err, errUnimplemented)` plus `require.ErrorAs(t, err, &fieldErr)` and `require.Equal(t, "File", fieldErr.Field)`.
- `t.Parallel()` at both levels; `require` from testify.

### 6. `$ARGUMENTS[0]/encoder.go`
Scaffold with:
- An unexported `countingWriter` wrapping `io.Writer` with field `n int64`; `Write` delegates and increments `n`.
- Internal `encoder` struct holding `*countingWriter` and a `binary.ByteOrder` field.
- Constructor `newEncoder(w io.Writer) *encoder` defaulting to `binary.BigEndian`.
- Symmetric helper `func (e *encoder) wrapErr(field string, err error) error` returning `&FieldError{Field: field, Err: &OffsetError{Offset: e.w.n, Err: err}}`.
- One stub method `func (e *encoder) writeFile(f *File) error` returning `e.wrapErr("File", errUnimplemented)`.
- Public `Encode(w io.Writer, f *File) error` that constructs an encoder and calls `writeFile(f)`.

### 7. `$ARGUMENTS[0]/encoder_test.go`
Scaffold with:
- One placeholder table-driven test (`TestEncode`) using `bytes.Buffer`, asserting the same chain shape as `TestDecode`.
- One placeholder round-trip test (`TestEncodeDecodeRoundTrip`) showing the `Encode → Decode → require.Equal` pattern, expected to fail at the unimplemented step until both sides are real.
- `t.Parallel()` at both levels; `require` from testify.

### 8. `$ARGUMENTS[0]/CLAUDE.md`
Document for future contributors:
- The types / decoder / encoder pipeline and where each component lives.
- Byte order: which `binary.ByteOrder` the package uses and why.
- The decode/encode error chain: every error must surface as `FieldError → OffsetError → <source error>`. Use `d.wrapErr` / `e.wrapErr` at every error site; don't construct `FieldError`/`OffsetError` directly. The chain enables `errors.Is(err, errFooBar)` for sentinels, `errors.As(err, &fe)` for the failing field path, and `errors.As(err, &oe)` for the byte offset where the read/write blew up.
- Testing style: hex byte literals, `bytes.NewReader` / `bytes.Buffer`, table-driven, `require`, `t.Parallel()` at both levels, round-trip tests for every encoder method, error-chain assertions on every decoder failure path.

Base the structure on any existing package-level `CLAUDE.md` in the repo, or write fresh from `references/architecture.md`.

## After Scaffolding

1. `(cd ./$ARGUMENTS[0] && go mod tidy)`.
2. `(cd ./$ARGUMENTS[0] && go build ./...)` to verify compilation.
3. `(cd ./$ARGUMENTS[0] && go test -race ./...)` — placeholder tests should pass against the unimplemented stubs (the chain is real even though the bytes aren't).
4. Report what was created, the byte order chosen, and what the user should implement next (typically: define the real types from `SPEC.md`, then run the `implement-go-binary-file-library` agent).
