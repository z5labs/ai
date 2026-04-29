# gzip

A Go binary file library for the gzip wire format.

This package follows the **types / decoder / encoder** pipeline pattern
documented in the repo-level `references/architecture.md`. Each component owns
one concern and can be tested in isolation.

## Pipeline

```
bytes  ─── Decode ─►  *File (typed AST)  ─── Encode ─►  bytes
              │              │                  │
           decoder.go     types.go           encoder.go
```

- `types.go` — Go translation of the wire format. Top-level `File` struct,
  enums (with `String()`), bit-field mask/shift constants, `ErrInvalid`
  sentinel, and the `OffsetError` / `FieldError` wrapper types live here.
- `decoder.go` — pull-based reader. Owns the `io.Reader` (wrapped in a
  `countingReader`) and the byte order; exposes `Decode(r) (*File, error)`
  plus internal `readX` methods and the `wrapErr` helper.
- `encoder.go` — the decoder's inverse. Owns the `io.Writer` (wrapped in a
  `countingWriter`) and the byte order; exposes `Encode(w, f) error` plus
  internal `writeX` methods and the symmetric `wrapErr` helper.

## Byte order

This package currently uses `binary.BigEndian` because that is the scaffold's
default. **The real gzip format is little-endian** (RFC 1952). When you fill
in `SPEC.md`, switch the constant in **both** `newDecoder` and `newEncoder`
to `binary.LittleEndian` — they must agree, otherwise round-trip tests will
fail in confusing ways.

## The decode/encode error chain

Every error surfaced by the decoder or encoder must follow the same shape:

```
FieldError{Field: "Header.Length"}
   → OffsetError{Offset: 4}
      → <leaf error: io.ErrUnexpectedEOF, ErrInvalid, ...>
```

This gives callers three independent handles:

- `errors.Is(err, leafSentinel)` to test for stable conditions.
- `errors.As(err, &fieldErr)` to recover the dotted field path.
- `errors.As(err, &offsetErr)` to recover the byte offset where the
  read/write failed.

### Always go through wrapErr

Every error site in `decoder.go` returns its error through `d.wrapErr(field,
err)`. Every error site in `encoder.go` returns through `e.wrapErr(field,
err)`. **Do not** construct `FieldError` or `OffsetError` directly anywhere
else — the offset is read straight off `countingReader.n` /
`countingWriter.n`, and bypassing the helper is the easy way to make the
reported offset drift from reality.

### Nested fields

When a parent reader/writer delegates to a child, name the field with a
dotted path (`"Header.Length"`, `"Members[3].Trailer.CRC32"`). The convention
in this package is to let the deepest call site name the most specific field
— `errors.As` will pull out the innermost `FieldError` from the chain, which
is what tests and tools want.

### Sentinels vs typed errors

- Use `ErrInvalid` (and any future format-specific sentinels) for stable,
  comparable conditions where `errors.Is` is enough.
- Reach for a typed struct error (with its own `Error()` and `Unwrap()`) only
  when the caller benefits from extra fields beyond field path + offset.
- `errUnimplemented` is the placeholder sentinel returned by the scaffold
  stubs; remove it once `readFile` and `writeFile` are real.

## Testing style

- Hex byte literals for binary inputs and outputs:
  `[]byte{0x1F, 0x8B, 0x08}`. Comment each block with the field it
  represents in the wire format.
- `bytes.NewReader` for decoder inputs, `bytes.Buffer` for encoder outputs.
- Table-driven tests with a `testCases` slice and `t.Run(tc.name, ...)`.
- `t.Parallel()` at **both** the test function and each subtest. The
  decoder/encoder methods are pure once implemented; parallel tests catch
  hidden global state.
- Assertions via `github.com/stretchr/testify/require` (not `assert`) — a
  binary test that keeps running after the first failure produces noise, not
  signal.
- A round-trip test for **every** encoder method: `Encode → Decode →
  require.Equal` against the original struct. Round-trip is the cheapest
  end-to-end correctness check in the package.
- Every decoder failure path asserts the full chain: `require.ErrorIs` for
  the leaf sentinel **and** `require.ErrorAs` for the `FieldError` (and
  optionally `OffsetError`) shape.

## What to implement next

1. Fill in `SPEC.md` with the real gzip member layout (RFC 1952): header
   (ID1/ID2/CM/FLG/MTIME/XFL/OS), optional extra fields gated by FLG bits
   (FEXTRA, FNAME, FCOMMENT, FHCRC), DEFLATE payload, and trailer (CRC32,
   ISIZE).
2. Switch the byte order in `newDecoder` and `newEncoder` to
   `binary.LittleEndian`.
3. Replace `File`'s `Placeholder` field with the real top-level structure
   (likely a slice of members, each with a header/payload/trailer).
4. Replace the example `Kind` enum with real types (e.g. `CompressionMethod`,
   the `OS` byte) and replace the placeholder bit-field constants with the
   FLG mask/shift constants (`flagsFTEXT`, `flagsFHCRC`, `flagsFEXTRA`,
   `flagsFNAME`, `flagsFCOMMENT`).
5. Implement `readFile` / `writeFile` (and per-substructure `readHeader`,
   `readMember`, `readTrailer` / `writeHeader`, ...) following the patterns
   in `references/architecture.md`. Route every error through `wrapErr`.
6. Run the `implement-binary-file-library` agent to drive the test-first
   implementation, or fill in tests + implementation by hand.
