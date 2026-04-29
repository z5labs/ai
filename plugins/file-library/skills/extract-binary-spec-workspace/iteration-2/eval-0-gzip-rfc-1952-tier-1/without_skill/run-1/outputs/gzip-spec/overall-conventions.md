# Overall conventions (RFC 1952 Section 2.1)

## Byte size

A byte is exactly 8 bits (an octet), regardless of the host platform's
native character size.

## Diagram conventions

In the layout diagrams, a single-byte cell is drawn as:

    +---+
    |   |
    +---+

A variable-length region is drawn with double-bar edges:

    +==============+
    |              |
    +==============+

## Bit numbering within a byte

Within a byte considered as an integer 0..255, bit 0 is the
least-significant bit and bit 7 is the most-significant bit. The RFC
draws bytes most-significant-bit-first:

    +--------+
    |76543210|
    +--------+

The format is byte-oriented; the on-the-wire transmission order of bits
within a byte is not addressed.

## Multi-byte integer encoding (endianness)

All multi-byte integer fields are stored **little-endian** — the
least-significant byte comes first (at the lower file offset / lower
memory address).

Example: decimal 520 (= 0x0208) is stored as the two bytes:

        offset 0   offset 1
       +--------+--------+
       |00001000|00000010|
       +--------+--------+
        = 0x08    = 0x02
          (low)    (high, contributes 2 * 256)

This applies to every multi-byte integer in the format: MTIME (u32),
XLEN (u16), the FEXTRA subfield LEN (u16), the optional CRC16 (u16),
CRC32 (u32), and ISIZE (u32).
