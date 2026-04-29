# ExtraField

The optional extra-field block, present immediately after [`member-header.md`](member-header.md)
when [`flg.md`](flg.md) bit `FEXTRA` is set. Carries a length-prefixed
sequence of subfields, each of which is a `(SI1, SI2, LEN, data)` TLV
addressed by a 2-byte ID.

## Byte diagram

```
 0   1
+---+---+========================================+
| XLEN  |   XLEN bytes of subfield records       |
+---+---+========================================+
```

The `XLEN`-byte payload is parsed as a sequence of [`extra-subfield.md`](extra-subfield.md)
records until exactly `XLEN` bytes have been consumed.

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 | XLEN | Length in bytes of the subfield payload that follows. **Excludes** the 2 bytes of `XLEN` itself. Stored little-endian. |
| 2 | XLEN | []byte | Subfields | Concatenation of zero or more [`extra-subfield.md`](extra-subfield.md) records. Total length equals `XLEN`. |

## Variable-length fields

- **Subfields**:
  - **Length determination**: length-prefix. Read exactly `XLEN` bytes; do
    not run past that boundary.
  - **Length prefix format**: `XLEN` is `uint16` little-endian. It does **not**
    count itself.
  - **Maximum length**: 65535 bytes (the largest value of a `uint16`).
  - **Encoding**: opaque bytes; the inner subfields impose their own
    structure (see [`extra-subfield.md`](extra-subfield.md)).

## Conditional / optional fields

- **Condition**: present iff `FLG.FEXTRA` (bit 2 of [`flg.md`](flg.md)) is set.
- **When absent**: zero bytes; the next block (`FName`, `FComment`, `FHCRC`,
  or compressed data) starts immediately after the 10-byte header.

## Ambiguities

> **Ambiguity:** RFC 1952 does not say whether `XLEN = 0` is legal (an
> "extra field" that is present but empty). The grammar permits it — the
> length prefix is a `uint16` with no stated minimum — and a decoder
> reading exactly `XLEN` bytes will trivially handle it. Encoders should
> avoid emitting `XLEN = 0` with `FEXTRA` set, since it carries no
> information.
