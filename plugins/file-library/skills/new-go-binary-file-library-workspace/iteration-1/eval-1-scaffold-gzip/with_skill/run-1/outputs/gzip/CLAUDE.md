# gzip

Go binary file library for the gzip format. Follows the **types / decoder / encoder** pipeline pattern.

## Pipeline

```
bytes  ─── Decode ─►  *File (typed AST)  ─── Encode ─►  bytes
              │              │                  │
           decoder.go     types.go           encoder.go
```

- `types.go` — Go translation of the wire format. One struct per structure, typed-integer enums with `String()` methods, mask/shift constants for bit fields, and sentinel/struct error types.
- `decoder.go` — Pull-based reader. Internal `decoder` struct wraps `io.Reader` and stores `binary.ByteOrder`. Methods named `readX` (`readFile`, future `readHeader`, etc.). Public surface is `Decode(r io.Reader) (*File, error)`.
- `encoder.go` — Inverse of the decoder. Internal `encoder` wraps `io.Writer` with the same byte order. Methods named `writeX`. Public surface is `Encode(w io.Writer, f *File) error`.

## Byte order

Default scaffold uses `binary.BigEndian` per the `new-go-binary-file-library` skill's default. The real gzip wire format is **little-endian**; switch both `newDecoder` and `newEncoder` to `binary.LittleEndian` when wiring up real reads/writes. The decoder and encoder must always agree.

## Error wrapping convention

Every non-trivial decode/encode error is wrapped with structural context so a hex-dump debugger can locate the offending field:

```go
fmt.Errorf("decoding %s: %w", structName, err)
fmt.Errorf("encoding %s: %w", structName, err)
```

Examples:
- `decoding File: unimplemented`
- `decoding Header.Length: unexpected EOF`
- `encoding Trailer.CRC32: short write`

Use `UnexpectedEOFError` (or a typed variant) when the bytes ran out, and `ErrInvalid` / `InvalidFieldError` when the bytes were fine but the value was illegal.

## Testing style

- Hex byte literals for inputs and outputs: `[]byte{0x1F, 0x8B, 0x08, 0x00, ...}` with field-labeling comments above each block.
- `bytes.NewReader` for decode inputs, `bytes.Buffer` for encode outputs.
- Table-driven tests: a `testCases` slice plus `t.Run(tc.name, ...)`.
- `t.Parallel()` at **both** the test function and each subtest. The action functions are pure; parallelism catches hidden global state.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a binary test that keeps running after the first failure produces noise.
- **Round-trip tests for every encoder method**: `Encode → Decode → require.Equal` against the original. Round-trip is the cheapest end-to-end correctness check available.
- A `binary.Size()` test per fixed-size struct documents the wire-size invariant.

## Next steps for the implementer

1. Define the real `File`, header, optional fields, and trailer types in `types.go` from `SPEC.md`.
2. Switch byte order to `binary.LittleEndian` in `newDecoder` and `newEncoder`.
3. Replace the unimplemented stubs in `decoder.go` and `encoder.go` with real reads/writes, wrapping each field with structural context.
4. Replace the placeholder tests with real spec-example inputs/outputs and turn the round-trip test green.
