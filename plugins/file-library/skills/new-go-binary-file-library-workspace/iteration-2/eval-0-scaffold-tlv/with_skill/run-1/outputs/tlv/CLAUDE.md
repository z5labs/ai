# tlv

A Go binary file library for the TLV (type-length-value) wire format.

This package follows the **types / decoder / encoder** pipeline pattern
documented in the repo-level `references/architecture.md`. Each component
owns one concern and can be tested in isolation.

## Pipeline

```
bytes  ─── Decode ─►  *File (typed AST)  ─── Encode ─►  bytes
              │              │                  │
           decoder.go     types.go           encoder.go
```

- `types.go` — Go translation of the wire format. Structs, enums (with
  `String()`), bit-field mask/shift constants, and the error wrapper types
  (`OffsetError`, `FieldError`, `ErrInvalid`) live here.
- `decoder.go` — pull-based reader. Owns the `io.Reader` (wrapped in a
  `countingReader`) and the byte order; exposes `Decode(r) (*File, error)`
  plus internal `readX` methods.
- `encoder.go` — the decoder's inverse. Owns the `io.Writer` (wrapped in a
  `countingWriter`) and the byte order; exposes `Encode(w, f) error` plus
  internal `writeX` methods.

## Byte order

This package uses `binary.BigEndian`. Most network and container formats are
big-endian, and that's the scaffold's default. If the real TLV spec turns
out to be little-endian, change the constant in **both** `newDecoder` and
`newEncoder` — they must agree, otherwise round-trip tests will fail in
confusing ways.

## The decode/encode error chain

Every error surfaced by this package has the same shape:

```
FieldError{Field: "Header.Length"}
   → OffsetError{Offset: 4}
      → <leaf sentinel or wrapped error>
```

Two helpers do all the wrapping. Call them at every error site; never
construct `FieldError` or `OffsetError` directly outside the helper, or the
offset will drift away from the counting reader/writer's actual position.

```go
// decoder.go
func (d *decoder) wrapErr(field string, err error) error { ... }
// encoder.go
func (e *encoder) wrapErr(field string, err error) error { ... }
```

That uniform chain gives callers three independent handles:

- `errors.Is(err, ErrInvalid)` (or any leaf sentinel) finds the underlying
  cause.
- `errors.As(err, &fe)` extracts the field path: `fe.Field` is the dotted
  name (e.g. `"Header.Length"`).
- `errors.As(err, &oe)` extracts the byte offset: `oe.Offset` is the
  position in the input/output stream where the failure was detected.

### Nested fields

When a parent reads a child structure, prefer letting the child's `wrapErr`
name the leaf field — `errors.As` will pull out the most specific
`FieldError` at the bottom of the chain. Only re-wrap with a parent prefix
(`"Header." + childErr.Field`) if the leaf name would be ambiguous on its
own. Pick one convention per package and stick to it.

### Why two layers, not one combined `DecodeError`

A combined struct would force every caller to know about the wrapper. Two
narrow types let `errors.As` walk the chain and pull out exactly the
dimension the caller cares about — useful in tests (assert offset OR field,
not both at once), and useful in tools that print human-friendly
diagnostics ("decoding Header.Length at byte 4: unexpected EOF").

## Testing style

- Hex byte literals for binary inputs and outputs:
  `[]byte{0x00, 0x01, 0xFF}`. Comment each block with the field it
  represents.
- `bytes.NewReader` for decoder inputs, `bytes.Buffer` for encoder outputs.
- Table-driven tests with a `testCases` slice and `t.Run(tc.name, ...)`.
- `t.Parallel()` at **both** the test function and each subtest. The
  decoder/encoder methods are pure; parallel tests catch hidden global
  state.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a
  binary test that keeps running after the first failure produces noise,
  not signal.
- A round-trip test for **every** encoder method: `Encode → Decode →
  require.Equal` against the original struct. This is the cheapest
  end-to-end correctness check available.
- Every decoder failure-path test should assert the chain with
  `require.ErrorIs` for the leaf sentinel **and** `require.ErrorAs` for
  `*FieldError` / `*OffsetError`.

## What to implement next

1. Fill in `SPEC.md` for the real TLV layout you're targeting.
2. Replace `File`'s placeholder field with the real top-level structure.
3. Replace the example `Kind` enum and bit-field constants with the real
   wire-format values.
4. Implement `readFile` and `writeFile` (and any per-substructure `readX` /
   `writeX` helpers) following the patterns in
   `references/architecture.md`. Funnel every error through `d.wrapErr` /
   `e.wrapErr`.
5. Once `readFile` / `writeFile` are real, flip the placeholder
   `require.ErrorIs(..., errUnimplemented)` assertions in the tests to
   `require.NoError` (and re-enable the `require.Equal` in the round-trip
   test).
6. Run the `implement-binary-file-library` agent to drive the test-first
   implementation, or fill in tests + implementation by hand.
