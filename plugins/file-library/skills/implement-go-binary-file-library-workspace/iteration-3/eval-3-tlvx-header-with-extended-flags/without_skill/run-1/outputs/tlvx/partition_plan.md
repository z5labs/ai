# TLVX Header Implementation — Partition Plan

The TLVX decoder-phase slices clear the 600-line gate, so this work is
partitioned into sub-units before any subagent (or in-process) execution.

## Scope of this iteration

This iteration is limited to the **Header** structure end-to-end:
Magic / Version / Flags / ChecksumAlg / Reserved1 / IndexCount / ExtCount /
TrailerOffset, plus the seven defined Header.Flags bits and the
`ChecksumAlg` enum from the SPEC's "Checksum algorithms" table.

Out of scope: Index, Record, Extension Table, Trailer (filed for later
iterations).

## Sub-units

### A. Types (`types.go`)
- `Header` struct with all eight fields.
- `Flags` (uint8) named type plus seven exported flag constants:
  `FlagCompressed`, `FlagEncrypted`, `FlagSigned`, `FlagIndexed`,
  `FlagExtended`, `FlagStrict`, `FlagSealed` (bits 0..6). Bit 7 reserved.
- `ChecksumAlg` (uint8) named type plus five exported constants:
  `ChecksumCRC32IEEE` (0x01), `ChecksumCRC64ECMA` (0x02),
  `ChecksumSHA256T32` (0x03), `ChecksumXXH64` (0x04),
  `ChecksumBLAKE3T32` (0x05).
- Update `File` to embed `Header Header`.
- Constants: `Magic = "TLVX"`, `Version = 1`, `HeaderSize = 16`.
- Sentinel: `ErrMagicMismatch` leaf error so the failure-path test can
  assert on `FieldError → OffsetError → ErrMagicMismatch`.

### B. Decoder (`decoder.go`)
- `readHeader()` — reads 16 bytes via small per-field reads so the
  `OffsetError.Offset` reflects where the failing field starts, wrapping
  every leaf error through `wrapErr` to produce
  `FieldError{Field: "Header.<Name>", Err: OffsetError{...}}`.
- `readFile()` — calls `readHeader`, fills `File.Header`, returns `*File`.

### C. Encoder (`encoder.go`)
- `writeHeader(h Header)` — writes the 16 header bytes big-endian, also
  funneling errors through `wrapErr`.
- `writeFile(f *File)` — calls `writeHeader(f.Header)`.

### D. Tests
- `decoder_test.go`: hex-literal decode tests covering the minimal
  header, all-flags-set header, and an XXH64 checksum tag, plus a
  magic-mismatch failure-path test asserting the
  `FieldError → OffsetError → ErrMagicMismatch` chain.
- `encoder_test.go`: encode tests with hex-literal expectations
  mirroring the decode tests.
- `types_test.go`: keeps the existing error-chain assertion plus a
  round-trip test that encodes a populated `Header`, decodes the bytes,
  and asserts equality.

## Order of execution

A → B → C → D. All sub-units land within this single working package
directory; no cross-package edits.
