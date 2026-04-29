# Byte and Bit Order Conventions

From RFC 1035 section 2.3.2 "Data Transmission Order" and section 4.

## Word size in diagrams

All ASCII diagrams in RFC 1035 are drawn 32 bits (4 octets / 4 bytes) wide:

```
                                    1  1  1  1  1  1
      0  1  2  3  4  5  6  7  8  9  0  1  2  3  4  5
    +--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
```

Each tick on the top is a bit position. Two such 16-bit halves form a 32-bit
row in most diagrams; some are drawn 16 bits wide (one 16-bit field per row).

## Byte order

DNS uses **network byte order (big-endian)** for all multi-octet integer
fields. The most significant octet of a 16-bit or 32-bit field is transmitted
first.

For example, a 16-bit field with value 0x1234 is transmitted as the byte
sequence `0x12, 0x34`.

## Bit order within a byte

Within an octet, bit 0 is the **most significant bit** (MSB) and bit 7 is the
**least significant bit** (LSB), per the diagram convention. This matters for
the bit-packed flags word in the header (see `header.md`): the QR bit is bit 0
of byte 2 of the header, which is the MSB (value 0x80) of that byte.

## Field alignment

Fields in the header and in each section are octet-aligned. There is no
padding between sections or between RRs. All section and RR concatenation is
contiguous on the wire.
