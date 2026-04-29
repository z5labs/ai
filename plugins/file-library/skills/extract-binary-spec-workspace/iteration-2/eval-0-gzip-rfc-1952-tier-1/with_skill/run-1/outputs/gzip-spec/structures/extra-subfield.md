# ExtraSubfield

A single subfield within [`extra-field.md`](extra-field.md). Each subfield is
a 4-byte header (two-byte ID plus a `uint16` length) followed by `LEN` bytes
of opaque subfield data.

## Byte diagram

```
 0   1   2   3
+---+---+---+---+==================================+
|SI1|SI2|  LEN  |   LEN bytes of subfield data    |
+---+---+---+---+==================================+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 1 | uint8 | SI1 | Subfield ID byte 1. Conventionally an ASCII letter. See [`../encoding-tables/extra-subfield-ids.md`](../encoding-tables/extra-subfield-ids.md). |
| 1 | 1 | uint8 | SI2 | Subfield ID byte 2. Conventionally an ASCII letter. `SI2 == 0` is reserved for future use. |
| 2 | 2 | uint16 | LEN | Length in bytes of the `Data` field that follows. **Excludes** the 4-byte subfield header. Stored little-endian. |
| 4 | LEN | []byte | Data | Subfield-specific payload, opaque to the gzip layer. Length is exactly `LEN`. |

## Variable-length fields

- **Data**:
  - **Length determination**: length-prefix.
  - **Length prefix format**: `LEN` is `uint16` little-endian. It counts only
    the data bytes; the 4 bytes of `(SI1, SI2, LEN)` are excluded.
  - **Maximum length**: 65535 bytes.
  - **Encoding**: opaque bytes; interpretation depends on `(SI1, SI2)`.

## Nested structures

Multiple `ExtraSubfield` records appear back-to-back inside the `Subfields`
payload of [`extra-field.md`](extra-field.md). The decoder stops when it has
consumed exactly `XLEN` bytes from the parent block.

## Ambiguities

> **Ambiguity:** RFC 1952 §2.3.1.1 reserves `SI2 == 0` for future use but
> does not say what a decoder should do if it sees one. Reading the `LEN`
> field and skipping the data preserves stream sync and is the safest
> behavior for an unknown subfield.
