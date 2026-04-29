# FNAME (original file name)

Optional NUL-terminated LATIN-1 string carrying the original file name
of the source file. Appears after [FEXTRA](fextra.md) (if present) and
before [FCOMMENT](fcomment.md) when `FLG.FNAME` (bit 3) is set.

## Byte diagram

```
+============================+---+
|  ...LATIN-1 file name...   | 0 |
+============================+---+
                              ^
                       NUL terminator
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | `[]byte` | Name | LATIN-1 file name, no NUL bytes inside |
| variable | 1 | uint8 | NUL | Terminator — fixed value `0x00` |

The total wire size is `len(Name) + 1` bytes.

## Variable-length fields

- **Length determination:** **sentinel** — terminated by a single
  `0x00` byte. There is no length prefix.
- **Length prefix format:** N/A (sentinel-terminated).
- **Maximum length:** unspecified by the RFC. Implementations SHOULD
  impose a sanity cap (4096 bytes is common) to avoid pathological
  inputs.
- **Encoding:** ISO 8859-1 (LATIN-1). The terminator is `0x00`. Names
  that cannot be represented in LATIN-1 are translated by the encoder
  before storage; only the base name (no directory components) is
  stored. On case-insensitive file systems, the encoder forces the
  name to lower case.

## Conditional / optional fields

- **Condition:** present iff bit 3 (`FNAME`) of [`flg.md`](flg.md) is set.
- **When present:** the bytes above immediately follow FEXTRA (if any),
  otherwise immediately follow the fixed 10-byte header.
- **When absent:** zero bytes — the decoder proceeds directly to
  FCOMMENT / FHCRC / payload as gated by the remaining FLG bits.

## Padding and alignment

None. The single `0x00` terminator is the entire framing overhead.

## Ambiguities

> **Ambiguity:** RFC 1952 does not specify whether a zero-length name
> followed by the terminator (i.e., a single `0x00` byte) is legal.
> Implementations typically treat it as an empty name rather than an
> error.

> **Ambiguity:** The RFC says "ISO 8859-1 (LATIN-1)" but does not
> specify decoder behavior for bytes outside the LATIN-1 printable
> range. Decoders typically pass the bytes through as-is and let the
> caller decide.
