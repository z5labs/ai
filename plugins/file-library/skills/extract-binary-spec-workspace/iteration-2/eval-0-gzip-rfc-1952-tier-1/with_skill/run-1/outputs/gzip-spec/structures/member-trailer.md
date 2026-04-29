# MemberTrailer

The fixed 8-byte trailer that closes every gzip member. Carries the CRC32 of
the **uncompressed** input and the input length (modulo 2^32).

## Byte diagram

```
 0   1   2   3   4   5   6   7
+---+---+---+---+---+---+---+---+
|     CRC32     |     ISIZE     |
+---+---+---+---+---+---+---+---+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | uint32 | CRC32 | CRC-32 of the **uncompressed** original input bytes. Stored little-endian. |
| 4 | 4 | uint32 | ISIZE | Length in bytes of the **uncompressed** original input, taken modulo 2^32. Stored little-endian. |

## Checksums and integrity

- **Algorithm**: CRC-32 from ISO 3309 / ITU-T V.42, also referred to in
  RFC 1952 §2.3.1 as "the CRC-32 algorithm".
- **Polynomial / parameters** (from RFC 1952 §8 sample code):
  - Polynomial reflection: `0xEDB88320`
  - Initial register: `0xFFFFFFFF` (i.e. CRC accumulator XOR'd with all-ones at start)
  - Final XOR: `0xFFFFFFFF` (i.e. result XOR'd with all-ones at end)
  - Bit order: reflected input and output (the sample loop processes the
    least-significant bit of each input byte first).
- **Scope**: only the uncompressed original input is hashed. The gzip
  framing (header, optional fields, compressed-data bytes) is **not** part
  of the CRC32 — they are protected separately by `FHCRC` for the header
  and by the inner codec for the compressed data.
- **Byte order of the value**: little-endian.
- **Verification**: the decoder accumulates a CRC32 over the bytes it
  produces from decompressing `CompressedData`, then compares to this
  field. A mismatch is a corrupt stream.
- **Length verification**: the decoder also tracks the number of
  uncompressed bytes it produced, takes that count modulo 2^32, and
  compares to `ISIZE`. A mismatch is a corrupt stream.

## Ambiguities

> **Ambiguity:** Because `ISIZE` is the input length modulo 2^32, a member
> compressed from input larger than 4 GiB cannot have its full length
> verified from the trailer alone. RFC 1952 does not propose a workaround;
> implementations have generally just carried `length & 0xFFFFFFFF` as
> documented, and consumers needing exact lengths track them out-of-band.
