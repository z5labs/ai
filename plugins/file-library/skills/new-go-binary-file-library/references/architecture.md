# Binary File Library Architecture

A Go binary file library is a package that reads and writes one binary format. It follows a **types / decoder / encoder** pipeline, mirroring `encoding/json` or `encoding/xml` in shape but specialized to a single wire format.

The pipeline is intentionally narrow. Each component owns one concern, and each can be tested in isolation:

```
bytes  ─── Decode ─►  *File (typed AST)  ─── Encode ─►  bytes
              │              │                  │
           decoder.go     types.go           encoder.go
```

This document explains what each component must do and why. The scaffolding skill produces stubs for each; the implementer (or the `implement-go-binary-file-library` agent) fills them in against a real `SPEC.md`.

## 1. Types (`types.go`)

The types layer is a faithful Go translation of the wire format.

### Structs
- One Go struct per structure in the format. Field order matches wire order so a reader can mentally line them up.
- For fixed-size structures, fields use sized integer types (`uint8`, `uint16`, `uint32`, `uint64`) and fixed-size byte arrays (`[N]byte`) so `binary.Size()` and `binary.Read()` work directly.
- For variable-length structures, a length-prefixed slice (`[]byte`, `[]Record`) is the norm. The decoder reads the length, then reads that many elements.

### Enums
- Define as a typed integer (`type Opcode uint8`) with a `const` block of named values.
- Provide a `String()` method. This pays for itself the first time you read a test failure or a hex dump.

### Bit fields
- Store the underlying integer in its natural Go type. Don't try to model bit-packed fields as separate struct fields — the wire layout has to round-trip cleanly.
- Expose mask and shift constants (e.g., `flagsQRMask = 0x80`, `flagsOpcodeShift = 3`) and, when it improves clarity, accessor methods on the parent struct.

### Errors
- Sentinel errors (`ErrInvalidLength`) for stable, comparable conditions.
- Struct error types (`type UnexpectedValueError struct { Field string; Got uint32 }`) when context matters. Always implement `Error()` and, if wrapping, `Unwrap()`.
- Two universal wrapper types — `OffsetError{Offset int64; Err error}` and `FieldError{Field string; Err error}` — sit between every leaf error and the caller. See "The decode/encode error chain" below.

## The decode/encode error chain

Decode failures need to answer three questions: **what failed**, **where in the byte stream**, and **what underlying error caused it**. The scaffold pre-wires a uniform answer:

```
FieldError{Field: "Header.Length"}
   → OffsetError{Offset: 4}
      → io.ErrUnexpectedEOF   (or any sentinel/leaf error)
```

Two wrapper types do all the work:

```go
type OffsetError struct { Offset int64; Err error }
type FieldError  struct { Field  string; Err error }
```

Both implement `Error()` and `Unwrap()`. That gives the caller three handles:

- `errors.Is(err, io.ErrUnexpectedEOF)` finds the leaf sentinel.
- `errors.As(err, &fe)` extracts the field path (`fe.Field`).
- `errors.As(err, &oe)` extracts the byte offset where the read failed (`oe.Offset`).

### Why two layers, not one combined `DecodeError`

A single combined struct would force every caller to know about the wrapper type. Two narrow types let `errors.As` walk the chain and pull out exactly the dimension the caller cares about — useful in tests (assert offset OR field, not both at once), and useful in tools that print human-friendly diagnostics ("decoding Header.Length at byte 4: unexpected EOF").

### Tracking offset

The decoder wraps its `io.Reader` in an unexported `countingReader` that tallies bytes on every `Read`. The encoder mirrors this with a `countingWriter`. The current offset is always `d.r.n` / `e.w.n`. No manual accounting in the read methods; the wrapper does it for you.

### One helper, every error site

The decoder gets a small helper:

```go
func (d *decoder) wrapErr(field string, err error) error {
    if err == nil { return nil }
    return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}
```

Every `readX` method funnels its errors through `d.wrapErr("Header.Length", err)`. The encoder has a symmetric `e.wrapErr`. This keeps the chain shape uniform — never construct `FieldError` or `OffsetError` directly outside the helper, or the offset will drift.

### Nested fields

When a parent reads a child structure, prepend the parent's name to the child's `FieldError.Field`. Either re-wrap (`return d.wrapErr("Header." + childErr.Field, errors.Unwrap(childErr))`) or, more commonly, let the parent's call site name the field as `"Header"` and let the child's `wrapErr` name the leaf — `errors.As` will still pull out the most specific `FieldError` at the bottom of the chain. Pick one convention and document it in the package `CLAUDE.md`.

