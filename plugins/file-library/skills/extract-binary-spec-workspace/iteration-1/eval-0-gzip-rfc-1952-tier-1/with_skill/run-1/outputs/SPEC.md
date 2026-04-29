# gzip Binary Specification Reference

## Overview

The gzip file format is a single-file compressed container originally designed
for the GNU `gzip` utility. A gzip file is a concatenation of one or more
"members"; each member carries a fixed header, optional metadata fields, a
deflate-compressed payload (RFC 1951), and an 8-byte trailer with an
uncompressed-data CRC32 and the uncompressed size mod 2^32.

This reference covers **framing only** — the bytes that wrap the compressed
payload. The deflate compressed-data format itself is specified by RFC 1951
and is intentionally out of scope here.

- **Standard:** RFC 1952 — "GZIP file format specification version 4.3"
  (P. Deutsch, May 1996).
- **Versions covered:** version 4.3 (CM = 8 / deflate); this is the only
  version in practical use.
- **Companion standard:** RFC 1951 (deflate) for the compressed payload
  between header and trailer.

## Conventions

- **Byte order:** **little-endian** for every multi-byte numeric field
  (`MTIME`, `XLEN`, subfield `LEN`, `CRC16`/FHCRC, `CRC32`, `ISIZE`). Every
  per-structure file in this reference inherits this convention; no
  structure overrides it.
- **Bit numbering:** **LSB-0** for the FLG flags byte and any other
  bit-packed field. Bit 0 is the least significant bit of the byte.
- **Size units:** bytes (octets). All length-prefix fields count bytes.
- **String encoding:** ISO 8859-1 (LATIN-1), NUL-terminated, for `FNAME`
  and `FCOMMENT`.
- **Notation:** Field tables use Go-friendly types (`uint8`, `uint16`,
  `uint32`, `[N]byte`, `[]byte`). ASCII wire diagrams are byte-oriented;
  bit-field diagrams label the bit numbers explicitly.

## Top-level structure

A gzip file is one or more concatenated members. Each member is independent
and self-describing — a decoder can stop after any member, or continue
reading the next member if more bytes remain.

```
gzip file
 ├── member 1
 │    ├── fixed header (10 bytes: ID1 ID2 CM FLG MTIME XFL OS)
 │    ├── FEXTRA   (optional, gated by FLG bit 2)
 │    ├── FNAME    (optional, gated by FLG bit 3)
 │    ├── FCOMMENT (optional, gated by FLG bit 4)
 │    ├── FHCRC    (optional, gated by FLG bit 1)
 │    ├── compressed data (deflate, RFC 1951 — out of scope)
 │    └── trailer  (8 bytes: CRC32 ISIZE)
 ├── member 2
 ...
```

Optional fields appear only when their FLG bit is set, and always in the
order listed above (FEXTRA, FNAME, FCOMMENT, FHCRC).

## Structures index

- [`structures/member.md`](structures/member.md) — top-level container: header + optional fields + deflate payload + trailer
- [`structures/header.md`](structures/header.md) — fixed 10-byte member header (ID1, ID2, CM, FLG, MTIME, XFL, OS)
- [`structures/flg.md`](structures/flg.md) — FLG bit field decoded into FTEXT/FHCRC/FEXTRA/FNAME/FCOMMENT + reserved bits
- [`structures/fextra.md`](structures/fextra.md) — optional extra field: XLEN-prefixed sequence of (SI1 SI2 LEN data) subfields
- [`structures/fname.md`](structures/fname.md) — optional NUL-terminated LATIN-1 original file name
- [`structures/fcomment.md`](structures/fcomment.md) — optional NUL-terminated LATIN-1 file comment
- [`structures/fhcrc.md`](structures/fhcrc.md) — optional 2-byte CRC16 over the preceding header bytes
- [`structures/trailer.md`](structures/trailer.md) — fixed 8-byte trailer: CRC32 of uncompressed data + ISIZE mod 2^32

## Encoding tables index

- [`encoding-tables/compression-methods.md`](encoding-tables/compression-methods.md) — CM byte values (0..7 reserved, 8 = deflate)
- [`encoding-tables/operating-system.md`](encoding-tables/operating-system.md) — OS byte values (0..13, 255)
- [`encoding-tables/xfl-deflate.md`](encoding-tables/xfl-deflate.md) — XFL byte values when CM = 8 (deflate-specific)
- [`encoding-tables/extra-subfield-ids.md`](encoding-tables/extra-subfield-ids.md) — FEXTRA SI1/SI2 conventions

## Examples index

- [`examples/minimal.md`](examples/minimal.md) — smallest valid member: no optional fields, empty payload
- [`examples/typical.md`](examples/typical.md) — typical member with FNAME and MTIME set
- [`examples/complex.md`](examples/complex.md) — all optional fields present (FEXTRA + FNAME + FCOMMENT + FHCRC) plus a multi-member file

## Appendix

### Maximum sizes and implementation limits

- **`XLEN`** (FEXTRA total size): uint16 — maximum 65535 bytes of subfields per member.
- **`LEN`** (per-subfield data size): uint16 — maximum 65535 bytes per subfield.
- **`FNAME` / `FCOMMENT`**: no explicit length prefix. Bounded only by the NUL terminator. Implementations SHOULD impose a reasonable cap (commonly 4096 bytes) to avoid pathological inputs.
- **`ISIZE`**: uint32 — original size **modulo 2^32**. Files larger than 4 GiB cannot have their true size recovered from the trailer alone.
- **Members**: any number of members may be concatenated; total file size is otherwise unbounded.

### Related standards

- RFC 1951 — DEFLATE compressed data format specification (the payload between header and trailer).
- RFC 1950 — ZLIB format (a sibling format using the same deflate compressor; not gzip).
- ITU-T Recommendation V.42 / IEEE 802.3 — defines the CRC-32 polynomial used by gzip.

### Registry references

The FEXTRA SI1/SI2 subfield-ID space is informally registered with the gzip
maintainers at the address listed in RFC 1952 §2.3.1.1. There is no IANA
registry for gzip subfield IDs. See
[`encoding-tables/extra-subfield-ids.md`](encoding-tables/extra-subfield-ids.md).

### Version history summary

- v4.3 (May 1996, RFC 1952) — current and only widely deployed version.

### Ambiguity callouts encountered

- **FHCRC scope.** RFC 1952 §2.3.1.2 was historically read two ways (fixed
  header only vs. all bytes preceding the CRC16). Errata and de-facto
  practice settle on "all preceding header bytes." See
  [`structures/fhcrc.md`](structures/fhcrc.md).
- **Reserved FLG bits.** The RFC says reserved bits "must be zero" but does
  not require the decoder to error if they are non-zero in older readings.
  Modern compliant decoders MUST reject. See
  [`structures/flg.md`](structures/flg.md).
