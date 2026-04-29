# Member

The top-level container in a gzip file. A gzip file is a concatenation of
one or more members; each member is a self-contained compressed dataset
with its own header, optional metadata, deflate payload, and trailer.

## Byte diagram

```
+--------+ +--------+ +--------+ +--------+ +--------+ +========+ +--------+
| header | | FEXTRA | | FNAME  | |FCOMMENT| | FHCRC  | |deflate | |trailer |
| 10 B   | | opt.   | | opt.   | | opt.   | | opt.   | |payload | |  8 B   |
+--------+ +--------+ +--------+ +--------+ +--------+ +========+ +--------+
   |           |           |          |          |         |          |
   v           v           v          v          v         v          v
 always    if FLG.    if FLG.    if FLG.    if FLG.    RFC 1951    always
           FEXTRA     FNAME      FCOMMENT   FHCRC      (out of
                                                       scope)
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 10 | (struct) | Header | Fixed header — see [`header.md`](header.md) |
| 10 | variable | (struct) | Extra | Optional FEXTRA — present iff `Header.FLG.FEXTRA` (bit 2) is set; see [`fextra.md`](fextra.md) |
| variable | variable | `[]byte` (NUL-terminated) | Name | Optional FNAME — present iff `Header.FLG.FNAME` (bit 3) is set; see [`fname.md`](fname.md) |
| variable | variable | `[]byte` (NUL-terminated) | Comment | Optional FCOMMENT — present iff `Header.FLG.FCOMMENT` (bit 4) is set; see [`fcomment.md`](fcomment.md) |
| variable | 2 | uint16 (LE) | HeaderCRC | Optional FHCRC — present iff `Header.FLG.FHCRC` (bit 1) is set; see [`fhcrc.md`](fhcrc.md) |
| variable | variable | `[]byte` | CompressedData | Deflate stream per RFC 1951 — out of scope for this reference |
| variable | 8 | (struct) | Trailer | Fixed trailer — see [`trailer.md`](trailer.md) |

## Variable-length fields

- **Length determination** for the entire member: there is **no top-level
  length field**. The decoder identifies the end of the deflate payload
  by deflate's own end-of-stream signal (the last block has BFINAL = 1),
  after which the trailer follows immediately.
- **Concatenated members:** if more bytes remain in the input stream
  after a member's trailer, those bytes are the start of the next
  member (a new `header` beginning with `1f 8b`).

## Conditional / optional fields

The presence of FEXTRA, FNAME, FCOMMENT, and FHCRC is each independently
controlled by a single bit in the header's FLG byte. See
[`flg.md`](flg.md) for the bit-to-field mapping.

When present, the optional fields appear in this fixed order:
**FEXTRA → FNAME → FCOMMENT → FHCRC**. They are never reordered.

When absent, the decoder skips that field entirely; no zero-length
placeholder is written.

## Nested structures

- [`header.md`](header.md) — fixed 10-byte prefix, always present
- [`flg.md`](flg.md) — bit-field decode of `Header.FLG`
- [`fextra.md`](fextra.md), [`fname.md`](fname.md),
  [`fcomment.md`](fcomment.md), [`fhcrc.md`](fhcrc.md) — optional, gated by FLG
- [`trailer.md`](trailer.md) — fixed 8-byte suffix, always present
