---
name: new-go-binary-file-library
description: Scaffold a new Go binary file library package with types, decoder, encoder, and tests
disable-model-invocation: true
argument-hint: "[package-name]"
---

Scaffold a new Go binary file library package at `./$ARGUMENTS[0]/` following the **types / decoder / encoder** pipeline. The package name is `$ARGUMENTS[0]`. Read `references/architecture.md` for the patterns each generated file must implement and why — especially the section on the decode/encode error chain, which the scaffold pre-wires.

## Before Scaffolding

1. Check the repo for any existing Go binary file library (a package with `types.go`, `decoder.go`, `encoder.go`). If one exists, read its source files and use them as the reference for structure, helpers, naming, and license-header style.
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

1. `go mod tidy`.
2. `go build ./$ARGUMENTS[0]/...` to verify compilation.
3. `go test ./$ARGUMENTS[0]/...` — placeholder tests should pass against the unimplemented stubs (the chain is real even though the bytes aren't).
4. Report what was created, the byte order chosen, and what the user should implement next (typically: define the real types from `SPEC.md`, then run the `implement-go-binary-file-library` agent).
