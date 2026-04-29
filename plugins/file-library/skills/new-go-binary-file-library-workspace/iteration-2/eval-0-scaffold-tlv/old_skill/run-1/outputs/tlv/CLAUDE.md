# tlv

A Go binary file library for the TLV (type-length-value) wire format.

This package follows the **types / decoder / encoder** pipeline pattern
documented in the repo-level `references/architecture.md`. Each component owns
one concern and can be tested in isolation.

## Pipeline

```
bytes  ─── Decode ─►  *File (typed AST)  ─── Encode ─►  bytes
              │              │                  │
           decoder.go     types.go           encoder.go
```

- `types.go` — Go translation of the wire format. Structs, enums (with
  `String()`), bit-field mask/shift constants, and error types live here.
- `decoder.go` — pull-based reader. Owns the `io.Reader` and the byte order;
  exposes `Decode(r) (*File, error)` plus internal `readX` methods.
- `encoder.go` — the decoder's inverse. Owns the `io.Writer` and the byte
  order; exposes `Encode(w, f) error` plus internal `writeX` methods.

## Byte order

This package uses `binary.BigEndian`. Most network and container formats are
big-endian, and that's the scaffold's default. If the real TLV spec turns out
to be little-endian, change the constant in **both** `newDecoder` and
`newEncoder` — they must agree, otherwise round-trip tests will fail in
confusing ways.

## Error wrapping convention

Every non-trivial error wraps with structural context so a hex-dump debugger
can find the offending field:

- Decoder: `fmt.Errorf("decoding %s: %w", structName, err)`
- Encoder: `fmt.Errorf("encoding %s: %w", structName, err)`

Sentinel errors (`ErrInvalid`) are used for stable, comparable conditions.
Struct error types (`InvalidFieldError`, `UnexpectedEOFError`) are used when
field-level context matters; both implement `Error()` and `Unwrap()`.

Distinguish "the bytes ran out unexpectedly" (`io.ErrUnexpectedEOF` /
`UnexpectedEOFError`) from "the bytes were fine but the value was illegal"
(`ErrInvalid` / `InvalidFieldError`).

## Testing style

- Hex byte literals for binary inputs and outputs:
  `[]byte{0x00, 0x01, 0xFF}`. Comment each block with the field it represents.
- `bytes.NewReader` for decoder inputs, `bytes.Buffer` for encoder outputs.
- Table-driven tests with a `testCases` slice and `t.Run(tc.name, ...)`.
- `t.Parallel()` at **both** the test function and each subtest. The
  decoder/encoder methods are pure; parallel tests catch hidden global state.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a
  binary test that keeps running after the first failure produces noise, not
  signal.
- A round-trip test for **every** encoder method: `Encode → Decode →
  require.Equal` against the original struct. This is the cheapest end-to-end
  correctness check available.

## What to implement next

1. Fill in `SPEC.md` for the real TLV layout you're targeting.
2. Replace `File`'s placeholder field with the real top-level structure
   (typically a slice of records with type, length, and value fields).
3. Replace the example `Kind` enum and bit-field constants with the real
   wire-format values.
4. Implement `readFile` and `writeFile` (and any per-substructure `readX` /
   `writeX` helpers) following the patterns in `references/architecture.md`.
5. Run the `implement-binary-file-library` agent to drive the test-first
   implementation, or fill in tests + implementation by hand.
