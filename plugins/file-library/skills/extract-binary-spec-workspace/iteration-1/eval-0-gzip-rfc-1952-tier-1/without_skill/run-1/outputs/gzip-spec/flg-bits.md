# The FLG byte

`FLG` (offset 3 of the member header) is an 8-bit bit field that
signals the presence of optional fields and a header CRC. RFC 1952
numbers bits with bit 0 = least-significant.

```
       7 6 5 4 3 2 1 0
      +-+-+-+-+-+-+-+-+
FLG = |0|0|0|F|F|F|F|F|
      | | | |C|N|E|H|T|
      | | | |O|A|X|C|E|
      | | | |M|M|T|R|X|
      | | | |M|E|R|C|T|
      | | | |E| |A| | |
      | | | |N| | | | |
      | | | |T| | | | |
      +-+-+-+-+-+-+-+-+
       reserved (0)
```

## Verbatim bit assignments

> FLG (FLaGs)
>    This flag byte is divided into individual bits as follows:
>
>      bit 0   FTEXT
>      bit 1   FHCRC
>      bit 2   FEXTRA
>      bit 3   FNAME
>      bit 4   FCOMMENT
>      bit 5   reserved
>      bit 6   reserved
>      bit 7   reserved
>
> Reserved FLG bits must be zero.

| Bit | Mask  | Name       | Meaning when set |
| --: | ----- | ---------- | ---------------- |
| 0   | 0x01  | `FTEXT`    | The file is probably ASCII text. Advisory only; may be ignored. |
| 1   | 0x02  | `FHCRC`    | A 2-byte CRC16 of the header (everything from `ID1` up to but not including the `FHCRC` field itself) is present immediately before the compressed blocks. |
| 2   | 0x04  | `FEXTRA`   | An "extra field" (`XLEN` + data) is present after the fixed header. |
| 3   | 0x08  | `FNAME`    | A zero-terminated original file name (ISO 8859-1) is present. |
| 4   | 0x10  | `FCOMMENT` | A zero-terminated file comment (ISO 8859-1) is present. |
| 5   | 0x20  | reserved   | MUST be zero. |
| 6   | 0x40  | reserved   | MUST be zero. |
| 7   | 0x80  | reserved   | MUST be zero. |

## FTEXT semantics (verbatim)

> If FTEXT is set, the file is probably ASCII text. This is an
> optional indication, which the compressor may set by checking a
> small amount of the input data at the beginning of the file to
> see if any non-ASCII characters are present. In case of doubt,
> FTEXT is cleared, indicating binary data. For systems which have
> different file formats for ascii text and binary data, the
> decompressor can use FTEXT to choose the appropriate format. We
> deliberately do not specify the algorithm used to set this bit,
> since a compressor always has the option of leaving it cleared
> and a decompressor always has the option of ignoring it and
> letting some other program handle issues of text/binary format.

A decoder MAY ignore `FTEXT`; an encoder MAY always clear it.

## Field-presence bits

> If FHCRC is set, a CRC16 for the gzip header is present, immediately
> before the compressed data. The CRC16 consists of the two least
> significant bytes of the CRC32 for all bytes of the gzip header up
> to and not including the CRC16. [...]
>
> If FEXTRA is set, an "extra field" is present (see below).
>
> If FNAME is set, an original file name is present, terminated by a
> zero byte. The name must consist of ISO 8859-1 (LATIN-1) characters;
> on operating systems using EBCDIC or any other character set for
> file names, the name must be translated to the ISO LATIN-1 character
> set. This is the original name of the file being compressed, with
> any directory components removed, and, if the file being compressed
> is on a file system with case insensitive names, forced to lower
> case. There is no original file name if the file being compressed
> originated on a file system with no concept of file name, or if the
> name was lost during compression (e.g., compression was done on a
> stream rather than a file).
>
> If FCOMMENT is set, a "file comment" is present, terminated by a
> zero byte. The comment must consist of ISO 8859-1 (LATIN-1)
> characters. Line terminators should consist of a single line feed
> character (10 decimal); a CR LF sequence (13 decimal followed by 10
> decimal) or a single carriage return (13 decimal) is acceptable but
> should be converted to a line feed character (10 decimal) for
> display. The file comment is not interpreted; it is only intended
> for human consumption.

## Reserved bits -- decoder behavior

> If any reserved bits are non-zero, a compliant decompressor must
> give an error. Future versions of this specification may use these
> bits.

A correct gzip decoder MUST surface this as an error rather than
attempt to skip the unknown bit.
