# FHCRC

The optional 2-byte header CRC16, present when [`flg.md`](flg.md) bit `FHCRC`
is set. Sits immediately after the last present optional metadata block
(`ExtraField` / `FName` / `FComment`) and immediately before the compressed
data.

## Byte diagram

```
 0   1
+---+---+
| CRC16 |
+---+---+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 | CRC16 | The two **least-significant** bytes of the CRC32 of every header byte preceding this field. Stored little-endian. |

## Conditional / optional fields

- **Condition**: present iff `FLG.FHCRC` (bit 1 of [`flg.md`](flg.md)) is set.
- **When absent**: zero bytes.

## Checksums and integrity

- **Algorithm**: CRC-32 (ISO 3309 / ITU-T V.42), the same algorithm used in
  [`member-trailer.md`](member-trailer.md). The 16-bit value stored here is
  `crc32(header) & 0xFFFF` — i.e. the low 16 bits of the full CRC32, not a
  separate CRC-16 polynomial.
- **Scope**: every byte of the header from `ID1` (offset 0 of the member) up
  to and **not including** the CRC16 field itself. This covers the fixed
  10-byte [`member-header.md`](member-header.md) plus any optional blocks
  that precede `FHCRC`: [`extra-field.md`](extra-field.md) (when present),
  [`fname.md`](fname.md) (when present, including its NUL terminator), and
  [`fcomment.md`](fcomment.md) (when present, including its NUL terminator).
- **Byte order of the value**: little-endian (matching the format-wide
  convention).
- **Pseudo-header**: none.
- **Computation**:
  1. Run CRC-32 over the byte range described above with the standard gzip
     CRC-32 (initial register `0xFFFFFFFF`, polynomial reflection
     `0xEDB88320`, final XOR `0xFFFFFFFF`).
  2. Take the low 16 bits of the result.
  3. Store little-endian.

## Ambiguities

> **Ambiguity:** RFC 1952 §2.3.1 notes that "the FHCRC bit was never set by
> versions of gzip up to 1.2.4, even though it was documented with a
> different meaning in gzip 1.2.4." A decoder accepting older streams may
> see `FHCRC` set with a CRC computed differently than this spec. There is
> no portable way to recover from this; treat any mismatch as a corrupt
> stream.
