# Testing Conventions for Binary File Libraries

Tests are how you know the round-trip is honest. These conventions apply to every phase (types, decoder, encoder) — write the tests first, watch them fail for the right reason, then implement.

## Universal rules

- `t.Parallel()` at both the test function and each subtest. Action functions are pure; parallel runs catch hidden global state.
- Table-driven with a `testCases` slice and `t.Run(tc.name, ...)`. Names are lowercase, descriptive, e.g. `"header_with_optional_extra_field"`.
- Assertions via `github.com/stretchr/testify/require` — never `assert`. A binary test that keeps running after the first failure produces noise, not signal.
- Hex byte literals for binary I/O: `[]byte{0x1f, 0x8b, 0x08, 0x00}`. Add a comment above each block labelling the field it corresponds to in the spec.
- Run `go test -race ./...` after each step. The race flag is cheap and catches concurrent reuse of the encoder/decoder structs.

## Phase 1 — type tests (`types_test.go`)

Three things to verify:

### Fixed-size struct sizing
```go
func TestHeaderSize(t *testing.T) {
    t.Parallel()
    require.Equal(t, 12, binary.Size(Header{}))
}
```
A failure here means a field has a non-fixed type. Fix the struct, not the test.

### Enum String() round-trip
```go
testCases := []struct{ name string; v Opcode; want string }{
    {"query",  OpcodeQuery,  "QUERY"},
    {"iquery", OpcodeIQuery, "IQUERY"},
}
```
Validates the `const` block matches the `String()` switch.

### Error chain shape
```go
err := &FieldError{Field: "Header", Err: &OffsetError{Offset: 4, Err: errUnimplemented}}
require.ErrorIs(t, err, errUnimplemented)
var fe *FieldError
require.ErrorAs(t, err, &fe)
require.Equal(t, "Header", fe.Field)
var oe *OffsetError
require.ErrorAs(t, err, &oe)
require.Equal(t, int64(4), oe.Offset)
```
Locks in `FieldError → OffsetError → leaf`. Every decoder/encoder failure path test relies on this.

## Phase 2 — decoder tests (`decoder_test.go`)

### Happy path: bytes in, struct out
```go
input := []byte{
    // ID1, ID2, CM, FLG (gzip member header)
    0x1f, 0x8b, 0x08, 0x00,
    // MTIME (4 bytes, little-endian)
    0x00, 0x00, 0x00, 0x00,
    // XFL, OS
    0x00, 0xff,
}
f, err := Decode(bytes.NewReader(input))
require.NoError(t, err)
require.Equal(t, uint8(0x08), f.Header.CM)
```

One subtest per scenario in the spec's `## Examples` section. The hex literal is the source of truth — copy from the spec, don't paraphrase.

### Failure path: assert the chain
```go
input := []byte{0x1f, 0x8b}    // truncated
_, err := Decode(bytes.NewReader(input))
require.ErrorIs(t, err, io.ErrUnexpectedEOF)

var fe *FieldError
require.ErrorAs(t, err, &fe)
require.Equal(t, "Header.CM", fe.Field)

var oe *OffsetError
require.ErrorAs(t, err, &oe)
require.Equal(t, int64(2), oe.Offset)   // bytes consumed before the failure
```
Every new `readX` method gets at least one truncation test. The offset assertion catches `wrapErr` regressions.

## Phase 3 — encoder tests (`encoder_test.go`)

### Direct: struct in, bytes out
```go
var buf bytes.Buffer
err := Encode(&buf, &File{Header: Header{ID1: 0x1f, ID2: 0x8b, CM: 0x08}})
require.NoError(t, err)
require.Equal(t, []byte{0x1f, 0x8b, 0x08, /* ... */}, buf.Bytes())
```

### Round-trip: every encoder method gets one
```go
original := &File{Header: Header{...}, Records: []Record{...}}

var buf bytes.Buffer
require.NoError(t, Encode(&buf, original))

decoded, err := Decode(&buf)
require.NoError(t, err)
require.Equal(t, original, decoded)
```

Round-trip is the cheapest end-to-end correctness check available, and it's what keeps the byte-order and length-prefix invariants honest. A round-trip mismatch is almost always one of:
- byte-order disagreement between encoder and decoder
- length-prefix unit mismatch (bytes vs records)
- a struct field that wasn't actually written (or read past)

If round-trip passes, the encoder and decoder agree about the wire format — even if both happen to be wrong about the spec. Pair round-trip with at least one direct hex test per structure, so the spec itself anchors the bytes.

## Bit-field tests

Test the mask/shift constants directly when the field is non-trivial. For example, gzip's FLG byte:

```go
flg := byte(0)
flg |= flgFTEXT
flg |= flgFNAME
require.True(t, flg&flgFTEXT != 0)
require.True(t, flg&flgFHCRC == 0)
```

For DNS-style bit-packed headers, decode the natural integer, assert the unpacked sub-fields match expected values, then re-pack and assert byte equality.

## Variable-length sizing

Variable-length records are the most common source of round-trip bugs. Add at least:
1. An empty record (length prefix = 0, no payload).
2. A single-byte record (smallest non-empty).
3. A record at or near the maximum length the prefix can express (catches truncation in the prefix).

## Fixture style

Inline byte literals beat external fixture files for binary tests — the bytes live next to the assertion that depends on them. Reach for a `testdata/` file only when the input is too large to read inline (multi-KB) or when you want to exercise a real-world specimen captured from another implementation.
