# Constants quick reference

A flat list of every magic number, bit mask, and well-known value
defined by RFC 1952 framing, suitable for transcription into Go.

## Magic / identification

| Constant | Value | Meaning |
|----------|------:|---------|
| `ID1`    | `0x1f` | First magic byte of every gzip member. |
| `ID2`    | `0x8b` | Second magic byte of every gzip member. |

## Compression method (CM)

| Value | Meaning |
|-----:|---------|
| 0..7 | reserved |
|    8 | DEFLATE (RFC 1951) |

## FLG bit masks

| Bit | Mask  | Name     |
|----:|------:|----------|
|   0 | 0x01  | FTEXT    |
|   1 | 0x02  | FHCRC    |
|   2 | 0x04  | FEXTRA   |
|   3 | 0x08  | FNAME    |
|   4 | 0x10  | FCOMMENT |
|   5 | 0x20  | reserved |
|   6 | 0x40  | reserved |
|   7 | 0x80  | reserved |

Composite mask of reserved bits: `0xE0`.

## XFL values for CM = 8 (DEFLATE)

| Value | Meaning |
|-----:|---------|
|     2 | maximum compression / slowest |
|     4 | fastest |

## OS values

| Value | OS / filesystem |
|-----:|-----------------|
|     0 | FAT (MS-DOS, OS/2, NT/Win32) |
|     1 | Amiga |
|     2 | VMS / OpenVMS |
|     3 | Unix |
|     4 | VM/CMS |
|     5 | Atari TOS |
|     6 | HPFS (OS/2, NT) |
|     7 | Macintosh |
|     8 | Z-System |
|     9 | CP/M |
|    10 | TOPS-20 |
|    11 | NTFS (NT) |
|    12 | QDOS |
|    13 | Acorn RISCOS |
|   255 | unknown |

## Field sizes

| Field | Size (bytes) | Encoding |
|-------|-------------:|----------|
| ID1   | 1 | u8 |
| ID2   | 1 | u8 |
| CM    | 1 | u8 |
| FLG   | 1 | u8 |
| MTIME | 4 | u32, little-endian |
| XFL   | 1 | u8 |
| OS    | 1 | u8 |
| XLEN  | 2 | u16, little-endian (only if FEXTRA) |
| FEXTRA body | XLEN | bytes |
| FNAME | variable | LATIN-1 + 0x00 terminator (only if FNAME) |
| FCOMMENT | variable | LATIN-1 + 0x00 terminator (only if FCOMMENT) |
| CRC16 | 2 | u16, little-endian (only if FHCRC) |
| CRC32 | 4 | u32, little-endian |
| ISIZE | 4 | u32, little-endian |

Fixed header total: 10 bytes. Trailer total: 8 bytes.

## FEXTRA subfield layout

    +---+---+---+---+==================================+
    |SI1|SI2|  LEN  |... LEN bytes of subfield data ...|
    +---+---+---+---+==================================+

- `SI1`, `SI2`: u8 each. ID is the pair (SI1, SI2). `SI2 = 0` reserved.
- `LEN`: u16 little-endian, length of the subfield data, *excluding*
  the 4 ID/length header bytes.
- On-disk cost of one subfield: `4 + LEN`.
- The sum of `4 + LEN` over all subfields must equal `XLEN`.

Defined subfield IDs (from the RFC):

| SI1 | SI2 | Meaning |
|----:|----:|---------|
| `0x41` ('A') | `0x70` ('P') | Apollo file type information |

## CRC32 algorithm

- ISO 3309 / ITU-T V.42 CRC-32.
- Reflected polynomial: `0xEDB88320`.
- Initial value: `0xFFFFFFFF`.
- Final XOR: `0xFFFFFFFF`.
- Equivalent to Go's `hash/crc32.IEEE`.

## Suggested Go declarations

    const (
        gzipID1 = 0x1f
        gzipID2 = 0x8b

        cmDeflate = 8

        flagText    = 1 << 0 // FTEXT
        flagHCRC    = 1 << 1 // FHCRC
        flagExtra   = 1 << 2 // FEXTRA
        flagName    = 1 << 3 // FNAME
        flagComment = 1 << 4 // FCOMMENT

        flagReservedMask = 0xE0

        osUnknown = 255
    )
