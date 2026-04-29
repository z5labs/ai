# gzip Framing Spec (RFC 1952)

A scoped reference for implementing a Go encoder/decoder for the gzip
file format **framing** (member header, FLG bit field, optional fields,
and trailer). The deflate compressed-data details (RFC 1951) are
intentionally **out of scope**.

Source: RFC 1952, "GZIP file format specification version 4.3",
P. Deutsch, May 1996. <https://www.rfc-editor.org/rfc/rfc1952.txt>

## Files

| File | Contents |
| ---- | -------- |
| [`overview.md`](./overview.md) | High-level structure of a gzip stream (members, byte order, conformance). |
| [`member-header.md`](./member-header.md) | Fixed 10-byte member header: `ID1`, `ID2`, `CM`, `FLG`, `MTIME`, `XFL`, `OS`. |
| [`flg-bits.md`](./flg-bits.md) | The `FLG` byte: `FTEXT`, `FHCRC`, `FEXTRA`, `FNAME`, `FCOMMENT`, reserved bits. |
| [`optional-fields.md`](./optional-fields.md) | `FEXTRA` (with `SI1`/`SI2`/`LEN`/data subfields), `FNAME`, `FCOMMENT`, `FHCRC`. |
| [`trailer.md`](./trailer.md) | Trailer: `CRC32` and `ISIZE`. |
| [`encode-decode-checklist.md`](./encode-decode-checklist.md) | Quick conformance checklist for a Go encoder/decoder. |

## What is *not* covered

- The deflate compressed block format (RFC 1951).
- The CRC-32 computation appendix in RFC 1952 (use Go's
  `hash/crc32.IEEE` polynomial).
- The `gzip` program's user interface or command-line behavior.

## Byte order at a glance

All multi-byte integers in the gzip framing are stored
**little-endian, least-significant byte first**. Bit 0 of a byte is the
**least-significant** bit. See [`overview.md`](./overview.md) for the
exact wording from the RFC.
