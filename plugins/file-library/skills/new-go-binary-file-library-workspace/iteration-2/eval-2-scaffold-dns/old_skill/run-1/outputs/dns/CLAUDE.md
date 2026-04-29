# dns

Go package for reading and writing DNS wire-format messages.

## Pipeline

This package follows the **types / decoder / encoder** pipeline:

```
bytes  ─── Decode ─►  *File  ─── Encode ─►  bytes
              │          │          │
          decoder.go  types.go   encoder.go
```

- `types.go` — Go translation of the DNS wire format. One struct per wire
  structure, typed-integer enums with `String()` methods, mask/shift
  constants for bit-fields, sentinel and struct error types.
- `decoder.go` — pull-based reader. Internal `decoder` struct wraps an
  `io.Reader` and a `binary.ByteOrder`. Public surface is `Decode(r) (*File, error)`;
  internal methods follow the `readX` naming convention.
- `encoder.go` — push-based writer. Internal `encoder` struct wraps an
  `io.Writer` and a `binary.ByteOrder`. Public surface is `Encode(w, f) error`;
  internal methods follow the `writeX` naming convention.

## Byte Order

`binary.BigEndian`. DNS wire format is network byte order (RFC 1035), which is
big-endian. The encoder and decoder must agree on byte order — round-trip
tests catch any drift.

## Error Wrapping

Every non-trivial error is wrapped with structural context naming the field
being read or written:

- Decoder: `fmt.Errorf("decoding %s: %w", structOrField, err)` — e.g.,
  `"decoding Header.Length"`.
- Encoder: `fmt.Errorf("encoding %s: %w", structOrField, err)`.

`UnexpectedEOFError` distinguishes "the bytes ran out" from "the bytes were
fine but the value was illegal" (`InvalidFieldError`, which wraps
`ErrInvalid`).

## Testing Style

- **Hex byte literals** for all binary inputs/outputs: `[]byte{0x00, 0x01, 0xFF}`.
  Comment each block with the field it corresponds to.
- **`bytes.NewReader`** for decoder inputs, **`bytes.Buffer`** for encoder
  outputs.
- **Table-driven** tests with a `testCases` slice and `t.Run(tc.name, ...)`.
  Names are lowercase descriptive.
- **`t.Parallel()` at both levels** — the test function and every subtest.
- **`github.com/stretchr/testify/require`** (not `assert`) so the first
  failure halts the subtest.
- **Round-trip tests for every encoder method**: `Encode → Decode →
  require.Equal` against the original. This is the cheapest end-to-end
  correctness check; do not skip it.
- **`binary.Size()` checks** for fixed-size structs to catch accidental
  variable-length fields.

## Implementing the Spec

The current contents of this package are scaffolding stubs. To fill them in:

1. Define real types in `types.go` from the DNS RFCs (RFC 1035 and friends):
   `Header`, `Question`, `ResourceRecord`, `Opcode`, `RCode`, `QType`,
   `QClass`, etc.
2. Run the `implement-binary-file-library` agent, or implement the decoder
   and encoder methods by hand following the patterns above.
3. Replace each placeholder test with real spec-driven cases, and flip the
   `wantErr: errUnimplemented` assertions to happy-path `require.Equal`s.
