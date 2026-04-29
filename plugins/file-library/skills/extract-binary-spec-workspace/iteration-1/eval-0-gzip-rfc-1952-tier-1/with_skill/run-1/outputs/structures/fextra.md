# FEXTRA (extra field)

Optional structure that appears immediately after the fixed
[header](header.md) when `FLG.FEXTRA` (bit 2) is set. Carries a sequence
of length-prefixed subfields, each tagged with a 2-byte ID, intended for
implementation-specific or vendor-specific metadata.

## Byte diagram

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|             XLEN              |     subfield bytes (XLEN B)   ...
+---------------+---------------+-------------------------------+

Subfield (one of many, total bytes = XLEN):
+-------+-------+---------------+===============================+
|  SI1  |  SI2  |     LEN       |   LEN bytes of subfield data  |
+-------+-------+---------------+===============================+
   1 B     1 B    2 B (uint16 LE)         LEN bytes
```

## Field table

Top-level FEXTRA layout:

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 (LE) | XLEN | Total size, in bytes, of the subfield region that follows. Does **not** count itself. Range 0..65535 |
| 2 | XLEN | `[]Subfield` | Subfields | Concatenated subfields whose total size is exactly `XLEN` |

Per-subfield layout:

| Offset (within subfield) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | uint8 | SI1 | First byte of the subfield ID (LATIN-1 character by convention) |
| 1 | 1 | uint8 | SI2 | Second byte of the subfield ID; `SI2 = 0` denotes "random / local use" |
| 2 | 2 | uint16 (LE) | LEN | Size, in bytes, of the subfield data that follows. Does **not** count `SI1`, `SI2`, or itself |
| 4 | LEN | `[]byte` | Data | Subfield payload, opaque from the format's perspective |

A subfield's total wire size is `4 + LEN` bytes. The sum of `(4 + LEN)`
over all subfields equals `XLEN` exactly.

## Variable-length fields

- **`XLEN` length determination:** uint16 length prefix; XLEN counts only
  the subfield bytes, **not** itself (so the field consumes `2 + XLEN`
  bytes total on the wire).
- **`LEN` (per subfield) length determination:** uint16 length prefix;
  LEN counts only `Data`, **not** `SI1`, `SI2`, or `LEN` itself (so the
  subfield consumes `4 + LEN` bytes on the wire).
- **Maximum:** XLEN and LEN are each uint16 — maximum 65535 bytes.
- **Encoding (Data):** opaque bytes; interpretation depends on the
  `(SI1, SI2)` pair. See
  [`../encoding-tables/extra-subfield-ids.md`](../encoding-tables/extra-subfield-ids.md).

## Conditional / optional fields

- **Condition:** present iff bit 2 (`FEXTRA`) of [`flg.md`](flg.md) is set.
- **When present:** the layout above immediately follows the fixed
  10-byte header.
- **When absent:** zero bytes — there is no length placeholder. The
  decoder proceeds directly to FNAME / FCOMMENT / FHCRC / payload as
  gated by the remaining FLG bits.

## Padding and alignment

None. Subfields are packed tightly; there is no alignment requirement
and no padding between subfields or at the end of the FEXTRA block.

## Nested structures

- Subfield IDs are catalogued in
  [`../encoding-tables/extra-subfield-ids.md`](../encoding-tables/extra-subfield-ids.md).

## Ambiguities

> **Ambiguity:** RFC 1952 does not specify decoder behavior when the
> sum of `(4 + LEN)` across subfields disagrees with `XLEN`. Reasonable
> behavior: error out — the file is malformed.

> **Ambiguity:** It is unspecified whether duplicate `(SI1, SI2)`
> subfields in the same member are legal. Implementations typically
> preserve the first and ignore the rest, but the RFC is silent.
