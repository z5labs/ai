# Trailer

The trailer is fixed-size: 4 bytes.

**Byte order:** big-endian.

## Field table

| Offset (bytes) | Size | Type     | Name  | Description                                   |
|----------------|------|----------|-------|-----------------------------------------------|
| 0              | 4    | `uint32` | CRC32 | IEEE CRC32 of all bytes preceding the trailer. |

## Notes

- The CRC32 covers every byte from offset 0 (the start of the header) up to (but not including) the CRC32 itself.
- Decoders must compute the running CRC32 as they read, then compare it to the value in the trailer; on mismatch, return a typed error wrapping the leaf sentinel.
