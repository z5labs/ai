# Optional header fields (RFC 1952 Sections 2.3 and 2.3.1.1)

When their corresponding `FLG` bit is set, these sections appear after
the 10-byte fixed header, in this order:

1. FEXTRA (if `FLG.FEXTRA`)
2. FNAME  (if `FLG.FNAME`)
3. FCOMMENT (if `FLG.FCOMMENT`)
4. FHCRC  (if `FLG.FHCRC`)

If a flag is clear, that section is *entirely absent* — no length-zero
placeholder. Sections that are present always appear in the order
above; they are not reorderable.

## FEXTRA — extra field

Layout:

    +---+---+=================================+
    | XLEN  |...XLEN bytes of "extra field"...|
    +---+---+=================================+

- `XLEN` is a little-endian unsigned 16-bit integer (range 0..65535).
- Exactly `XLEN` bytes of subfield data follow.
- Total cost on disk when FEXTRA is set: `2 + XLEN` bytes.
- The `XLEN` bytes are interpreted as a concatenation of zero or more
  **subfields**:

      +---+---+---+---+==================================+
      |SI1|SI2|  LEN  |... LEN bytes of subfield data ...|
      +---+---+---+---+==================================+

  - `SI1`, `SI2`: a 2-byte subfield ID. Conventionally two ASCII letters
    chosen for mnemonic value.
  - `LEN`: little-endian unsigned 16-bit length of the subfield's data
    payload, **excluding the 4 header bytes** (SI1, SI2, LEN). On-disk
    cost of one subfield is `4 + LEN` bytes.
  - Subfield IDs with `SI2 = 0` are reserved.
  - Subfield IDs are administered by Jean-Loup Gailly. The only ID
    spelled out in the RFC is:

        SI1 = 0x41 ('A'), SI2 = 0x70 ('P')  ->  Apollo file type info

- The sum of `4 + LEN` over all subfields must exactly equal `XLEN`.
  An encoder must keep them consistent; a decoder should treat a
  mismatch as malformed input.
- Decoders that do not recognise a subfield ID should skip its `LEN`
  bytes and continue.

## FNAME — original file name

Layout:

    +=========================================+
    |...original file name, zero-terminated...|
    +=========================================+

- Bytes are ISO 8859-1 (LATIN-1).
- On systems with non-LATIN-1 file names (e.g., EBCDIC), the encoder
  must transcode the name to LATIN-1 before writing.
- The name is the basename only — directory components are removed.
- On case-insensitive filesystems the encoder writes the name forced to
  lower case.
- The field is terminated by a single `0x00` byte. The terminator is
  part of the field (it is read but not part of the logical name).
- There is no length prefix; readers must scan for the NUL terminator.

## FCOMMENT — file comment

Layout:

    +===================================+
    |...file comment, zero-terminated...|
    +===================================+

- ISO 8859-1 (LATIN-1) bytes.
- Free-form, human-readable. Not interpreted by the format.
- Newlines should be encoded as a single LF byte (`0x0A`, decimal 10).
- Terminated by a single `0x00` byte (which is part of the on-disk
  field but not part of the logical comment text).
- No length prefix; readers must scan for the NUL terminator.

## FHCRC — header CRC16

Layout:

    +---+---+
    | CRC16 |
    +---+---+

- Little-endian unsigned 16-bit integer.
- Equal to the two least-significant bytes of the CRC32 (same algorithm
  as the trailer CRC32) computed over **all** bytes of the gzip header
  preceding the CRC16 itself. That includes:
  - The 10-byte fixed header.
  - The FEXTRA section (XLEN + body), if present.
  - The FNAME bytes including the NUL terminator, if present.
  - The FCOMMENT bytes including the NUL terminator, if present.
- It does **not** include the CRC16 bytes themselves nor any of the
  compressed payload nor the trailer.
- A decoder may verify or skip this value; the RFC does not require
  verification. If verification is performed and fails, treat the input
  as corrupt.
