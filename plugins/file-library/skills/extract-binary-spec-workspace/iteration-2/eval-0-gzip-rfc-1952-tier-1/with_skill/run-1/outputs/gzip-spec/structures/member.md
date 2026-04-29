# Member

A complete gzip member: fixed header, optional metadata blocks (selected by
the header's `FLG` byte), an opaque compressed-data body, and an 8-byte
trailer. A gzip stream is one or more members concatenated end-to-end with no
separators.

## Byte diagram

```
+===============+===============+========+==========+========+================+================+
| MemberHeader  | ExtraField    | FName  | FComment | FHCRC  | CompressedData | MemberTrailer  |
| (10 bytes)    | (if FEXTRA)   |(if .FNAME)|(if .FCOMMENT)|(if .FHCRC)| (variable)     |  (8 bytes)     |
+===============+===============+========+==========+========+================+================+
```

A field whose enabling flag in `FLG` is clear contributes zero bytes; the next
present field begins immediately.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 10 | MemberHeader | Header | Fixed-length identification + flags + timestamp + XFL + OS. See [`member-header.md`](member-header.md). |
| 10 | variable | ExtraField | Extra | Present iff [`flg.md`](flg.md) bit `FEXTRA` is set. See [`extra-field.md`](extra-field.md). |
| variable | variable | []byte | FName | Zero-terminated original file name, ISO 8859-1. Present iff `FLG.FNAME` is set. See [`fname.md`](fname.md). |
| variable | variable | []byte | FComment | Zero-terminated comment, ISO 8859-1. Present iff `FLG.FCOMMENT` is set. See [`fcomment.md`](fcomment.md). |
| variable | 2 | uint16 | FHCRC | CRC16 of the header bytes preceding this field. Present iff `FLG.FHCRC` is set. See [`fhcrc.md`](fhcrc.md). |
| variable | variable | []byte | CompressedData | Compressed payload. When `Header.CM == 8`, this is a deflate stream (RFC 1951). Opaque to the gzip framing layer. |
| variable | 8 | MemberTrailer | Trailer | CRC32 of the uncompressed data and ISIZE (input length mod 2^32). See [`member-trailer.md`](member-trailer.md). |

## Variable-length fields

- **CompressedData**: not length-prefixed by gzip itself. The decoder for the
  active compression method (e.g. deflate) consumes bytes from the stream
  until its own end-of-stream marker, then the gzip layer reads the
  `MemberTrailer` immediately after. RFC 1952 does not provide a separate
  byte length for this block — its length is determined entirely by the
  inner codec.

## Conditional / optional fields

- **ExtraField**: present iff `FLG.FEXTRA` (bit 2) is set.
- **FName**: present iff `FLG.FNAME` (bit 3) is set.
- **FComment**: present iff `FLG.FCOMMENT` (bit 4) is set.
- **FHCRC**: present iff `FLG.FHCRC` (bit 1) is set.

A clear flag means the corresponding block is **absent** (zero bytes); the
decoder must not attempt to read or skip a default-length placeholder.

## Nested structures

- [`member-header.md`](member-header.md) — fixed 10-byte prefix
- [`extra-field.md`](extra-field.md) — XLEN + subfield list
- [`fname.md`](fname.md) — original file name
- [`fcomment.md`](fcomment.md) — file comment
- [`fhcrc.md`](fhcrc.md) — header integrity check
- [`member-trailer.md`](member-trailer.md) — body integrity + length

## Versioning notes

- The on-the-wire member layout described here corresponds to gzip file
  format version 4.3 (RFC 1952). Earlier or vendor-modified streams are
  outside the scope of this reference.
- Reserved `FLG` bits (5, 6, 7) must be zero. A compliant decoder errors out
  if any reserved bit is non-zero, since such a bit could indicate a new
  optional block whose unknown length would desynchronize subsequent reads.

## Ambiguities

> **Ambiguity:** RFC 1952 does not specify the order in which optional blocks
> appear when more than one `FLG` flag is set. Section 2.3 lists them in the
> order `FEXTRA`, `FNAME`, `FCOMMENT`, `FHCRC`, and the byte diagram in §2.3
> matches that order. Implementations universally follow that order, but the
> spec text never uses the word "order" or "must". Treat the diagram order as
> normative.
