# Record

Records are variable-length. Type and Length are fixed; Value is `Length` bytes.

**Byte order:** big-endian.

## Field table

| Offset (bytes) | Size     | Type        | Name   | Description                                 |
|----------------|----------|-------------|--------|---------------------------------------------|
| 0              | 1        | `RecordType`| Type   | One of the values in `../encoding-tables/record-type.md`. |
| 1              | 2        | `uint16`    | Length | Length of `Value` in bytes (0 ≤ Length ≤ 65535). |
| 3              | `Length` | `[]byte`    | Value  | Raw payload. Interpretation depends on `Type`. |

## Notes

- A record with `Length = 0` is legal and carries an empty `Value`.
- Unknown record types must surface as a typed error so the caller can choose to skip or fail.
- For `Type = NESTED`, the `Value` is itself a complete TLV1 file (header + records + trailer).
