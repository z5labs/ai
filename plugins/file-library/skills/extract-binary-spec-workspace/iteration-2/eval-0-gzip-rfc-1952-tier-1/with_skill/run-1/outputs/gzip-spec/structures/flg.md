# FLG

The `FLG` byte in [`member-header.md`](member-header.md) at offset 3. Each bit
selects whether a particular optional metadata block follows the fixed header.

**Layout:** 1 byte, bit-packed.

**Bit numbering:** LSB-0 (bit 0 is least-significant) — matches `SPEC.md#Conventions`.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | uint8 | FLG | Packed flag bits. See [Bit fields](#bit-fields). |

## Bit fields

| Bit(s) | Name | Description |
|---|---|---|
| 0 | FTEXT | If set, the original input is *probably* ASCII text. Advisory only — encoders may always leave it clear; decoders may always ignore it. |
| 1 | FHCRC | If set, a 2-byte header CRC16 ([`fhcrc.md`](fhcrc.md)) is present immediately after the other optional header blocks and before the compressed data. |
| 2 | FEXTRA | If set, an extra field ([`extra-field.md`](extra-field.md)) is present immediately after the fixed header. |
| 3 | FNAME | If set, a zero-terminated original file name ([`fname.md`](fname.md)) is present. |
| 4 | FCOMMENT | If set, a zero-terminated comment ([`fcomment.md`](fcomment.md)) is present. |
| 5 | reserved | Must be 0. A compliant decoder errors if set. |
| 6 | reserved | Must be 0. A compliant decoder errors if set. |
| 7 | reserved | Must be 0. A compliant decoder errors if set. |

When more than one optional flag is set, the corresponding blocks appear in
the order: `ExtraField` (FEXTRA), `FName` (FNAME), `FComment` (FCOMMENT),
`FHCRC` (FHCRC) — matching the diagram in [`member.md`](member.md).

## Ambiguities

> **Ambiguity:** RFC 1952 §2.3.1 notes that "the FHCRC bit was never set by
> versions of gzip up to 1.2.4, even though it was documented with a different
> meaning in gzip 1.2.4." Encoders targeting maximum compatibility with old
> decoders may prefer to leave `FHCRC` clear; decoders should not assume
> `FHCRC` is always clear in well-formed input.
