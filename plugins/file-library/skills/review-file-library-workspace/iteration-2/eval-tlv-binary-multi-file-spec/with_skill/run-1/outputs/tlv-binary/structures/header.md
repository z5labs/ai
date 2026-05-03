# Header

The header is fixed-size: 8 bytes. It identifies the file and carries format-wide flags.

**Byte order:** big-endian (matches global convention).

## Field table

| Offset (bytes) | Size | Type      | Name     | Description                             |
|----------------|------|-----------|----------|-----------------------------------------|
| 0              | 4    | `[4]byte` | Magic    | ASCII `"TLV1"` (`0x54 0x4C 0x56 0x31`). |
| 4              | 1    | `uint8`   | Version  | Format version. Must be `1`.            |
| 5              | 1    | `Flags`   | Flags    | Bit field; see below.                   |
| 6              | 2    | `uint16`  | Reserved | Reserved for future use. Must be `0`.   |

## Header.Flags bit field

A single byte holding three independent boolean flags.

| Bit(s) | Mask  | Name       | Description                                  |
|--------|-------|------------|----------------------------------------------|
| 0      | 0x01  | COMPRESSED | Record values are zlib-compressed.           |
| 1      | 0x02  | ENCRYPTED  | Record values are AES-encrypted.             |
| 2      | 0x04  | SIGNED     | The trailer carries a signature in addition to the CRC. |
| 3-7    | 0xF8  | (reserved) | Must be 0.                                   |

## Notes

- The Reserved field must be zero on write; readers must fail with a typed error if Reserved ≠ 0.
- The SIGNED flag is reserved for a future signed-trailer extension; readers must fail if encountered.
