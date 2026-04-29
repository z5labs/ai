# Compression methods (CM byte)

Maps the `CM` byte at offset 2 of the [member header](../structures/header.md)
to a compression algorithm. Only `8` (deflate, RFC 1951) is in practical use.

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 1 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 2 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 3 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 4 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 5 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 6 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 7 | (reserved) | Reserved — decoders MUST reject | RFC 1952 §2.3.1 |
| 8 | deflate | RFC 1951 deflate stream | RFC 1951; RFC 1952 §2.3.1 |
| 9..255 | (unassigned) | Decoders MUST reject | RFC 1952 §2.3.1 |

## Notes

- All deployed gzip files use `CM = 8`. Encoders SHOULD always emit `8`.
- There is no IANA registry; the value space is owned by RFC 1952.
- A decoder that encounters any value other than 8 MUST treat the
  member as undecodable.
