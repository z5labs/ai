# FHCRC (header CRC16)

Optional 2-byte CRC integrity check over the gzip header bytes that
precede it. Appears after [FCOMMENT](fcomment.md) (if present) and
immediately before the deflate payload when `FLG.FHCRC` (bit 1) is set.

## Byte diagram

```
+---+---+
| CRC16 |   (uint16, little-endian)
+---+---+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 (LE) | CRC16 | Two least-significant bytes of the CRC-32 over all preceding header bytes |

## Conditional / optional fields

- **Condition:** present iff bit 1 (`FHCRC`) of [`flg.md`](flg.md) is set.
- **When present:** appears as the last 2 bytes of header framing — after
  any of `FEXTRA` / `FNAME` / `FCOMMENT` that are present, and immediately
  before the deflate-compressed payload.
- **When absent:** zero bytes.

## Checksums and integrity

- **Algorithm:** the same CRC-32 used for the trailer. CRC16 here is
  the **two least significant bytes** of the 32-bit result, taken
  after the standard final XOR — see [`trailer.md`](trailer.md) for the
  polynomial and computation.
- **Scope:** every byte from the start of the member (offset 0, the
  `ID1` byte of the header) up to and not including the 2 bytes of
  this CRC16 field. Concretely, the scope covers:
  - the 10-byte fixed header,
  - the FEXTRA block if present (`XLEN` plus its `XLEN` bytes of
    subfield data),
  - the FNAME bytes including the NUL terminator if present,
  - the FCOMMENT bytes including the NUL terminator if present.
- **Byte order of the CRC16 value:** little-endian (matches the format
  convention).
- **Pseudo-header:** none.
- **Computation:** compute CRC-32 over the bytes in scope, then take
  the low 16 bits (`crc32_value & 0xFFFF`) and store them as a uint16
  little-endian.

## Padding and alignment

None.

## Ambiguities

> **Ambiguity:** RFC 1952 §2.3.1.2 was historically read two ways:
> (a) CRC covers only the fixed 10-byte header, or (b) CRC covers
> all bytes preceding the CRC16 (including FEXTRA/FNAME/FCOMMENT).
> Errata and de-facto practice (zlib, Go's `compress/gzip`) settle on
> interpretation (b). Implementations following this reference SHOULD
> use (b).

> **Ambiguity:** The RFC does not specify decoder action on CRC
> mismatch. Standard practice is to reject the member.
