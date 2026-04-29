# dns

Go binary file library for DNS wire-format messages.

## Pipeline

This package follows the **types / decoder / encoder** pipeline:

```
bytes  ── Decode ─►  *File  ── Encode ─►  bytes
              │         │           │
          decoder.go  types.go   encoder.go
```

- `types.go` — Go translation of the DNS wire format. One struct per spec
  structure, fields in wire order, sized integers and `[N]byte` arrays so
  `binary.Size` / `binary.Read` work directly. Enums use `type X uintN` plus
  a `const` block and a `String()` method. Bit fields keep their underlying
  integer type and expose mask/shift constants (and accessor methods when
  helpful).
- `decoder.go` — internal `decoder` struct wrapping an `io.Reader` and the
  byte order, with `readX` methods (`readFile`, `readHeader`, ...) and the
  public `Decode(r io.Reader) (*File, error)` entry point.
- `encoder.go` — internal `encoder` struct wrapping an `io.Writer` and the
  byte order, with `writeX` methods and the public
  `Encode(w io.Writer, f *File) error` entry point.

## Byte order

DNS messages are framed in **network byte order** (big-endian) per
RFC 1035 section 2.3.2. Both the decoder and encoder default to
`binary.BigEndian`. They must always agree — a byte-order mismatch is the
single most common cause of round-trip failures in this package.

## Error wrapping convention

Every non-trivial decode/encode error is wrapped with structural context:

```go
return nil, fmt.Errorf("decoding Header.Length: %w", err)
return fmt.Errorf("encoding Header.Length: %w", err)
```

Use `fmt.Errorf("decoding %s: %w", structName, err)` and
`fmt.Errorf("encoding %s: %w", structName, err)` when wrapping at the
struct boundary. Distinguish "the bytes ran out" (wrap
`io.ErrUnexpectedEOF`, e.g. via `UnexpectedEOFError`) from "the bytes were
fine but the value was illegal" (return `ErrInvalid` /
`InvalidFieldError`).

## Testing

- Hex byte literals for binary inputs/outputs: `[]byte{0x00, 0x01, 0xFF}`.
  Comment each block with the field name from the spec.
- `bytes.NewReader` for decoder tests, `bytes.Buffer` for encoder tests.
- Table-driven tests with `testCases` slices and `t.Run(tc.name, ...)`.
- `t.Parallel()` at both the test function level and inside each subtest.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a
  binary-format test that keeps running after the first mismatch produces
  noise, not signal.
- Every encoder method gets a round-trip subtest: `Encode -> Decode ->
  require.Equal(original, decoded)`. Round-trip is the cheapest end-to-end
  correctness check available.

## What to implement next

The current scaffolding contains placeholder types (`File`, `Kind`,
`InvalidFieldError`) and stub `readFile` / `writeFile` methods that return
an "unimplemented" error. To turn the scaffold into a real DNS library:

1. Define the real types in `types.go` from the DNS wire-format spec
   (RFC 1035 plus relevant extensions). Start with the 12-byte header,
   then question / resource record sections.
2. Run the `implement-binary-file-library` agent (or follow the same
   test-first workflow manually) to flesh out `decoder.go` and
   `encoder.go`, replacing the stubs and updating the placeholder tests.
