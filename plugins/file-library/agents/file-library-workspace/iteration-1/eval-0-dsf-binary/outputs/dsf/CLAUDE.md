# dsf — X-Plane DSF binary file library

## Pipeline

- `types.go` — pure data structures: top-level `File`, sub-structs, enum types, bit-field constants, error types/sentinels.
- `decoder.go` — `Decode(r io.Reader) (*File, error)` plus internal `decoder` and per-structure `readX` helpers.
- `encoder.go` — `Encode(w io.Writer, f *File) error` plus internal `encoder` and per-structure `writeX` helpers.

Every public entry point lives in either `decoder.go` or `encoder.go`. Tests are siblings (`*_test.go`).

## Byte order

DSF is little-endian on the wire (X-Plane is little-endian throughout). The scaffold defaults `decoder.byteOrder` and `encoder.byteOrder` to `binary.LittleEndian`. If the implementer encounters a sub-structure that overrides this, document it both in `SPEC.md`'s `Conventions` section and in a comment on the relevant `readX` / `writeX` method.

## Error chain

Every decode and encode error must surface as

    FieldError → OffsetError → <leaf error>

where the leaf is either a sentinel (`errors.Is(err, ErrInvalid)`), a typed error declared in `types.go`, or an `io` error like `io.ErrUnexpectedEOF`. Always go through `d.wrapErr(field, leaf)` / `e.wrapErr(field, leaf)` — never construct `FieldError` or `OffsetError` directly. The chain enables:

- `errors.Is(err, errFooBar)` for sentinels
- `errors.As(err, &fe)` (`*FieldError`) for the failing field path
- `errors.As(err, &oe)` (`*OffsetError`) for the byte offset where the read/write blew up

`Field` is a dotted path (e.g., `"Header.Length"`, `"Records[3].Type"`) so a single `FieldError` value identifies the failing field unambiguously.

## Testing style

- Hex byte literals + `bytes.NewReader` (decode) and `bytes.Buffer` (encode).
- Table-driven, `t.Parallel()` at both the function and subtest level, assertions via `github.com/stretchr/testify/require`.
- Every new `readX` decode method gets a happy-path test plus at least one failure-path test that asserts the full `FieldError → OffsetError → leaf` chain.
- Every new `writeX` encode method gets a happy-path test plus a round-trip test (`Encode → Decode → require.Equal`).
- Round-trip is the cheapest end-to-end correctness check; add one for every new structure.
