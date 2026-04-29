# FLG (flag byte)

The single-byte `FLG` field at offset 3 of the [member header](header.md).
Five bits gate optional metadata fields and one advisory text hint; the
remaining three bits are reserved and must be zero.

**Bit numbering:** LSB-0 (bit 0 is the least significant bit of the byte;
bit 7 is the most significant). Matches the format-wide convention from
`SPEC.md#Conventions`.

## Byte diagram

```
 7   6   5   4         3       2        1       0
+---+---+---+----------+-------+--------+-------+-------+
| 0 | 0 | 0 | FCOMMENT | FNAME | FEXTRA | FHCRC | FTEXT |
+---+---+---+----------+-------+--------+-------+-------+
 reserved (must be 0)
```

## Bit fields

| Bit(s) | Name | Description |
|---|---|---|
| 0 | FTEXT | Advisory: payload is probably ASCII text. No effect on decoding |
| 1 | FHCRC | If set, a 2-byte CRC16 follows the optional name/comment fields and precedes the deflate payload — see [`fhcrc.md`](fhcrc.md) |
| 2 | FEXTRA | If set, an `XLEN`-prefixed extra-field block follows the fixed header — see [`fextra.md`](fextra.md) |
| 3 | FNAME | If set, a NUL-terminated LATIN-1 original file name follows — see [`fname.md`](fname.md) |
| 4 | FCOMMENT | If set, a NUL-terminated LATIN-1 file comment follows — see [`fcomment.md`](fcomment.md) |
| 5 | (reserved) | Must be zero. Decoders MUST reject the member if non-zero |
| 6 | (reserved) | Must be zero. Decoders MUST reject the member if non-zero |
| 7 | (reserved) | Must be zero. Decoders MUST reject the member if non-zero |

The five defined flags are independent — any combination is legal. Note
that the optional fields (FEXTRA, FNAME, FCOMMENT, FHCRC), when present,
**always appear in that fixed order** regardless of bit ordering.

## Conditional / optional fields

Each set bit signals presence of exactly one optional field after the
fixed header:

| Bit | When set | Effect |
|---|---|---|
| 1 (FHCRC) | The 2-byte header CRC16 is present after FNAME/FCOMMENT (if any), immediately before the deflate payload |
| 2 (FEXTRA) | An `XLEN`-prefixed extra-field block follows the fixed header before FNAME |
| 3 (FNAME) | A NUL-terminated original file name string follows FEXTRA (if any) |
| 4 (FCOMMENT) | A NUL-terminated comment string follows FNAME (if any) |

## Ambiguities

> **Ambiguity:** RFC 1952 §2.3.1.2 states reserved bits "must be zero,"
> but the original wording does not unambiguously require the decoder to
> error. Modern reference implementations (zlib, Go's `compress/gzip`)
> reject members with any reserved FLG bit set; new implementations
> SHOULD do the same.

> **Ambiguity:** `FTEXT` is purely advisory. The RFC permits an encoder
> to set it heuristically and a decoder to ignore it. Implementations
> should not rely on it being accurate.
