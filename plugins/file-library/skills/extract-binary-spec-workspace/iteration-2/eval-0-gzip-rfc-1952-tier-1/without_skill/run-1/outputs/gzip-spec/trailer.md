# Trailer (RFC 1952 Section 2.3.1)

Every member ends with a fixed 8-byte trailer that immediately follows
the compressed payload:

      0   1   2   3   4   5   6   7
    +---+---+---+---+---+---+---+---+
    |     CRC32     |     ISIZE     |
    +---+---+---+---+---+---+---+---+

| Offset | Size | Field | Type   | Notes |
|-------:|-----:|-------|--------|-------|
|      0 |    4 | CRC32 | u32 LE | CRC-32 of the *uncompressed* data. |
|      4 |    4 | ISIZE | u32 LE | Uncompressed input size, modulo 2^32. |

## CRC32

- Algorithm: CRC-32 as defined in ISO 3309 / ITU-T V.42 section
  8.1.1.6.2.
- This is the same CRC used in PNG and zlib; it is computed over the
  *uncompressed* input bytes, not over the compressed bytes.
- Polynomial (reflected form): `0xEDB88320`.
- Initial register: `0xFFFFFFFF` (equivalent to "init to 0 with pre/post
  one's-complement conditioning").
- Final register XOR: `0xFFFFFFFF`.
- The RFC's appendix sample code (Section 8) implements pre- and
  post-conditioning inside `update_crc`, so callers initialise their
  running CRC value to `0` and pass it through `update_crc` to start the
  computation.
- Stored little-endian in the trailer.
- Go's standard library exposes this exact algorithm as
  `hash/crc32.IEEE` / `hash/crc32.IEEETable`.

## ISIZE

- The uncompressed size of the input, **modulo 2^32**.
- Stored as a little-endian unsigned 32-bit integer.
- For inputs larger than 4 GiB the value wraps; decoders should not
  reject such files but must understand that ISIZE alone cannot
  authenticate length for >= 2^32-byte streams.

## Decoder verification

A decoder that has just finished decompressing the payload should:

1. Read 8 trailer bytes.
2. Decode CRC32 and ISIZE as little-endian u32.
3. Compare CRC32 against the running CRC of the bytes it produced.
4. Compare ISIZE against `(uncompressed_byte_count) mod 2^32`.
5. On mismatch, signal a corrupt-stream error.
6. Loop back to read another member if more bytes remain in the input.
