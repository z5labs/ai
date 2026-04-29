# Compliance rules (RFC 1952 Section 2.3.1.2)

## Compliant compressor MUST

- Produce correct values for `ID1`, `ID2`, `CM`, `CRC32`, and `ISIZE`.
- Set all reserved `FLG` bits (bits 5, 6, 7) to zero.

## Compliant compressor MAY

- Set all other fixed-header fields to default values:
  - `OS = 255` (unknown).
  - All other defaults are `0` (i.e., `FLG = 0`, `MTIME = 0`,
    `XFL = 0`).

## Compliant decompressor MUST

- Verify `ID1`, `ID2`, and `CM`; report an error on mismatch.
- Examine `FEXTRA`/`XLEN`, `FNAME`, `FCOMMENT`, and `FHCRC` at least to
  the extent of being able to *skip past* those optional sections when
  they are present.
- Report an error if any reserved `FLG` bit is non-zero (since a new
  meaning for that bit could cause subsequent data to be misinterpreted
  if silently ignored).

## Compliant decompressor MAY

- Ignore `FTEXT` and `OS` entirely.
- Always produce binary output, regardless of `FTEXT`.
- Skip CRC16 verification when `FHCRC` is set (the RFC does not require
  it).

## Practical Go-implementation checklist

For an encoder:

- Always emit the 10-byte fixed header with valid magic and CM = 8 for
  DEFLATE.
- Compute and emit CRC32 (over uncompressed input) and ISIZE
  (uncompressed length mod 2^32) in the trailer.
- Reserved bits in FLG are always zero.
- If any optional section is included, the *order* must be FEXTRA,
  FNAME, FCOMMENT, FHCRC.
- If FHCRC is set, the CRC16 is computed over every header byte that
  precedes it, including the FEXTRA / FNAME / FCOMMENT bytes that are
  themselves only present because of other FLG bits.

For a decoder:

- Validate magic and CM.
- Reject any FLG with bits 5..7 set.
- Read optional sections in fixed order, gated by their FLG bits.
- Decompress until the DEFLATE stream signals end-of-stream.
- Verify CRC32 and ISIZE.
- Loop to handle multi-member files: continue reading members until
  EOF.
