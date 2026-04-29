# XFL (extra flags, deflate)

Maps the `XFL` byte at offset 8 of the [member header](../structures/header.md)
to a compression-quality hint. Only meaningful when `CM = 8` (deflate).
Advisory; decoders typically ignore it.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | default | No hint specified | RFC 1952 §2.3.1 |
| 2 | maxcompression | Compressor used the slowest, maximum-compression algorithm | RFC 1952 §2.3.1 |
| 4 | fastest | Compressor used the fastest algorithm | RFC 1952 §2.3.1 |

## Notes

- The RFC defines `XFL` only for `CM = 8`. For any other CM value the
  byte is unspecified.
- Values other than 0, 2, and 4 are unspecified for `CM = 8`. Decoders
  SHOULD ignore unknown values rather than reject the member.
- `XFL` is purely informational. Encoders MAY set it; decoders MUST
  NOT depend on it for correctness.
