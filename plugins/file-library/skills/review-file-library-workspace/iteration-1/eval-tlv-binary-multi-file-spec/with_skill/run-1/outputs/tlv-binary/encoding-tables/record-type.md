# Record.Type values

Encoding for the `RecordType` enum used in `../structures/record.md`.

| Value | Name   | Meaning                                |
|-------|--------|----------------------------------------|
| 0x01  | STRING | UTF-8 string (no null terminator).     |
| 0x02  | INT    | Big-endian signed 64-bit integer (Length must be 8). |
| 0x03  | BLOB   | Opaque byte payload.                   |
| 0x04  | NESTED | Value is itself a TLV1 file (header + records + trailer). |

## Notes

- Unknown record types must surface as a typed error so the caller can choose to skip or fail.
- For `Type = INT`, the decoder must validate that `Length == 8` and surface a typed error otherwise.
