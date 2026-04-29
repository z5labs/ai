# Member layout (RFC 1952 Section 2.3)

The complete on-the-wire layout of a single gzip member, in order, is:

## 1. Fixed header (always present, 10 bytes)

    +---+---+---+---+---+---+---+---+---+---+
    |ID1|ID2|CM |FLG|     MTIME     |XFL|OS |
    +---+---+---+---+---+---+---+---+---+---+

## 2. Extra field (only if FLG.FEXTRA is set)

    +---+---+=================================+
    | XLEN  |...XLEN bytes of "extra field"...|
    +---+---+=================================+

`XLEN` is a little-endian uint16 giving the length in bytes of the extra
field that follows it. The XLEN bytes themselves are *not* included in
XLEN's count; the four byte cost (XLEN itself + body) means the on-disk
cost of the extra field when FEXTRA is set is `2 + XLEN` bytes.

## 3. Original file name (only if FLG.FNAME is set)

    +=========================================+
    |...original file name, zero-terminated...|
    +=========================================+

ISO 8859-1 (LATIN-1) bytes, terminated by a single 0x00 byte.

## 4. File comment (only if FLG.FCOMMENT is set)

    +===================================+
    |...file comment, zero-terminated...|
    +===================================+

ISO 8859-1 (LATIN-1) bytes, terminated by a single 0x00 byte.

## 5. Header CRC16 (only if FLG.FHCRC is set)

    +---+---+
    | CRC16 |
    +---+---+

A little-endian uint16 equal to the two least-significant bytes of the
CRC32 of *all* gzip header bytes that precede the CRC16 itself.

## 6. Compressed blocks (DEFLATE payload — out of scope here)

    +=======================+
    |...compressed blocks...|
    +=======================+

The framing is agnostic to the contents; for CM=8 ("deflate") this is a
DEFLATE stream as defined by RFC 1951.

## 7. Trailer (always present, 8 bytes)

      0   1   2   3   4   5   6   7
    +---+---+---+---+---+---+---+---+
    |     CRC32     |     ISIZE     |
    +---+---+---+---+---+---+---+---+

Both CRC32 and ISIZE are little-endian uint32 values. See `trailer.md`.

## Read order summary

A decoder reading a member must:

1. Read 10 fixed header bytes; verify ID1, ID2, CM.
2. Inspect FLG to determine which optional sections are present.
3. If FEXTRA: read XLEN (2 bytes), then XLEN bytes of subfields.
4. If FNAME: read bytes until (and including) a 0x00 terminator.
5. If FCOMMENT: read bytes until (and including) a 0x00 terminator.
6. If FHCRC: read 2 bytes; optionally verify against CRC32-of-header.
7. Read/decompress compressed blocks (DEFLATE for CM=8).
8. Read 8 trailer bytes (CRC32, ISIZE); verify against the
   uncompressed-data CRC and uncompressed length mod 2^32.
