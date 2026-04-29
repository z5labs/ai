# FComment

The optional human-readable comment block, present when [`flg.md`](flg.md) bit
`FCOMMENT` is set. A zero-terminated ISO 8859-1 (Latin-1) string. The gzip
layer does not interpret the contents.

## Byte diagram

```
+========================================+---+
|  ISO 8859-1 comment bytes (no NUL)     | 0 |
+========================================+---+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | variable | []byte | Comment | ISO 8859-1 comment bytes, **not** including the terminating NUL. May contain line breaks encoded as a single line-feed byte (`0x0A`, decimal 10). |
| variable | 1 | uint8 | Terminator | Zero byte (`0x00`) that marks the end of the comment. |

## Variable-length fields

- **Comment**:
  - **Length determination**: sentinel-terminated. Read bytes until a `0x00`
    is encountered; that NUL is **not** part of `Comment`.
  - **Maximum length**: not specified by RFC 1952.
  - **Encoding**: ISO 8859-1 (Latin-1). RFC 1952 §2.3.1 specifies that line
    breaks within the comment "should be denoted by a single line feed
    character (10 decimal)" — i.e. encoders should not embed CR or CRLF
    sequences for newlines.

## Conditional / optional fields

- **Condition**: present iff `FLG.FCOMMENT` (bit 4 of [`flg.md`](flg.md)) is set.
- **When absent**: zero bytes.

## Ambiguities

> **Ambiguity:** RFC 1952 says the comment "is not interpreted" but also
> recommends LF-only line breaks. A decoder that materializes the comment
> on a CRLF-native platform may reasonably translate; a strict round-trip
> decoder must not. Implementations should expose the raw bytes and let the
> caller decide.
