# Optional header fields

These fields appear, in this fixed order, after the 10-byte member
header and before the compressed data, when their corresponding `FLG`
bit is set:

1. `FEXTRA` (when `FLG.FEXTRA` is set)
2. `FNAME` (when `FLG.FNAME` is set)
3. `FCOMMENT` (when `FLG.FCOMMENT` is set)
4. `FHCRC` (when `FLG.FHCRC` is set)

Skipping any one of these does not shift the others; each is
self-delimiting given that its presence bit is set.

## FEXTRA -- extra field

```
+---+---+=================================+
| XLEN  |...XLEN bytes of "extra field"...|
+---+---+=================================+
   2          XLEN
```

`XLEN` is a 2-byte little-endian unsigned integer giving the length in
bytes of the extra-field payload that follows. The payload itself is
a sequence of zero or more **subfields**, each of the form:

```
+---+---+---+---+==================================+
|SI1|SI2|  LEN  |... LEN bytes of subfield data ...|
+---+---+---+---+==================================+
  1   1     2              LEN
```

| Field    | Size | Description |
| -------- | ---: | ----------- |
| `SI1`    | 1    | Subfield ID byte 1. |
| `SI2`    | 1    | Subfield ID byte 2. |
| `LEN`    | 2    | Length of `data` in bytes, little-endian. |
| `data`   | LEN  | Subfield payload. |

### Verbatim text from RFC 1952

> 2.3.1.1. Extra field
>
>   If the FLG.FEXTRA bit is set, an "extra field" is present in the
>   header, with total length XLEN bytes. It consists of a series of
>   subfields, each of the form:
>
>      +---+---+---+---+==================================+
>      |SI1|SI2|  LEN  |... LEN bytes of subfield data ...|
>      +---+---+---+---+==================================+
>
>   SI1 and SI2 provide a subfield ID, typically two ASCII letters
>   with some mnemonic value. Jean-Loup Gailly <gzip@prep.ai.mit.edu>
>   is maintaining a registry of subfield IDs; please send him any
>   subfield ID you wish to use. Subfield IDs with SI2 = 0 are
>   reserved for future use. The following IDs are currently defined:
>
>      SI1         SI2         Data
>      ----------  ----------  ----
>      0x41 ('A')  0x70 ('P')  Apollo file type information
>
>   LEN gives the length of the subfield data, excluding the 4
>   initial bytes.

### Decoder rules

- Read `XLEN` first, then read exactly `XLEN` bytes; treat that slice
  as the extra-field payload.
- Inside the payload, parse subfields sequentially. The sum of the
  4 + `LEN` bytes of all subfields MUST equal `XLEN`. A short or
  over-long subfield indicates a malformed header.
- Unknown `(SI1, SI2)` pairs MUST be preserved (or skipped) but MUST
  NOT cause failure; the registry is open-ended.
- `SI2 = 0` is reserved; encoders SHOULD avoid emitting such IDs.

### Encoder rules

- `XLEN` is a 16-bit field, so the total extra-field payload cannot
  exceed 65535 bytes. Encoders MUST refuse longer payloads.
- Each subfield's `LEN` is also 16 bits; subfield payload is at most
  65531 bytes (since 4 bytes are consumed by the SI/LEN header) when
  it is the only subfield.

## FNAME -- original file name

```
+=========================================+
|...original file name, zero-terminated...|
+=========================================+
```

- Encoded in **ISO 8859-1 (LATIN-1)**.
- Terminated by exactly one `0x00` byte.
- The terminator is part of the encoded stream and MUST be consumed.
- Should be a basename only (no directory components), forced to
  lowercase on case-insensitive file systems.
- Empty names are permitted (a single `0x00` byte) but encoders
  typically clear `FLG.FNAME` rather than emit an empty name.
- Decoders converting the name to UTF-8 should map each byte directly
  to its Unicode code point (LATIN-1 -> Unicode is one-to-one for
  bytes 0x00-0xFF).

## FCOMMENT -- file comment

```
+===================================+
|...file comment, zero-terminated...|
+===================================+
```

- Encoded in **ISO 8859-1 (LATIN-1)**.
- Terminated by exactly one `0x00` byte.
- Line terminators within the comment SHOULD be a single LF
  (`0x0A`). CR (`0x0D`) and CRLF (`0x0D 0x0A`) are acceptable on
  input but SHOULD be normalized to LF for display.
- Intended for human consumption only; it is not interpreted by the
  decoder.

## FHCRC -- header CRC16

```
+---+---+
| CRC16 |
+---+---+
   2
```

- Stored as a 2-byte little-endian unsigned integer.
- Equals the **two least significant bytes** of the CRC-32 (the same
  polynomial used for the trailer's `CRC32`) computed over **all
  preceding bytes of the header**, starting at `ID1` and continuing
  up to (but not including) the `FHCRC` field itself. The covered
  region therefore includes the fixed 10 bytes plus any of `FEXTRA`,
  `FNAME`, `FCOMMENT` that are present.
- Verbatim:
  > If FHCRC is set, a CRC16 for the gzip header is present,
  > immediately before the compressed data. The CRC16 consists of
  > the two least significant bytes of the CRC32 for all bytes of
  > the gzip header up to and not including the CRC16. [Note that
  > at the time the CRC16 is being computed, the CRC32 of the
  > uncompressed data is not yet known.]
- A decoder MUST verify the CRC16 when `FLG.FHCRC` is set and reject
  the member on mismatch.
- This is **not** the same as the trailer `CRC32`, which is computed
  over the *uncompressed* data, not the header.
