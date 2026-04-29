# Typical gzip member

A realistic single-file gzip member: header with `MTIME` and `OS` set,
an `FNAME` original-file-name field, a small deflate-compressed payload,
and the trailer.

This example exercises:
- `FLG.FNAME = 1` only (FEXTRA, FCOMMENT, FHCRC absent)
- a non-zero `MTIME`
- `OS = 3` (Unix)
- a NUL-terminated LATIN-1 file name
- a non-trivial CRC32 and ISIZE in the trailer

The original file is named `hello.txt` and contains the 6 bytes
`hello\n` (0x68 0x65 0x6c 0x6c 0x6f 0x0a). It was created at Unix time
`0x5DA59B40` (2019-10-15 13:42:24 UTC).

```
Offset    Hex                                                ASCII
00000000  1f 8b 08 08 40 9b a5 5d  00 03 68 65 6c 6c 6f 2e  ....@..]..hello.
00000010  74 78 74 00 cb 48 cd c9  c9 e7 02 00 20 30 3a 36  txt..H...... 0:6
00000020  06 00 00 00                                       ....
```

(The deflate payload bytes `cb 48 cd c9 c9 e7 02 00` are a fixed-Huffman
deflate encoding of `hello\n`. The exact bytes are illustrative — any
valid deflate encoding of the same input is acceptable; only the framing
is normative for this reference.)

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | ID1 ID2 | `1f 8b` | gzip magic |
| 2 | CM | `08` | deflate |
| 3 | FLG | `08` | binary `0000 1000` — only `FNAME` (bit 3) set; see [`../structures/flg.md`](../structures/flg.md) |
| 4–7 | MTIME | `40 9b a5 5d` | uint32 LE = `0x5DA59B40` = 1571146944 = 2019-10-15 13:42:24 UTC |
| 8 | XFL | `00` | no hint |
| 9 | OS | `03` | Unix |
| 10–19 | FNAME | `68 65 6c 6c 6f 2e 74 78 74 00` | LATIN-1 `"hello.txt"` + NUL terminator |
| 20–27 | (deflate) | `cb 48 cd c9 c9 e7 02 00` | Deflate fixed-Huffman block encoding `hello\n` (RFC 1951; out of scope) |
| 28–31 | CRC32 | `20 30 3a 36` | uint32 LE = `0x363A3020`; CRC-32 of the 6 uncompressed bytes `hello\n` |
| 32–35 | ISIZE | `06 00 00 00` | uint32 LE = `6`; uncompressed size in bytes |

A decoder reading this member would:

1. Verify `ID1 ID2 = 1f 8b`.
2. Read `CM = 8` and prepare a deflate decoder.
3. Read `FLG = 0x08`, note `FNAME` is set, FEXTRA / FCOMMENT / FHCRC clear.
4. Read `MTIME`, `XFL`, `OS`.
5. Read bytes until NUL to recover the FNAME `hello.txt`.
6. Hand the deflate stream to the deflate decoder; it emits 6 bytes of
   uncompressed data and signals end-of-stream.
7. Read the trailer; verify `CRC32` over the 6 emitted bytes equals
   `0x363A3020` and `ISIZE` equals `6`.
