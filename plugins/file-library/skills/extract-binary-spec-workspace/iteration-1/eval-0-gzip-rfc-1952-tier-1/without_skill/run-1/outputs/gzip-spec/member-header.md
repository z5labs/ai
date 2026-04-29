# Member header (fixed 10-byte prefix)

Every gzip member begins with a fixed 10-byte header, followed by the
optional fields enabled by `FLG`.

```
+---+---+---+---+---+---+---+---+---+---+
|ID1|ID2|CM |FLG|     MTIME     |XFL|OS |
+---+---+---+---+---+---+---+---+---+---+
  0   1   2   3   4   5   6   7   8   9
```

| Offset | Size (bytes) | Field   | Type   | Description |
| -----: | -----------: | ------- | ------ | ----------- |
| 0      | 1            | `ID1`   | u8     | Magic byte 1, fixed `0x1f` (31). |
| 1      | 1            | `ID2`   | u8     | Magic byte 2, fixed `0x8b` (139). |
| 2      | 1            | `CM`    | u8     | Compression method. |
| 3      | 1            | `FLG`   | u8     | Bit field; see [`flg-bits.md`](./flg-bits.md). |
| 4      | 4            | `MTIME` | u32 LE | Modification time, seconds since the Unix epoch. |
| 8      | 1            | `XFL`   | u8     | Extra flags specific to the compression method. |
| 9      | 1            | `OS`    | u8     | Operating system / file system on which compression took place. |

## ID1 / ID2 -- magic bytes

> These have the fixed values ID1 = 31 (0x1f, \\037), ID2 = 139
> (0x8b, \\213), to identify the file as being in gzip format.

A decoder MUST verify both bytes and reject any other value.

## CM -- compression method

> CM (Compression Method)
>    This identifies the compression method used in the file. CM
>    = 0-7 are reserved. CM = 8 denotes the "deflate" compression
>    method, which is the one customarily used by gzip and which
>    is documented elsewhere.

For the framing parser, the only valid value today is `CM = 8`. Any
other value should be treated as "unknown compression method".

## FLG -- flag byte

See the dedicated reference: [`flg-bits.md`](./flg-bits.md). Bits 5,
6, 7 are reserved and MUST be zero.

## MTIME -- modification time

> MTIME (Modification TIME)
>    This gives the most recent modification time of the original
>    file being compressed. The time is in Unix format, i.e.,
>    seconds since 00:00:00 GMT, Jan. 1, 1970. (Note that this may
>    cause problems for MS-DOS and other systems that use local
>    rather than universal time.) If the compressed data did not
>    come from a file, MTIME is set to the time at which compression
>    started. MTIME = 0 means no time stamp is available.
>
> The MTIME value is a 4 byte unsigned integer.

Stored little-endian. `MTIME = 0` means "no timestamp"; treat it as
`time.Time{}` / unset, not as `1970-01-01T00:00:00Z`, when round-tripping.

## XFL -- extra flags

`XFL` is interpreted in a way that depends on `CM`. For deflate
(`CM = 8`) the only defined values are:

> XFL (eXtra FLags)
>    These flags are available for use by specific compression
>    methods. The "deflate" method (CM = 8) sets these flags as
>    follows:
>
>      XFL = 2 - compressor used maximum compression, slowest algorithm
>      XFL = 4 - compressor used fastest algorithm

Other values are not assigned by RFC 1952 and should pass through
unchanged on decode/re-encode round-trips.

## OS -- operating system

> OS (Operating System)
>    This identifies the type of file system on which compression
>    took place. This may be useful in determining end-of-line
>    convention for text files. The currently defined values are
>    as follows:
>
>      0 - FAT filesystem (MS-DOS, OS/2, NT/Win32)
>      1 - Amiga
>      2 - VMS (or OpenVMS)
>      3 - Unix
>      4 - VM/CMS
>      5 - Atari TOS
>      6 - HPFS filesystem (OS/2, NT)
>      7 - Macintosh
>      8 - Z-System
>      9 - CP/M
>     10 - TOPS-20
>     11 - NTFS filesystem (NT)
>     12 - QDOS
>     13 - Acorn RISCOS
>    255 - unknown

A Go encoder typically writes `OS = 3` (Unix) on Unix builds and
`OS = 255` (unknown) when no good value is available; standard library
`compress/gzip` writes `255` by default.
