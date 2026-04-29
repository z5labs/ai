# Binary File Library Architecture

A Go binary file library reads and writes one wire format. Three files, three responsibilities:

```
bytes  ─── Decode ─►  *File (typed AST)  ─── Encode ─►  bytes
              │              │                  │
           decoder.go     types.go           encoder.go
```

The orchestrator hands each phase subagent only the spec slices and source files it needs. This document is the architectural spine those subagents work from.

## 1. Types (`types.go`)

The types layer is a faithful Go translation of the wire format. Field order matches wire order so a reader of the struct can mentally line them up against the byte layout.

### Structs
- One Go struct per structure in the format.
- Fixed-size structures use sized integer types (`uint8`, `uint16`, `uint32`, `uint64`) and fixed-size byte arrays (`[N]byte`) so `binary.Size()` and `binary.Read` work directly. A `binary.Size()` test guards this — break it and the decoder slows to a crawl as `binary.Read` falls back to reflection.
- Variable-length structures hold a length-prefixed slice (`[]byte`, `[]Record`). The decoder reads the length first, then reads that many elements.

### Enums
- Define as a typed integer (`type Opcode uint8`) with a `const` block of named values.
- Provide a `String()` method. This pays off the first time you read a hex-dump test failure.

### Bit fields
- Store the underlying integer in its natural Go type. Don't model bit-packed fields as separate struct fields — the wire layout has to round-trip cleanly.
- Expose mask and shift constants (e.g., `flagsQRMask = 0x80`, `flagsOpcodeShift = 3`) and, when it improves readability, accessor methods on the parent struct.

### Errors

The scaffold pre-wires a uniform error chain:

```
FieldError{Field: "Header.Length"}
   → OffsetError{Offset: 4}
      → io.ErrUnexpectedEOF   (or any sentinel/leaf error)
```

```go
type OffsetError struct { Offset int64; Err error }
type FieldError  struct { Field  string; Err error }
```

Both implement `Error()` and `Unwrap()`. That gives the caller three handles:
- `errors.Is(err, io.ErrUnexpectedEOF)` finds the leaf sentinel.
- `errors.As(err, &fe)` extracts the field path.
- `errors.As(err, &oe)` extracts the byte offset where the read/write failed.

The two-layer split exists so callers can pull out exactly the dimension they care about — useful in tests (assert offset OR field, not both at once) and in tools that print human-friendly diagnostics.

Add new sentinels (`ErrInvalidLength`, `ErrChecksumMismatch`) and struct error types (`UnexpectedValueError{Field, Got}`) to `types.go` as needed. Always implement `Error()`, and `Unwrap()` if wrapping.

## 2. Decoder (`decoder.go`)

A pull-based reader. It owns the `io.Reader` and the byte order, and exposes one public function plus internal per-structure methods.

### Public surface
```go
func Decode(r io.Reader) (*File, error)
```
Keep it minimal. Add a separate `DecodeWithOptions` rather than overloading `Decode`.

### Internal `decoder` struct
- Wraps the `io.Reader` in an unexported `countingReader` that tallies bytes on every `Read`. Current offset is always `d.r.n` — no manual accounting in the read methods.
- Stores `binary.ByteOrder` as a field (`d.byteOrder.Uint32(buf)`).
- Methods are named `readX` where `X` is the struct being read: `readHeader`, `readRecord`.

### Reading patterns
- **Fixed-size fields**: `binary.Read(d.r, d.byteOrder, &x)` for any value that's a fixed-size sequence of fixed-size fields.
- **Variable-length fields**: read the length prefix, allocate, then `io.ReadFull` (don't use `r.Read` — it can short-read).
- **Bit fields**: read the underlying integer in one call, then mask and shift in Go. Never read individual bits.

### Errors — one helper, every site
```go
func (d *decoder) wrapErr(field string, err error) error {
    return &FieldError{Field: field, Err: &OffsetError{Offset: d.r.n, Err: err}}
}
```
Every `readX` method funnels its errors through `d.wrapErr("Header.Length", err)` — always inside the `if err != nil` branch, so the helper never sees a nil error and doesn't need to guard for one. Never construct `FieldError` or `OffsetError` directly outside the helper, or the offset will drift.

For nested fields, either re-wrap with the parent's name (`return d.wrapErr("Header." + childErr.Field, errors.Unwrap(childErr))`) or let the parent's call site name the field as `"Header"` and let the child's `wrapErr` name the leaf. Pick one convention per package and document it in the package `CLAUDE.md`.

## 3. Encoder (`encoder.go`)

The decoder's inverse. Same shape, opposite direction.

### Public surface
```go
func Encode(w io.Writer, f *File) error
```

### Internal `encoder` struct
- Wraps the `io.Writer` in an unexported `countingWriter`.
- Stores `binary.ByteOrder` (must match the decoder).
- Methods are named `writeX`: `writeHeader`, `writeRecord`.

### Writing patterns
- **Fixed-size fields**: `binary.Write(e.w, e.byteOrder, x)`.
- **Variable-length fields**: write the length prefix, then write the payload in one `e.w.Write` call.
- **Bit fields**: pack with shifts and OR, then write the resulting integer.
- **Padding and alignment**: write zero bytes explicitly. Never assume the writer pads.

### Errors
Mirror the decoder: every error funnels through `e.wrapErr("FieldName", err)`. The chain shape is identical; the offset comes from `e.w.n`.

## When tests fail

- A round-trip mismatch is almost always either a byte-order disagreement between encoder and decoder, or a length-prefix that's read in a different unit than it's written (bytes vs records).
- A `binary.Size()` test failure means a struct field has a non-fixed type. Fix the struct, not the test.
- A failure-path test that gets the wrong offset usually means an error site bypassed `wrapErr` and constructed a `FieldError` directly.

## Why this shape

A single binary format gets one package, and that package gets exactly three files of production code. Real implementations of DNS or PNG accrete dozens of types and hundreds of fields; a sprawling layout makes the round-trip property impossible to audit at a glance. Three files, three responsibilities, one byte order, round-trip tests on every writer — that's the contract every phase subagent maintains.