## 2. Decoder (`decoder.go`)

The decoder is a pull-based reader. It owns the `io.Reader` and the byte order, and exposes one public function plus internal per-structure methods.

### Public surface
```go
func Decode(r io.Reader) (*File, error)
```
Keep it minimal. If the format requires options (a version, a strictness flag), add a separate `DecodeWithOptions` rather than overloading `Decode`.

### Internal `decoder` struct
- Wraps the `io.Reader`.
- Stores `binary.ByteOrder` as a field so individual reads stay terse (`d.byteOrder.Uint32(buf)`).
- Methods are named `readX` where `X` is the struct being read: `readHeader`, `readRecord`, `readSection`.

### Reading patterns
- **Fixed-size fields**: `binary.Read(d.r, d.byteOrder, &x)` does the right thing for any type that's a fixed-size sequence of fixed-size fields.
- **Variable-length fields**: read the length prefix first, allocate, then `io.ReadFull` (don't use `r.Read` — it can short-read).
- **Bit fields**: read the underlying integer in one call, then mask and shift in Go. Never try to read individual bits.

### Errors
- Funnel every error through `d.wrapErr("FieldName", err)` so the chain is uniform: `FieldError → OffsetError → leaf`. See "The decode/encode error chain" above for why and how.
- Distinguish "the bytes ran out unexpectedly" (`io.ErrUnexpectedEOF`) from "the bytes were fine but the value was illegal" (your own typed error). Both wrap cleanly through `wrapErr`; the caller uses `errors.Is` to disambiguate.

## 3. Encoder (`encoder.go`)

The encoder is the decoder's inverse. Same shape, opposite direction.

### Public surface
```go
func Encode(w io.Writer, f *File) error
```

### Internal `encoder` struct
- Wraps the `io.Writer`.
- Stores `binary.ByteOrder` (same as the decoder — they must match).
- Methods are named `writeX`: `writeHeader`, `writeRecord`.

### Writing patterns
- **Fixed-size fields**: `binary.Write(e.w, e.byteOrder, x)`.
- **Variable-length fields**: write the length prefix, then write the payload in one `e.w.Write` call.
- **Bit fields**: pack with shifts and OR, then write the resulting integer.
- **Padding and alignment**: write zero bytes explicitly. Never assume the writer pads.

### Errors
- Mirror the decoder: route every error through `e.wrapErr("FieldName", err)`. The chain shape is identical, but the offset comes from `e.w.n` (the counting writer) instead of `d.r.n`.
- Encoders fail less often than decoders, but they still fail — `io.Writer` returns errors, and the encoder must surface them with the same field-and-offset context the decoder provides.

## 4. Tests

Tests are how you know the round-trip is honest.

### Conventions
- `t.Parallel()` at both the test function and each subtest. The action functions are pure; parallel tests catch hidden global state.
- Table-driven with `testCases` slice and `t.Run(tc.name, ...)`. Names are lowercase descriptive.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a binary test that keeps running after the first failure produces noise, not signal.
- Hex byte literals for binary inputs/outputs: `[]byte{0x00, 0x01, 0xFF}`. Comments above each block label the field they correspond to in the format.

### Three test shapes
1. **Type tests** (`types_test.go`): `binary.Size()` matches the spec, `String()` methods round-trip enum values, bit-field accessors return the right slices, and the error-chain shape (`FieldError → OffsetError → leaf`) survives `errors.Is` and `errors.As`.
2. **Decode tests** (`decoder_test.go`): bytes in, struct out. One subtest per scenario in the spec's examples. Failure-path subtests use `require.ErrorIs` for the leaf sentinel and `require.ErrorAs` to assert the field path and byte offset.
3. **Encode tests** (`encoder_test.go`): struct in, bytes out. Plus a round-trip subtest that runs `Encode → Decode` and `require.Equal`s the original. Round-trip is the cheapest end-to-end correctness check available; every encoder method should have one.

### When tests fail
- A round-trip mismatch is almost always either a byte-order disagreement between encoder and decoder, or a length-prefix that's read in a different unit than it's written (bytes vs records).
- A `binary.Size()` test failure means a struct field has a non-fixed type. Fix the struct, not the test.

## 5. Why this shape

A single binary format gets one package, and that package gets exactly three files of production code. The constraint matters because binary formats accrete: a real implementation of DNS or PNG ends up with dozens of types and hundreds of fields, and a sprawling file layout makes the round-trip property impossible to audit at a glance. Three files, three responsibilities, one byte order, round-trip tests on every writer — that's the contract. Everything in the scaffold exists to make that contract obvious from the moment the package is created.
