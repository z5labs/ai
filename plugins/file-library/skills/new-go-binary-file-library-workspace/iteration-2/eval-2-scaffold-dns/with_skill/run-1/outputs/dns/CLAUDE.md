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
  `binary.Size` / `binary.Read` work directly. Enums use `type X uintN`
  plus a `const` block and a `String()` method (see `Kind` for the
  pattern). Bit fields keep their underlying integer type and expose
  mask/shift constants (and accessor methods when helpful).
- `decoder.go` — internal `decoder` struct wrapping an `io.Reader` (via
  `countingReader`) and the byte order, with `readX` methods (`readFile`,
  `readHeader`, ...) and the public `Decode(r io.Reader) (*File, error)`
  entry point.
- `encoder.go` — internal `encoder` struct wrapping an `io.Writer` (via
  `countingWriter`) and the byte order, with `writeX` methods and the
  public `Encode(w io.Writer, f *File) error` entry point.

## Byte order

DNS messages are framed in **network byte order** (big-endian) per
RFC 1035 section 2.3.2. Both the decoder and encoder default to
`binary.BigEndian`. They must always agree — a byte-order mismatch is the
single most common cause of round-trip failures in this package.

## The decode/encode error chain

Every decode and encode error must surface as a uniform three-layer chain:

```
FieldError{Field: "Header.Length"}
   → OffsetError{Offset: 4}
      → io.ErrUnexpectedEOF   (or any sentinel/leaf error)
```

That gives callers three independent handles:

- `errors.Is(err, ErrInvalid)` (or any leaf sentinel) finds the underlying
  cause.
- `errors.As(err, &fe)` with `var fe *FieldError` extracts the dotted
  field path, e.g. `"Header.Length"`.
- `errors.As(err, &oe)` with `var oe *OffsetError` extracts the byte
  offset where the read or write failed.

### Always go through `wrapErr`

The decoder and encoder each have a `wrapErr(field string, err error)
error` helper. **Every error site must funnel through it.** Constructing
`FieldError` or `OffsetError` directly will desync the offset (the
counting reader/writer is the only source of truth) and force callers to
re-check the chain shape. Don't.

```go
// good
return d.wrapErr("Header.Length", err)

// bad
return &FieldError{Field: "Header.Length", Err: &OffsetError{Offset: 4, Err: err}}
```

### Tracking offset

The decoder wraps its `io.Reader` in an unexported `countingReader` that
tallies bytes on every `Read`; the encoder mirrors this with
`countingWriter`. The current offset is always `d.r.n` / `e.w.n`. No
manual accounting in `readX` / `writeX`.

### Nested fields

When a parent's `readX` calls a child's `readX`, the child's `wrapErr`
already names the leaf field. Let `errors.As` walk to the most specific
`FieldError` rather than re-wrapping at every level. If you need a dotted
path, prepend the parent name once at the top of the parent reader and
document the convention here.

## Sentinels

- `ErrInvalid` — exported. Returned when the bytes were read fine but the
  value violates the DNS wire format (illegal enum, length prefix that
  exceeds the message, etc.). Compare with `errors.Is`.
- `errUnimplemented` — unexported. Returned by the stub `readFile` /
  `writeFile` so tests can pin down the chain shape before the real
  implementation lands. Remove or replace as you fill in real logic.

## Testing style

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
- Every decoder failure-path subtest asserts the full chain:
  `require.ErrorIs(t, err, leafSentinel)` plus `require.ErrorAs(t, err,
  &fieldErr)` (and the expected `fieldErr.Field`) plus `require.ErrorAs(t,
  err, &offsetErr)` (and, when meaningful, the expected
  `offsetErr.Offset`).

## What to implement next

The current scaffolding contains placeholder types (`File`, `Kind`,
`ErrInvalid`, `errUnimplemented`) and stub `readFile` / `writeFile`
methods that return the unimplemented sentinel through the canonical
error chain. To turn the scaffold into a real DNS library:

1. Define the real types in `types.go` from the DNS wire-format spec
   (RFC 1035 plus relevant extensions). Start with the 12-byte header,
   then question / resource record sections.
2. Run the `implement-binary-file-library` agent (or follow the same
   test-first workflow manually) to flesh out `decoder.go` and
   `encoder.go`, replacing the stubs and updating the placeholder tests.
