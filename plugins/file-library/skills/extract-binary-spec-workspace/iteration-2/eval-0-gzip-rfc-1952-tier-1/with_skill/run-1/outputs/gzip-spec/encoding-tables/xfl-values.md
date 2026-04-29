# XFL values (CM = 8 / DEFLATE)

Values for the `XFL` field of [`../structures/member-header.md`](../structures/member-header.md).
The interpretation of `XFL` is per-compression-method. The values below apply
when `CM == 8` (deflate, the only widely deployed compression method); for
other `CM` values the meaning is unspecified.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | (default) | Unspecified — encoder did not signal a compressor preset | RFC 1952 §2.3.1 (default value, §2.3.1.2) |
| 2 | MAX | Compressor used maximum compression / slowest algorithm | RFC 1952 §2.3.1 |
| 4 | FAST | Compressor used fastest algorithm | RFC 1952 §2.3.1 |

## Notes

- RFC 1952 only defines the values `2` and `4` for deflate. Other values
  are not assigned. Encoders that do not have a preference should write
  `0` (RFC 1952 §2.3.1.2 lists `0` as a permitted default for "all other"
  fixed-length header fields).
- A compliant decoder is not required to interpret `XFL`; it is purely
  advisory.
