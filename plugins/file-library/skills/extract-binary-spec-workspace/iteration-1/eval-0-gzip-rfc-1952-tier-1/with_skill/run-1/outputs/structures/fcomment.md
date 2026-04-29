# FCOMMENT (file comment)

Optional NUL-terminated LATIN-1 string carrying a human-readable comment
about the original file. Appears after [FNAME](fname.md) (if present) and
before [FHCRC](fhcrc.md) when `FLG.FCOMMENT` (bit 4) is set.

## Byte diagram

```
+================================+---+
|   ...LATIN-1 file comment...   | 0 |
+================================+---+
                                  ^
                           NUL terminator
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | `[]byte` | Comment | LATIN-1 comment text. May contain `0x0a` (LF) line breaks but no `0x00` bytes |
| variable | 1 | uint8 | NUL | Terminator — fixed value `0x00` |

The total wire size is `len(Comment) + 1` bytes.

## Variable-length fields

- **Length determination:** **sentinel** — terminated by a single
  `0x00` byte. There is no length prefix.
- **Length prefix format:** N/A (sentinel-terminated).
- **Maximum length:** unspecified. Implementations SHOULD impose a
  sanity cap.
- **Encoding:** ISO 8859-1 (LATIN-1). Line termination within the
  comment SHOULD use a single LF (`0x0a`) — even on systems whose
  native line ending is CR/LF or CR alone. Encoders convert before
  storage; decoders should leave LFs as-is.

## Conditional / optional fields

- **Condition:** present iff bit 4 (`FCOMMENT`) of [`flg.md`](flg.md) is set.
- **When present:** the bytes above immediately follow FNAME (if any),
  otherwise immediately follow FEXTRA (if any), otherwise immediately
  follow the fixed 10-byte header.
- **When absent:** zero bytes.

## Padding and alignment

None.

## Ambiguities

> **Ambiguity:** The RFC's "SHOULD use LF only" leaves it open whether
> a decoder must reject a comment that contains CR or CRLF. In
> practice, decoders pass the bytes through unchanged.
