# FName

The optional original-file-name block, present when [`flg.md`](flg.md) bit
`FNAME` is set. A zero-terminated ISO 8859-1 (Latin-1) string that records the
name of the file the payload was compressed from, with directory components
stripped.

## Byte diagram

```
+========================================+---+
|   ISO 8859-1 file-name bytes (no NUL)  | 0 |
+========================================+---+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | Name | ISO 8859-1 file-name bytes, **not** including the terminating NUL. Excludes any directory components. May be empty. |
| variable | 1 | uint8 | Terminator | Zero byte (`0x00`) that marks the end of the name. |

## Variable-length fields

- **Name**:
  - **Length determination**: sentinel-terminated. Read bytes until a `0x00`
    is encountered; that NUL is **not** part of `Name`.
  - **Maximum length**: not specified by RFC 1952 — bounded only by stream length.
  - **Encoding**: ISO 8859-1 (Latin-1). On systems whose native filename
    encoding is something else (EBCDIC, UTF-8, UTF-16), the encoder is
    expected to translate to Latin-1 before writing, and the decoder is
    expected to translate back when materializing the name on the local
    filesystem. If the source file lives on a case-insensitive filesystem,
    the encoder should force the name to lower case (RFC 1952 §2.3.1).

## Conditional / optional fields

- **Condition**: present iff `FLG.FNAME` (bit 3 of [`flg.md`](flg.md)) is set.
- **When absent**: zero bytes. The decoder must not attempt to read a name
  or the terminator.

## Ambiguities

> **Ambiguity:** RFC 1952 does not say whether `FNAME` may carry an empty
> name (a single `0x00` byte). The text states "an original file name is
> present, terminated by a zero byte" — an empty string followed by NUL fits
> that grammar. Decoders should accept it; encoders should leave `FNAME`
> clear when no source file name is available rather than emitting an empty
> name.

> **Ambiguity:** ISO 8859-1 is an 8-bit single-byte encoding with no
> embedded NULs in any defined codepoint, so the NUL terminator is
> unambiguous over any well-formed Latin-1 name. RFC 1952 does not address
> what a decoder should do if it encounters a NUL inside what was supposed
> to be a Latin-1 name; treat the first NUL as the terminator.
