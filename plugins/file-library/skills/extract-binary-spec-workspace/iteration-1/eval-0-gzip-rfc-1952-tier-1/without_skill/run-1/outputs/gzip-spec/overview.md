# Overview: gzip stream structure (RFC 1952 §2)

A gzip file (a "gzip stream") consists of a series of **members**.
Each member is independently decodable and has the following layout:

```
+---------------+----------------------+----------------+
|   header      |  compressed blocks   |    trailer     |
| (10+ bytes)   |  (deflate; OUT OF    |   (8 bytes:    |
|               |   SCOPE for this     |  CRC32 + ISIZE)|
|               |   spec)              |                |
+---------------+----------------------+----------------+
```

A gzip file consists of one or more such members concatenated end to
end. There is no count or length prefix at the file level; decoders
detect end-of-file after consuming a member's trailer.

## 2.1 Overall conventions (verbatim from RFC 1952)

> In the diagrams below, a box like this:
>
> ```
> +---+
> |   | <-- the vertical bars might be missing
> +---+
> ```
>
> represents one byte; a box like this:
>
> ```
> +==============+
> |              |
> +==============+
> ```
>
> represents a variable number of bytes.
>
> Bytes stored within a computer do not have a "bit order", since they
> are always treated as a unit. However, a byte considered as an
> integer between 0 and 255 does have a most- and least-significant
> bit, and since we write numbers with the most-significant digit on
> the left, we also write bytes with the most-significant bit on the
> left. In the diagrams below, we number the bits of a byte so that
> bit 0 is the least-significant bit, i.e., the bits are numbered:
>
> ```
> +--------+
> |76543210|
> +--------+
> ```
>
> This document does not address the issue of the order in which bits
> of a byte are transmitted on a bit-sequential medium, since the data
> format described here is byte- rather than bit-oriented.
>
> Within a computer, a number may occupy multiple bytes. All multi-
> byte numbers in the format described here are stored with the
> **least-significant byte first** (at the lower memory address).
> For example, the decimal number 520 is stored as:
>
> ```
>     0        1
> +--------+--------+
> |00001000|00000010|
> +--------+--------+
>  ^        ^
>  |        |
>  |        + more significant byte = 2 x 256
>  + less significant byte = 8
> ```

## 2.2 File format

> A gzip file consists of a series of "members" (compressed data
> sets). The format of each member is specified in the following
> section. The members simply appear one after another in the file,
> with no additional information before, between, or after them.

## 2.3 Member format

```
+---+---+---+---+---+---+---+---+---+---+
|ID1|ID2|CM |FLG|     MTIME     |XFL|OS | (more-->)
+---+---+---+---+---+---+---+---+---+---+

(if FLG.FEXTRA set)

+---+---+=================================+
| XLEN  |...XLEN bytes of "extra field"...| (more-->)
+---+---+=================================+

(if FLG.FNAME set)

+=========================================+
|...original file name, zero-terminated...| (more-->)
+=========================================+

(if FLG.FCOMMENT set)

+===================================+
|...file comment, zero-terminated...| (more-->)
+===================================+

(if FLG.FHCRC set)

+---+---+
| CRC16 |
+---+---+

+=======================+
|...compressed blocks...| (more-->)
+=======================+

  0   1   2   3   4   5   6   7
+---+---+---+---+---+---+---+---+
|     CRC32     |     ISIZE     |
+---+---+---+---+---+---+---+---+
```

The fixed 10-byte prefix is always present. The four optional
sections appear, in this order, only when the corresponding `FLG` bit
is set. The compressed-blocks section is followed immediately by the
8-byte trailer.

## Conformance highlights

- **Magic.** `ID1 = 0x1f` (31) and `ID2 = 0x8b` (139). A decoder MUST
  reject any input where these bytes do not match.
- **Reserved bits.** Bits 5, 6, and 7 of `FLG` are reserved and
  **MUST be zero** in a compliant stream. A decoder SHOULD reject a
  member whose `FLG` has any reserved bit set, since it cannot know
  how to interpret the resulting member.
- **Multi-member files.** After successfully decoding one member's
  trailer, a decoder must look for another member; concatenating two
  valid gzip streams yields a valid gzip stream whose decompressed
  output is the concatenation of the two original outputs.
- **Compression method.** `CM = 8` denotes deflate. RFC 1952 reserves
  values 0-7 for future use; only `CM = 8` is defined.
- **Encoding of text fields.** `FNAME` and `FCOMMENT` are stored in
  ISO 8859-1 (LATIN-1) and terminated by a single zero byte. Line
  terminators in `FCOMMENT` are a single `0x0a` (LF).
