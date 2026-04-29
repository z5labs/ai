# Complex gzip member with all four optional flags

Exercises every optional metadata block: `FEXTRA`, `FNAME`, `FCOMMENT`, and
`FHCRC` are all set in `FLG`. The blocks appear in spec order between the
fixed header and the compressed data. Empty input is used so the deflate
body is the canonical `03 00` and the trailer holds CRC32 = 0 / ISIZE = 0.

Optional content:

- `ExtraField`: one subfield with `(SI1, SI2) = ('A','P')` (Apollo file type
  information per [`../encoding-tables/extra-subfield-ids.md`](../encoding-tables/extra-subfield-ids.md))
  carrying 4 bytes of opaque data `01 02 03 04`. The whole subfield record
  is 8 bytes; `XLEN = 8`.
- `FName`: `"data.bin"` (8 bytes + NUL).
- `FComment`: `"hi\n"` (3 bytes + NUL — note the embedded LF for line break
  per [`../structures/fcomment.md`](../structures/fcomment.md)).
- `FHCRC`: 2-byte CRC16 of every preceding header byte. Shown below as
  `?? ??` because computing it requires running the CRC-32 algorithm from
  [`../structures/fhcrc.md`](../structures/fhcrc.md) over bytes 0..n-1 and
  taking the low 16 bits. An implementation of this spec can derive the
  exact 2 bytes mechanically.

```
Offset    Hex                                                ASCII
00000000  1f 8b 08 1e 00 00 00 00  00 03 08 00 41 50 04 00  ............AP..
00000010  01 02 03 04 64 61 74 61  2e 62 69 6e 00 68 69 0a  ....data.bin.hi.
00000020  00 ?? ?? 03 00 00 00 00  00 00 00 00 00           .............
```

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0     | ID1 | `0x1F` | gzip magic byte 1 |
| 1     | ID2 | `0x8B` | gzip magic byte 2 |
| 2     | CM  | `0x08` | DEFLATE |
| 3     | FLG | `0x1E` | bits 1,2,3,4 set: `FHCRC` + `FEXTRA` + `FNAME` + `FCOMMENT`. Reserved bits 5–7 clear. See [`../structures/flg.md`](../structures/flg.md). |
| 4–7   | MTIME | `0x00000000` | "no timestamp available" |
| 8     | XFL | `0x00` | Unspecified |
| 9     | OS  | `0x03` | UNIX |
| 10–11 | XLEN | `0x0008` (LE bytes `08 00`) | Extra-field payload is 8 bytes (does not count XLEN itself) |
| 12    | Subfield.SI1 | `0x41` | `'A'` |
| 13    | Subfield.SI2 | `0x50` | `'P'` |
| 14–15 | Subfield.LEN | `0x0004` (LE bytes `04 00`) | 4 bytes of subfield data follow |
| 16–19 | Subfield.Data | `01 02 03 04` | Opaque per the registered Apollo subfield |
| 20–27 | FName.Name | `64 61 74 61 2e 62 69 6e` | `"data.bin"` |
| 28    | FName.Terminator | `0x00` | NUL |
| 29–31 | FComment.Comment | `68 69 0a` | `"hi\n"` — single LF byte (`0x0A`) for the line break, per RFC 1952 §2.3.1 |
| 32    | FComment.Terminator | `0x00` | NUL |
| 33–34 | FHCRC.CRC16 | `?? ??` | Low 16 bits of CRC-32 over bytes 0–32 (everything before this field). LE-stored. Compute per [`../structures/fhcrc.md`](../structures/fhcrc.md). |
| 35–36 | CompressedData | `03 00` | Canonical 2-byte deflate stream for empty input |
| 37–40 | CRC32 | `0x00000000` | CRC-32 of empty input |
| 41–44 | ISIZE | `0x00000000` | Uncompressed length 0 |

> **Note:** The `??` bytes are intentionally not concrete in this example
> because their value is derived deterministically from the rest of the
> header. A round-trip-correct encoder must compute them; a decoder under
> test must reproduce them and compare. See [`../structures/fhcrc.md`](../structures/fhcrc.md)
> for the exact computation steps.
