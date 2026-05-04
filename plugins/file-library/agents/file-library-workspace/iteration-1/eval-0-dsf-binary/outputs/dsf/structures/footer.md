# Footer

The fixed 16-byte (128-bit) MD5 file footer that closes every DSF file.

## Byte diagram

```
+---------------------------------------+
|       MD5 hash (16 bytes raw)         |
+---------------------------------------+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 16 | `[16]byte` | MD5 | RFC 1321 MD5 of every byte in the file before the footer (i.e., bytes `[0, fileSize − 16)`). |

## Checksums and integrity

- **Algorithm**: MD5 (RFC 1321), 128-bit digest.
- **Scope**: every byte of the file before this footer — file header (12
  bytes) plus the entire atom region.
- **Byte order**: the digest is written as the MD5 native big-endian byte
  ordering (i.e., the canonical `md5sum` 16-byte output, byte-for-byte).
- **Pseudo-header**: none.
- **Computation**: `md5.Sum(allBytesBeforeFooter)`.

## Ambiguities

> **Ambiguity:** The spec does not state how a writer should handle the case
> where the user is incrementally streaming output and only learns the byte
> count after the fact. In practice writers buffer the file in memory or in
> a temporary file, hash on close, and write the footer last.
