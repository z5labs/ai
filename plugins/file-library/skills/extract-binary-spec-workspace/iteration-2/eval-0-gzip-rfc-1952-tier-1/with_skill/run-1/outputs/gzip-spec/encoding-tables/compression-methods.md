# Compression methods (CM)

Values for the `CM` field of [`../structures/member-header.md`](../structures/member-header.md).

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 1 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 2 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 3 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 4 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 5 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 6 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 7 | (reserved) | Reserved | RFC 1952 §2.3.1 |
| 8 | DEFLATE | Compressed data is a deflate stream | RFC 1952 §2.3.1; RFC 1951 |

## Notes

- Values 0–7 are reserved by RFC 1952. Values 9–255 are unassigned.
- All real-world gzip streams use `CM = 8` (deflate). A compliant decoder
  must error if any other value is present, since it has no way to
  interpret the body.
