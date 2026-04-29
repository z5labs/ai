# Binary File Library Architecture

A Go binary file library is a package that reads and writes one binary format. It follows a **types / decoder / encoder** pipeline, mirroring `encoding/json` or `encoding/xml` in shape but specialized to a single wire format.

The pipeline is intentionally narrow. Each component owns one concern, and each can be tested in isolation:

```
bytes  ─── Decode ─►  *File (typed AST)  ─── Encode ─►  bytes
              │              │                  │
           decoder.go     types.go           encoder.go
```

This document explains what each component must do and why. The scaffolding skill produces stubs for each; the implementer (or the `implement-binary-file-library` agent) fills them in against a real `SPEC.md`.

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
- Wrap every non-trivial error with structural context: `fmt.Errorf("decoding Header.Length: %w", err)`. The implementer chasing a bad input file from a hex dump will thank you.
- Distinguish "the bytes ran out unexpectedly" (`io.ErrUnexpectedEOF`) from "the bytes were fine but the value was illegal" (your own typed error).

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
- Mirror the decoder: `fmt.Errorf("encoding Header.Length: %w", err)`.
- Encoders fail less often than decoders, but they still fail — `io.Writer` returns errors, and the encoder must surface them.

## 4. Tests

Tests are how you know the round-trip is honest.

### Conventions
- `t.Parallel()` at both the test function and each subtest. The action functions are pure; parallel tests catch hidden global state.
- Table-driven with `testCases` slice and `t.Run(tc.name, ...)`. Names are lowercase descriptive.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a binary test that keeps running after the first failure produces noise, not signal.
- Hex byte literals for binary inputs/outputs: `[]byte{0x00, 0x01, 0xFF}`. Comments above each block label the field they correspond to in the format.

### Three test shapes
1. **Type tests** (`types_test.go`): `binary.Size()` matches the spec, `String()` methods round-trip enum values, bit-field accessors return the right slices.
2. **Decode tests** (`decoder_test.go`): bytes in, struct out. One subtest per scenario in the spec's examples.
3. **Encode tests** (`encoder_test.go`): struct in, bytes out. Plus a round-trip subtest that runs `Encode → Decode` and `require.Equal`s the original. Round-trip is the cheapest end-to-end correctness check available; every encoder method should have one.

### When tests fail
- A round-trip mismatch is almost always either a byte-order disagreement between encoder and decoder, or a length-prefix that's read in a different unit than it's written (bytes vs records).
- A `binary.Size()` test failure means a struct field has a non-fixed type. Fix the struct, not the test.

## 5. Why this shape

A single binary format gets one package, and that package gets exactly three files of production code. The constraint matters because binary formats accrete: a real implementation of DNS or PNG ends up with dozens of types and hundreds of fields, and a sprawling file layout makes the round-trip property impossible to audit at a glance. Three files, three responsibilities, one byte order, round-trip tests on every writer — that's the contract. Everything in the scaffold exists to make that contract obvious from the moment the package is created.
