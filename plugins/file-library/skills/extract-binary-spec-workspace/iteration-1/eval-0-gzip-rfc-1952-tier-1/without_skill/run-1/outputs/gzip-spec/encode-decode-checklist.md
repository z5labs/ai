# Go encoder / decoder conformance checklist

A condensed checklist for implementing gzip framing in Go. References
to RFC 1952 sections are in parentheses.

## Decoder

1. **Read magic** (§2.3): `ID1 == 0x1f && ID2 == 0x8b`. Reject otherwise.
2. **Read CM** (§2.3): only `CM == 8` (deflate) is currently defined.
3. **Read FLG** and validate (§2.3.1):
   - Reject if any of bits 5, 6, 7 are set ("compliant decompressor
     must give an error").
   - Capture `FTEXT`, `FHCRC`, `FEXTRA`, `FNAME`, `FCOMMENT` flags.
4. **Read MTIME, XFL, OS** (§2.3): all stored little-endian.
5. **If FEXTRA**: read 2-byte little-endian `XLEN`, then exactly
   `XLEN` bytes. Optionally walk the `(SI1, SI2, LEN, data)`
   subfield list; lengths MUST sum to `XLEN`.
6. **If FNAME**: read bytes until and including the first `0x00`.
   Treat the bytes (excluding terminator) as ISO 8859-1.
7. **If FCOMMENT**: same as FNAME, with optional CR/CRLF -> LF
   normalization for display.
8. **If FHCRC**: read 2-byte little-endian CRC16. Recompute CRC-32
   over all preceding header bytes (from `ID1` through the last byte
   before `FHCRC`); compare against the **low 16 bits** of that
   CRC-32. Reject on mismatch.
9. **Inflate** the deflate stream (out of scope here).
10. **Read trailer**: 4-byte little-endian `CRC32`, then 4-byte
    little-endian `ISIZE`.
11. **Verify**: CRC-32 of decompressed data matches `CRC32`; low 32
    bits of decompressed byte count match `ISIZE`.
12. **Loop**: attempt to read another `ID1` byte. EOF means done; a
    new `0x1f` starts another member.

## Encoder

1. **Write magic**: `0x1f`, `0x8b`.
2. **Write CM = 8** (deflate).
3. **Compute FLG** based on which optional fields you will emit:
   - `FTEXT` if you wish to flag text content (often left zero).
   - `FEXTRA` if any extra subfields are present.
   - `FNAME` if including the original file name.
   - `FCOMMENT` if including a comment.
   - `FHCRC` if including the header CRC16.
   - Reserved bits MUST be zero.
4. **Write MTIME** as a 4-byte little-endian Unix timestamp, or `0`
   for "no timestamp".
5. **Write XFL**: `2` for max compression, `4` for fastest, else
   `0`. (Informational; many encoders just write 0.)
6. **Write OS**: typically `3` (Unix) or `255` (unknown).
7. **If FEXTRA**: write `XLEN` (u16 LE) followed by exactly that
   many bytes of subfield data. Each subfield is
   `SI1 | SI2 | LEN(u16 LE) | LEN bytes of data`. Avoid `SI2 == 0`.
8. **If FNAME**: write LATIN-1 bytes followed by a single `0x00`.
9. **If FCOMMENT**: same encoding as FNAME; use LF (`0x0A`) line
   terminators.
10. **If FHCRC**: compute CRC-32 over every header byte written so
    far (from `ID1` through the end of `FCOMMENT` if present), take
    the low 16 bits, and write them little-endian.
11. **Compress payload** with deflate (RFC 1951; out of scope).
    While doing so, maintain:
    - a running CRC-32 over the *uncompressed* bytes
      (`hash/crc32.IEEETable`),
    - a 64-bit count of uncompressed bytes.
12. **Write trailer**: 4-byte LE `CRC32`, 4-byte LE
    `ISIZE = byte_count & 0xFFFFFFFF`.
13. **Multi-member output**: simply concatenate additional members.

## Common pitfalls

- **Endianness.** Every multi-byte field in the framing is
  little-endian. It is easy to accidentally use big-endian helpers.
- **CRC scopes.** `FHCRC` covers the *header*, and is a CRC-16
  (low 16 bits of CRC-32). The trailer `CRC32` covers the
  *uncompressed* data. They are not the same value.
- **Reserved FLG bits.** A decoder that silently ignores them is
  non-conformant.
- **`ISIZE` overflow.** Real-world gzip streams of >= 4 GiB
  uncompressed are common; do not gate decode on `ISIZE` matching
  a 64-bit length.
- **Multi-member streams.** Many real gzip files (e.g. concatenated
  log archives) contain multiple members; a decoder that stops after
  the first trailer will silently truncate output.
- **Empty `FNAME` / `FCOMMENT`.** A single `0x00` terminator with no
  preceding text is a valid (if odd) encoding.
- **`FEXTRA` framing vs. subfields.** `XLEN` always frames the entire
  extra-field block; never trust an internal subfield `LEN` to define
  the end of the block.
