# Member trailer (8 bytes)

Immediately after the compressed-blocks section, every gzip member
ends with an 8-byte trailer:

```
  0   1   2   3   4   5   6   7
+---+---+---+---+---+---+---+---+
|     CRC32     |     ISIZE     |
+---+---+---+---+---+---+---+---+
```

| Offset | Size | Field   | Type    | Description |
| -----: | ---: | ------- | ------- | ----------- |
| 0      | 4    | `CRC32` | u32 LE  | CRC-32 of the **uncompressed** data. |
| 4      | 4    | `ISIZE` | u32 LE  | Size of the original (uncompressed) input data, modulo 2^32. |

Both fields are 4-byte unsigned integers stored **least-significant
byte first** (little-endian).

## CRC32 -- verbatim from RFC 1952

> CRC32 (CRC-32)
>    This contains a Cyclic Redundancy Check value of the
>    uncompressed data computed according to CRC-32 algorithm
>    used in the ISO 3309 standard and in section 8.1.1.6.2 of
>    ITU-T recommendation V.42. (See <http://www.iso.ch> for
>    ordering ISO documents. See <gopher://info.itu.ch> for an
>    online version of ITU-T V.42.)

The polynomial is the standard CRC-32 with reflected input/output
and final XOR of `0xFFFFFFFF`. In Go, this is
`hash/crc32.MakeTable(crc32.IEEE)` (or `crc32.IEEETable`).

A decoder MUST recompute CRC-32 over the decompressed bytes it
produced and compare against `CRC32`; mismatch is a fatal error.

## ISIZE -- verbatim from RFC 1952

> ISIZE (Input SIZE)
>    This contains the size of the original (uncompressed) input
>    data modulo 2^32.

### Decoder rules

- After successfully decompressing the member, the decoder MUST
  verify that `(uncompressed_byte_count mod 2^32) == ISIZE`.
- Because `ISIZE` is taken modulo 2^32, it is **not** a reliable
  length for inputs >= 4 GiB. Decoders MUST NOT treat `ISIZE` as
  the authoritative length of the decompressed stream when the
  decompressed data exceeds 4 GiB; the deflate stream's own
  end-of-block marker is the canonical end-of-data signal.

### Encoder rules

- Maintain a running CRC-32 and a 64-bit byte counter while
  feeding uncompressed bytes into the deflate compressor.
- On finalization, write `CRC32` and `(byte_count & 0xFFFFFFFF)` as
  little-endian 32-bit values.
- For multi-member files, each member has its own independent
  `CRC32` / `ISIZE` covering only that member's uncompressed payload.

## End-of-stream

There is **no** byte sequence beyond the trailer that marks
end-of-file. After consuming a trailer the decoder MUST attempt to
read another `ID1` byte:

- If the read returns `io.EOF` cleanly, the gzip stream is complete.
- If a new `ID1 = 0x1f` byte is read, a new member begins; loop back
  to header parsing.
- Any other byte (or a truncated header read) is a format error.
