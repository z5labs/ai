# FLG bit field (RFC 1952 Section 2.3.1)

`FLG` is the fourth byte (offset 3) of the fixed header. Bit 0 is the
least-significant bit. The bit positions are:

| Bit | Mask  | Name      | Meaning |
|----:|------:|-----------|---------|
|   0 | 0x01  | FTEXT     | Source data is *probably* ASCII text. Hint only. |
|   1 | 0x02  | FHCRC     | A 16-bit header CRC follows the optional fields. |
|   2 | 0x04  | FEXTRA    | An "extra field" (XLEN + XLEN bytes) is present. |
|   3 | 0x08  | FNAME     | An original, zero-terminated file name is present. |
|   4 | 0x10  | FCOMMENT  | A zero-terminated comment is present. |
|   5 | 0x20  | reserved  | Must be zero. |
|   6 | 0x40  | reserved  | Must be zero. |
|   7 | 0x80  | reserved  | Must be zero. |

Order of optional sections in the byte stream when their flags are set
(this is the order in which a decoder must consume them):

    FEXTRA  ->  FNAME  ->  FCOMMENT  ->  FHCRC

## Per-flag semantics

### FTEXT (bit 0)
- A purely advisory hint that the uncompressed data is "probably ASCII
  text". Compressors may set it after sampling input.
- "In case of doubt, FTEXT is cleared, indicating binary data."
- Decoders may use it (e.g., to choose CRLF/LF translation) or ignore
  it. It does not affect framing.

### FHCRC (bit 1)
- When set, an additional 16-bit CRC for the gzip header is present
  immediately before the compressed data.
- The CRC16 is the **two least-significant bytes of the CRC32 over all
  bytes of the gzip header up to but not including the CRC16 itself**.
- Historical note from the RFC: gzip versions up to 1.2.4 never set this
  bit, even though gzip 1.2.4 documented it with a different meaning.
  Real-world readers should accept FHCRC absent, and verify it only when
  present.

### FEXTRA (bit 2)
- When set, a `XLEN` (u16 LE) and `XLEN` bytes of subfield data follow
  the fixed header.
- See `optional-fields.md` for the subfield structure.

### FNAME (bit 3)
- When set, the original file name is present (after the extra field if
  also present), as ISO 8859-1 (LATIN-1) bytes terminated by a single
  zero byte (`0x00`).
- The name is the original file name with directory components stripped.
  On case-insensitive file systems the source name is forced to lower
  case before being written.
- If the data did not come from a named file (e.g., piped from stdin),
  FNAME is not set and there is no file-name field.

### FCOMMENT (bit 4)
- When set, a free-form comment is present (after FEXTRA and FNAME if
  also present), again as ISO 8859-1 (LATIN-1) bytes terminated by a
  single zero byte.
- Intended for human consumption only — not interpreted by the format.
- Line breaks should be encoded as a single LF byte (decimal 10).

### Reserved bits (5, 6, 7)
- Must be written as zero by compressors.
- A compliant decoder **must signal an error if any reserved bit is
  non-zero**, because such a bit could indicate a new field that, if
  silently ignored, would cause subsequent data to be misinterpreted.

## Implementation hints (Go)

A typical Go decoder will read FLG into a `byte` and test:

    const (
        flagText    = 1 << 0 // FTEXT
        flagHCRC    = 1 << 1 // FHCRC
        flagExtra   = 1 << 2 // FEXTRA
        flagName    = 1 << 3 // FNAME
        flagComment = 1 << 4 // FCOMMENT
        flagReservedMask = 0xE0 // bits 5..7
    )

    if flg & flagReservedMask != 0 { return ErrReservedFlagSet }
    if flg & flagExtra != 0   { /* read XLEN + XLEN bytes */ }
    if flg & flagName != 0    { /* read NUL-terminated bytes */ }
    if flg & flagComment != 0 { /* read NUL-terminated bytes */ }
    if flg & flagHCRC != 0    { /* read 2 bytes; verify CRC16 */ }
