# Trailer (CRC32 + ISIZE)

The fixed 8-byte suffix that ends every gzip member. Carries an integrity
check over the original uncompressed data (`CRC32`) and the size of that
data modulo 2^32 (`ISIZE`).

## Byte diagram

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                            CRC32                              |
+---------------------------------------------------------------+
|                            ISIZE                              |
+---------------------------------------------------------------+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 4 | uint32 (LE) | CRC32 | CRC-32 of the **uncompressed** data |
| 4 | 4 | uint32 (LE) | ISIZE | Size of the uncompressed data, modulo 2^32 |

Both fields are little-endian, matching the format-wide convention.

## Checksums and integrity

- **Algorithm:** CRC-32 (ITU-T V.42 / IEEE 802.3 / PKZIP). Reference
  implementation appears in RFC 1952 §8.
- **Polynomial:** `x^32 + x^26 + x^23 + x^22 + x^16 + x^12 + x^11 +
  x^10 + x^8 + x^7 + x^5 + x^4 + x^2 + x + 1` — reflected polynomial
  `0xEDB88320`.
- **Initial register:** `0xFFFFFFFF`.
- **Final XOR:** `0xFFFFFFFF`.
- **Bit order:** least-significant bit of each byte processed first
  (matches the reflected polynomial form).
- **Scope:** the entire **uncompressed** byte stream — i.e. the bytes
  the deflate decoder would emit. **Not** the deflate-compressed
  bytes, **not** the gzip header, **not** any optional fields.
- **Byte order of the CRC32 value:** little-endian.
- **Pseudo-header:** none.

## Padding and alignment

None. The 8 bytes are packed flush against the end of the deflate
payload. If another member follows, its `ID1` byte begins at the very
next offset.

## Versioning notes

`ISIZE` is taken **modulo 2^32**. For source files larger than 4 GiB the
true uncompressed size cannot be recovered from this field alone; tools
that need exact sizes must keep an out-of-band record or scan the
deflate stream. The CRC32 likewise reflects the entire uncompressed
stream regardless of size, but a 32-bit CRC has correspondingly weaker
collision resistance for very large inputs.

## Ambiguities

> **Ambiguity:** The RFC does not specify decoder action on CRC32
> mismatch or ISIZE mismatch. Standard practice (zlib, Go) is to
> report an error and reject the member.

> **Ambiguity:** For multi-member files, RFC 1952 does not say whether
> a decoder MUST validate every member's CRC32/ISIZE before emitting
> bytes. Streaming decoders typically validate as each member is
> finalized.
