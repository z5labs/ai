# gzip

Go binary file library for the gzip format. Follows the **types / decoder / encoder** pipeline shared by every binary library in this repo.

## Pipeline

```
bytes  ── Decode ─►  *File  ── Encode ─►  bytes
            │           │          │
        decoder.go   types.go   encoder.go
```

- `types.go` — Go translation of the wire format: structs (one per wire structure, fields in wire order), typed-integer enums with `String()`, mask/shift constants for bit fields, and error types (`ErrInvalid`, `InvalidFieldError`).
- `decoder.go` — `Decode(r io.Reader) (*File, error)` plus the internal `decoder` struct with `readX` methods. Owns the `io.Reader` and the `binary.ByteOrder`.
- `encoder.go` — `Encode(w io.Writer, f *File) error` plus the internal `encoder` struct with `writeX` methods. Mirrors the decoder.

## Byte order

The scaffold defaults to `binary.BigEndian`. **The real gzip format is little-endian on the wire** — switch both `newDecoder` and `newEncoder` to `binary.LittleEndian` when filling in the spec. The decoder and encoder must always agree.

## Error wrapping

Every non-trivial error gets structural context:

- Decoder: `fmt.Errorf("decoding %s: %w", structOrField, err)` — e.g. `"decoding Header.Length"`.
- Encoder: `fmt.Errorf("encoding %s: %w", structOrField, err)`.

Use sentinel errors (`ErrInvalid`) for stable comparisons and struct error types (`InvalidFieldError`, `UnexpectedEOFError`) when the caller needs the field name or the offending value.

## Testing style

- Hex byte literals for binary input/output: `[]byte{0x1f, 0x8b, 0x08}`. Comment each block with the field name from the spec.
- `bytes.NewReader` for decoder inputs, `bytes.Buffer` for encoder outputs.
- Table-driven tests with a `testCases` slice and `t.Run(tc.name, …)`. Lowercase test names.
- Assertions via `github.com/stretchr/testify/require` — never `assert`.
- `t.Parallel()` at both the test function and every subtest.
- **Every encoder method gets a round-trip test** (`Encode → Decode → require.Equal`). Round-trip is the cheapest end-to-end correctness check available.

## Status

This is a scaffold. The `readFile` and `writeFile` methods return `errUnimplemented` and the placeholder tests assert that error path. Replace the `File` placeholder with the real top-level type from `SPEC.md`, then run the `implement-binary-file-library` agent (or implement by hand) to fill in the `readX` / `writeX` methods.
