# TLV1 Binary File Format

## Overview

TLV1 is a simple type-length-value container format. A file consists of an 8-byte header, zero or more variable-length records, and a 4-byte trailer:

```
+---------+-----------+-----------+
| Header  | Records   | Trailer   |
| 8 bytes | variable  | 4 bytes   |
+---------+-----------+-----------+
```

Records are read in order until the bytes preceding the trailer are exhausted.

## Conventions

- **Byte order**: big-endian for all multi-byte integers.
- **Sizes**: `Offset (bytes)` columns are zero-based and refer to the byte stream as a whole.
- **Bit numbering**: bit 0 is the least significant bit; bit 7 is the most significant.
- **Strings**: not null-terminated. Length is always carried by an explicit prefix.

## Field Definitions

### Header

The header is fixed-size: 8 bytes. It identifies the file and carries format-wide flags.

| Offset (bytes) | Size | Type      | Name     | Description                             |
|----------------|------|-----------|----------|-----------------------------------------|
| 0              | 4    | `[4]byte` | Magic    | ASCII `"TLV1"` (`0x54 0x4C 0x56 0x31`). |
| 4              | 1    | `uint8`   | Version  | Format version. Must be `1`.            |
| 5              | 1    | `uint8`   | Flags    | Bit field; see "Header.Flags bit field" below. |
| 6              | 2    | `uint16`  | Reserved | Reserved for future use. Must be `0`.   |

#### Header.Flags bit field

A single byte holding three independent boolean flags.

| Bit(s) | Mask  | Name       | Description                                  |
|--------|-------|------------|----------------------------------------------|
| 0      | 0x01  | COMPRESSED | Record values are zlib-compressed.           |
| 1      | 0x02  | ENCRYPTED  | Record values are AES-encrypted.             |
| 2      | 0x04  | SIGNED     | The trailer carries a signature in addition to the CRC. |
| 3-7    | 0xF8  | (reserved) | Must be 0.                                   |

### Record

Records are variable-length. Type and Length are fixed; Value is `Length` bytes.

| Offset (bytes) | Size     | Type      | Name   | Description                                 |
|----------------|----------|-----------|--------|---------------------------------------------|
| 0              | 1        | `uint8`   | Type   | One of the values in "Encoding Tables".     |
| 1              | 2        | `uint16`  | Length | Length of `Value` in bytes (0 ≤ Length ≤ 65535). |
| 3              | `Length` | `[]byte`  | Value  | Raw payload. Interpretation depends on `Type`. |

A record with `Length = 0` is legal and carries an empty `Value`.

### Trailer

The trailer is fixed-size: 4 bytes.

| Offset (bytes) | Size | Type     | Name  | Description                                   |
|----------------|------|----------|-------|-----------------------------------------------|
| 0              | 4    | `uint32` | CRC32 | IEEE CRC32 of all bytes preceding the trailer. |

## Encoding Tables

### Record.Type values

| Value | Name   | Meaning                                |
|-------|--------|----------------------------------------|
| 0x01  | STRING | UTF-8 string (no null terminator).     |
| 0x02  | INT    | Big-endian signed 64-bit integer (Length must be 8). |
| 0x03  | BLOB   | Opaque byte payload.                   |
| 0x04  | NESTED | Value is itself a TLV1 file (header + records + trailer). |

Unknown record types must surface as a typed error so the caller can choose to skip or fail.

## Conditional and Optional Fields

- A file with zero records is legal: header followed immediately by trailer.
- The Reserved field of the header must be zero on write; readers should fail with a typed error if Reserved ≠ 0.
- The SIGNED flag is reserved for a future signed-trailer extension; readers should fail if encountered (this version does not specify the signature layout).

## Checksums and Integrity

The trailer's CRC32 covers every byte from offset 0 (the start of the header) up to (but not including) the CRC32 itself. Decoders must compute the running CRC32 as they read, then compare it to the value in the trailer; on mismatch, return a typed error wrapping the leaf sentinel.

## Padding and Alignment

There is no padding between header and records, between records, or between the last record and the trailer. Implementations must not assume natural alignment of multi-byte fields.

## Versioning

Only Version 1 is defined. Future revisions will increment `Header.Version`. A decoder must reject files with an unrecognized version using a typed error.

## Examples

### Minimal: header + trailer (no records)

```
54 4C 56 31    Magic = "TLV1"
01             Version = 1
00             Flags = 0
00 00          Reserved = 0
A1 B2 C3 D4    CRC32 of the preceding 8 bytes (illustrative; recompute when wiring up tests)
```

### Typical: one STRING record, no flags

Header (8) + one record with Type=STRING, Length=5, Value="hello" + trailer (4) = 20 bytes total.

```
54 4C 56 31 01 00 00 00          Header
01 00 05 68 65 6C 6C 6F          Record: Type=0x01, Length=5, Value="hello"
.. .. .. ..                       CRC32 (computed over the 16 preceding bytes)
```

### Complex: COMPRESSED flag set, two records

```
54 4C 56 31 01 01 00 00          Header (Flags = COMPRESSED)
01 00 03 78 9C 03                Record 1: Type=STRING, Length=3, Value=zlib-compressed bytes
02 00 08 00 00 00 00 00 00 00 2A Record 2: Type=INT, Length=8, Value=42 (big-endian int64)
.. .. .. ..                       CRC32
```
