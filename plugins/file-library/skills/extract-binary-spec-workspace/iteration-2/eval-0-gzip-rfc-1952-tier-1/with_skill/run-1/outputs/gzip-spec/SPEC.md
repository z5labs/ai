# gzip Binary Specification Reference

## Overview
gzip is a stream / file format for compressed data. A gzip stream is a
concatenation of one or more independent **members**, each of which carries a
fixed identification header, optional metadata (extra field, original file
name, comment, header CRC16), a body of compressed data, and an 8-byte
trailer with an integrity CRC and the original input size. The format covered
here is gzip version 4.3 as specified in **RFC 1952** (May 1996). This
reference intentionally documents only the framing — the compressed-data body
itself is opaque to a gzip decoder and is specified separately in RFC 1951
(deflate).

## Conventions
- **Byte order**: little-endian. Multi-byte integers in the gzip header,
  extra-field length prefixes, and trailer are all stored
  least-significant-byte first (RFC 1952 §2.1).
- **Bit numbering**: LSB-0. Within a byte, bit 0 is the least-significant bit
  and bit 7 is the most-significant. This applies to the `FLG` and `XFL`
  flag bytes (RFC 1952 §2.1).
- **Size units**: bytes (octets). All lengths and offsets are in bytes.
- **Notation**: byte diagrams use the RFC 1952 box style — `+---+` boxes for
  one byte and `+===+` boxes for variable-length runs.
- **Character encoding for text fields**: ISO 8859-1 (Latin-1). Applies to
  `FNAME` and `FCOMMENT`.

## Top-level structure
A gzip stream is one or more concatenated **members**, with no separators or
trailer between them:

```
+--------+--------+--------+--------+--------+
| Member | Member | Member |  ...   | Member |
+--------+--------+--------+--------+--------+
```

Each member decomposes as:

```
+---------------+----------------+--------+----------+-------+----------+----------------+
| MemberHeader  | ExtraField?    | FName? | FComment?| FHCRC?| Compressed Data | MemberTrailer |
+---------------+----------------+--------+----------+-------+----------------+----------------+
   10 bytes      if FLG.FEXTRA   if .FNAME if .FCOMMENT if .FHCRC  variable           8 bytes
```

The presence of the `ExtraField`, `FName`, `FComment`, and `FHCRC` blocks is
controlled by individual bits of the header's `FLG` byte. If the corresponding
flag is clear, the block is absent (zero bytes) and the next block starts
immediately.

## Structures index
- [`structures/member.md`](structures/member.md) — top-level gzip member layout (header + optional fields + compressed body + trailer)
- [`structures/member-header.md`](structures/member-header.md) — fixed 10-byte header at the start of every member
- [`structures/flg.md`](structures/flg.md) — bit-packed `FLG` byte controlling which optional fields follow
- [`structures/extra-field.md`](structures/extra-field.md) — `XLEN`-prefixed optional extra field, present iff `FLG.FEXTRA` is set
- [`structures/extra-subfield.md`](structures/extra-subfield.md) — single subfield within the extra field (SI1, SI2, LEN, DATA)
- [`structures/fname.md`](structures/fname.md) — zero-terminated original file name, present iff `FLG.FNAME` is set
- [`structures/fcomment.md`](structures/fcomment.md) — zero-terminated comment, present iff `FLG.FCOMMENT` is set
- [`structures/fhcrc.md`](structures/fhcrc.md) — 2-byte header CRC16, present iff `FLG.FHCRC` is set
- [`structures/member-trailer.md`](structures/member-trailer.md) — fixed 8-byte trailer (CRC32 of uncompressed data + ISIZE mod 2^32)

## Encoding tables index
- [`encoding-tables/compression-methods.md`](encoding-tables/compression-methods.md) — values for the `CM` field
- [`encoding-tables/os-values.md`](encoding-tables/os-values.md) — values for the `OS` field
- [`encoding-tables/xfl-values.md`](encoding-tables/xfl-values.md) — values for the `XFL` field when `CM = 8` (deflate)
- [`encoding-tables/extra-subfield-ids.md`](encoding-tables/extra-subfield-ids.md) — registered subfield IDs (SI1, SI2)

## Examples index
- [`examples/minimal.md`](examples/minimal.md) — smallest legal member: no optional fields, empty input
- [`examples/typical.md`](examples/typical.md) — member with `FNAME` set
- [`examples/complex.md`](examples/complex.md) — member with `FEXTRA`, `FNAME`, `FCOMMENT`, and `FHCRC` all set

## Appendix
- **Maximum sizes**:
  - `XLEN` is `uint16`, so the extra field is at most 65535 bytes.
  - Each extra subfield's `LEN` is `uint16`, so a single subfield's data is at most 65535 bytes.
  - `FNAME` and `FCOMMENT` have no spec-imposed length cap; they are bounded only by the requirement to fit within the containing stream.
  - `ISIZE` is the original input length **modulo 2^32**, so it does not uniquely identify inputs larger than 4 GiB.
- **CRC algorithm**: CRC-32 from ISO 3309 / ITU-T V.42, polynomial reflection
  `0xEDB88320`, initial register `0xFFFFFFFF`, final XOR `0xFFFFFFFF`. See
  [`structures/member-trailer.md`](structures/member-trailer.md) and
  [`structures/fhcrc.md`](structures/fhcrc.md).
- **Related standards**:
  - RFC 1951 — DEFLATE compressed-data format (the body when `CM = 8`).
  - RFC 1950 — zlib container (sibling format, also wraps deflate).
  - ISO 3309 / ITU-T V.42 — CRC-32 specification.
- **Registry**: subfield IDs (SI1, SI2) are tracked by Jean-Loup Gailly per
  RFC 1952 §2.3.1.1; only one entry (`'A','P'` — Apollo file type information)
  was registered at the time RFC 1952 was published.
- **Version history**: RFC 1952 specifies gzip file format version 4.3,
  superseding earlier informal documentation distributed with the gzip
  utility itself.
