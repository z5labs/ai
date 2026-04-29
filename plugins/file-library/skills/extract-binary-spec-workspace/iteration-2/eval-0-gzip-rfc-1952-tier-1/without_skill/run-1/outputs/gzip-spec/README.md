# GZIP File Format Spec (RFC 1952)

A clean reference for implementing a Go encoder/decoder for the gzip framing
format. Source: RFC 1952, "GZIP file format specification version 4.3"
(P. Deutsch, May 1996).

Source URL: https://www.rfc-editor.org/rfc/rfc1952.txt

## Scope of this extract

This directory covers the *framing* of a gzip stream only:

- The fixed member header (ID1, ID2, CM, FLG, MTIME, XFL, OS).
- The FLG bit field (FTEXT, FHCRC, FEXTRA, FNAME, FCOMMENT, reserved bits).
- The optional fields gated by FLG (FEXTRA, FNAME, FCOMMENT, FHCRC).
- The trailer (CRC32 + ISIZE).
- Endianness, byte numbering, bit numbering, and compliance rules.

It deliberately does **not** cover the DEFLATE compressed-data stream
(see RFC 1951 for that).

## Files

- `overall-conventions.md` — byte/bit numbering, endianness rules.
- `file-format.md` — top-level structure (a gzip file is a series of members).
- `member-layout.md` — ASCII layout diagrams for an entire member.
- `member-header.md` — fixed 10-byte header field reference.
- `flg-bits.md` — FLG flag-byte bit definitions and semantics.
- `optional-fields.md` — FEXTRA, FNAME, FCOMMENT, FHCRC, including the
  FEXTRA subfield format.
- `trailer.md` — CRC32 + ISIZE trailer.
- `compliance.md` — what compliant compressors and decompressors must do.
- `constants.md` — magic numbers, CM/OS values, bit positions, in a form
  convenient for Go constants.
