# TLVX Binary File Format

## Overview

TLVX is an extended type-length-value container format. It builds on the simpler TLV1 format (header / records / trailer) by adding an index section, per-record framing metadata, multiple checksum algorithms, an extension table, and a richer record type enumeration. A TLVX file consists of a fixed-size header, an optional index, a variable-length sequence of records, an optional extension table, and a fixed-size trailer:

```
+---------+---------+---------+---------+---------+
| Header  | Index   | Records | ExtTab  | Trailer |
| 16 byte | optional| variable| optional| 16 byte |
+---------+---------+---------+---------+---------+
```

The header carries a magic, a version, a flag byte, an index-record count, and the offset (relative to start-of-file) of the trailer. The optional index is a packed table of `(record_id, offset)` pairs that lets readers jump directly to a record without scanning. Records are read sequentially after the index (or after the header, if no index) until the byte preceding the optional extension table is reached. The extension table holds key-value pairs of arbitrary metadata (creation timestamp, application version, custom tags) for forward-compatibility. The trailer carries two checksums — one over the body bytes, one over the trailer's own preceding bytes — plus an end-of-file magic.

TLVX is consciously designed to be readable from a stream and to allow random access via the index. Writers may choose to omit the index (setting `IndexCount = 0`) to keep small files compact; readers must handle both forms.

## Conventions

- **Byte order**: big-endian for all multi-byte integers in headers, lengths, offsets, and checksums. Record values are byte-typed (string, int, blob, nested) and follow the type-specific encoding.
- **Sizes**: `Offset (bytes)` columns are zero-based and refer to the byte stream as a whole.
- **Bit numbering**: bit 0 is the least significant bit; bit 7 is the most significant.
- **Strings**: not null-terminated. Length is always carried by an explicit prefix.
- **Reserved bytes**: every reserved field must be `0` on write; readers must reject non-zero reserved fields with a typed error.
- **Lengths and offsets**: 32-bit unsigned big-endian unless otherwise stated. The maximum file size is therefore 4 GiB - 1.
- **Negative integers**: where signed integers appear, they are two's complement big-endian.
- **Floats**: where IEEE 754 floats appear, they are big-endian (network order). TLVX uses 64-bit doubles in floating-point fields; 32-bit floats are not part of the format.
- **Checksums**: TLVX defines three algorithms identified by single-byte tags in the trailer (`0x01` CRC32-IEEE, `0x02` CRC64-ECMA, `0x03` SHA256-truncated-32). The body and trailer checksums must use the same algorithm; mixing is not legal.

## Field Definitions

### Header

The header is fixed-size: 16 bytes. It identifies the file, carries format-wide flags, and points at the trailer.

| Offset (bytes) | Size | Type      | Name           | Description                                                                                                  |
|----------------|------|-----------|----------------|--------------------------------------------------------------------------------------------------------------|
| 0              | 4    | `[4]byte` | Magic          | ASCII `"TLVX"` (`0x54 0x4C 0x56 0x58`).                                                                      |
| 4              | 1    | `uint8`   | Version        | Format version. Must be `1`.                                                                                  |
| 5              | 1    | `uint8`   | Flags          | Bit field; see "Header.Flags bit field" below.                                                                |
| 6              | 1    | `uint8`   | ChecksumAlg    | Checksum algorithm tag for the body and trailer checksums (see "Checksum algorithms" in Encoding Tables).    |
| 7              | 1    | `uint8`   | Reserved1      | Reserved for future use. Must be `0`.                                                                          |
| 8              | 2    | `uint16`  | IndexCount     | Number of `(record_id, offset)` entries in the optional Index. Zero means "no Index section".                |
| 10             | 2    | `uint16`  | ExtCount       | Number of `(tag, length, value)` entries in the optional Extension Table. Zero means "no extension table".   |
| 12             | 4    | `uint32`  | TrailerOffset  | Byte offset (from start-of-file) of the first byte of the trailer.                                            |

The header occupies exactly bytes `[0, 16)` of every TLVX file. A reader that parses these 16 bytes can compute exactly where the index, records, extension table, and trailer begin.

#### Header.Flags bit field

A single byte holding eight independent boolean flags.

| Bit(s) | Mask  | Name        | Description                                                                  |
|--------|-------|-------------|------------------------------------------------------------------------------|
| 0      | 0x01  | COMPRESSED  | Record values are zlib-compressed (after the per-record framing).            |
| 1      | 0x02  | ENCRYPTED   | Record values are AES-encrypted (after compression, if both flags are set). |
| 2      | 0x04  | SIGNED      | The trailer carries a 64-byte signature appended after the checksums.       |
| 3      | 0x08  | INDEXED     | Convenience flag: must be set iff `IndexCount > 0` (writers and readers must enforce). |
| 4      | 0x10  | EXTENDED    | Convenience flag: must be set iff `ExtCount > 0`.                            |
| 5      | 0x20  | STRICT      | If set, unknown record types are a hard error (rather than skip-and-continue). |
| 6      | 0x40  | SEALED      | If set, no further records may be appended to the file.                      |
| 7      | 0x80  | (reserved)  | Must be 0.                                                                    |

### Index

The Index is an optional fixed-stride table appearing immediately after the header (offset `[16, 16 + 8*IndexCount)`). Each entry is 8 bytes:

| Offset (bytes from start of entry) | Size | Type     | Name     | Description                                                |
|------------------------------------|------|----------|----------|------------------------------------------------------------|
| 0                                  | 4    | `uint32` | RecordID | Caller-assigned id; readers may use this to look up records. |
| 4                                  | 4    | `uint32` | Offset   | Byte offset (from start-of-file) of the indexed record.    |

Index entries are stored in ascending order by `Offset`. `RecordID` values need not be unique, and need not be ordered. A writer that wishes to avoid the index entirely sets `IndexCount = 0` (and clears the `INDEXED` flag).

### Record

Records are variable-length and appear immediately after the index (or after the header, if no index). Each record carries a fixed-size frame header followed by `Length` bytes of value.

| Offset (bytes from start of record) | Size      | Type     | Name      | Description                                                     |
|-------------------------------------|-----------|----------|-----------|-----------------------------------------------------------------|
| 0                                   | 1         | `uint8`  | Type      | Record type tag; one of the values in "Encoding Tables".         |
| 1                                   | 1         | `uint8`  | Subtype   | Type-specific subtype (zero for types without a subtype).        |
| 2                                   | 2         | `uint16` | Flags     | Per-record framing flags (see "Record.Flags bit field" below).   |
| 4                                   | 4         | `uint32` | Length    | Length of `Value` in bytes (0 ≤ Length ≤ 2^32 - 1).              |
| 8                                   | `Length`  | `[]byte` | Value     | Raw payload. Interpretation depends on `Type` and `Subtype`.     |

A record with `Length = 0` is legal and carries an empty `Value`. A record with `Length > 0` must have `Length` bytes of value content following the 8-byte frame header.

#### Record.Flags bit field

A 16-bit field holding per-record framing flags.

| Bit(s) | Mask    | Name             | Description                                                                  |
|--------|---------|------------------|------------------------------------------------------------------------------|
| 0      | 0x0001  | COMPRESSED       | This record's value is zlib-compressed (overrides Header.Flags COMPRESSED).  |
| 1      | 0x0002  | ENCRYPTED        | This record's value is AES-encrypted.                                        |
| 2      | 0x0004  | DEPRECATED       | Soft hint that callers should treat this record as deprecated; not enforced. |
| 3      | 0x0008  | INDEXED          | This record appears in the Index. Set by writers, validated by readers.      |
| 4-7    | 0x00F0  | Priority         | Four-bit priority (0-15) for caller use.                                     |
| 8      | 0x0100  | NESTED_HAS_INDEX | For NESTED records: nested file has its own Index.                           |
| 9      | 0x0200  | SCHEMA           | Value carries a schema descriptor; consumers may use it to validate Subtype. |
| 10     | 0x0400  | TIMESTAMPED      | First 8 bytes of Value are a big-endian Unix-millisecond timestamp.          |
| 11     | 0x0800  | CHUNKED          | Value is a chunked stream (see Conditional/Optional Fields below).           |
| 12-15  | 0xF000  | (reserved)       | Must be 0.                                                                   |

### Extension Table

The Extension Table is an optional sequence of `(tag, length, value)` triples appearing immediately after the records (offset = TrailerOffset - sum of extension entry sizes). Each entry is variable-length:

| Offset (bytes from start of entry) | Size      | Type     | Name    | Description                                          |
|------------------------------------|-----------|----------|---------|------------------------------------------------------|
| 0                                  | 2         | `uint16` | Tag     | Extension tag (see "Extension tags" in Encoding Tables). |
| 2                                  | 2         | `uint16` | Length  | Length of `Value` in bytes.                          |
| 4                                  | `Length`  | `[]byte` | Value   | Tag-specific payload.                                |

Extension entries appear in tag-ascending order. Duplicate tags are not legal; a reader must surface them with `DuplicateExtensionError{Tag, Offset}`.

A writer that wishes to omit the extension table sets `ExtCount = 0` (and clears the `EXTENDED` flag in `Header.Flags`).

### Section Descriptor

A SectionDescriptor is an optional auxiliary structure that may appear inside a record's value when the SCHEMA flag is set in `Record.Flags`. It describes the structural shape of the record's payload so consumers can validate or transform it without prior knowledge.

| Offset (bytes from start of descriptor) | Size      | Type     | Name        | Description                                                  |
|-----------------------------------------|-----------|----------|-------------|--------------------------------------------------------------|
| 0                                       | 2         | `uint16` | DescriptorMagic | Must be `0xD1D1`.                                          |
| 2                                       | 1         | `uint8`  | DescriptorVersion | Must be `1`.                                             |
| 3                                       | 1         | `uint8`  | FieldCount  | Number of `(name, type)` field entries that follow.          |
| 4                                       | variable  | `entry[]`| Fields      | Field-name + field-type pairs (see below).                   |
| ...                                     | 4         | `uint32` | DescriptorLength | Total byte length of the descriptor (for skip-over readers). |

Each field entry is encoded as:

| Offset (bytes from start of entry) | Size     | Type     | Name      | Description                                                  |
|------------------------------------|----------|----------|-----------|--------------------------------------------------------------|
| 0                                  | 1        | `uint8`  | NameLen   | Length of the field name in bytes (1-255).                   |
| 1                                  | NameLen  | `[]byte` | Name      | UTF-8 field name.                                            |
| 1+NameLen                          | 1        | `uint8`  | FieldType | Tag from "FieldType values" in Encoding Tables.              |
| 2+NameLen                          | 1        | `uint8`  | Subtype   | Type-specific subtype.                                       |

A SectionDescriptor whose `DescriptorMagic` is not `0xD1D1` must be rejected with `InvalidDescriptorError{ExpectedMagic: 0xD1D1, Got: <value>}`. A descriptor with `DescriptorVersion != 1` must be rejected with `UnknownDescriptorVersionError{Version}`.

The remaining bytes of the record's `Value` (after the descriptor) are the actual payload, interpreted according to the descriptor.

### Chunked Stream Frame

When a record's `Flags` field has the CHUNKED bit set, the record's `Value` is a sequence of length-prefixed chunks rather than a single contiguous payload. Each chunk is encoded as:

| Offset (bytes from start of chunk) | Size     | Type     | Name      | Description                                                  |
|------------------------------------|----------|----------|-----------|--------------------------------------------------------------|
| 0                                  | 4        | `uint32` | ChunkLen  | Length of this chunk's data in bytes; `0` is the terminator. |
| 4                                  | ChunkLen | `[]byte` | Data      | Chunk data.                                                  |

A chunked record is read by repeatedly reading `ChunkLen` and `Data` until `ChunkLen = 0`. The terminator's 4 zero bytes count toward the outer record's `Length`. A chunked record whose terminator is missing (the outer record's `Length` runs out before a `ChunkLen = 0` is seen) is rejected with `MissingChunkTerminatorError{Field: "Value"}`.

### BIGINT and DECIMAL encoding

The BIGINT record type (Record.Type = 0x0C) encodes an arbitrary-precision two's complement signed integer. The bytes are stored big-endian, most significant byte first. The sign is determined by the high bit of the first byte: `0` is non-negative, `1` is negative. A BIGINT with `Length = 0` is the zero integer (the canonical encoding of zero).

A BIGINT decoder must:

1. Read all `Length` bytes as a big-endian byte sequence.
2. Interpret as two's complement: if `Length > 0` and the high bit of the first byte is `1`, the value is negative.
3. Return the value as a `*big.Int` (Go) or equivalent arbitrary-precision integer in other languages.

Encoders must canonicalise: leading zero bytes that are not needed (i.e. removing them would not change the sign or magnitude) must be stripped before writing. A non-canonical BIGINT is rejected by the decoder with `NonCanonicalBigIntError{LeadingZeros}` only when STRICT mode is set; otherwise it is accepted and silently canonicalised on re-encode.

The DECIMAL record type (Record.Type = 0x0D) encodes a decimal value as `(scale, coefficient)` where:

| Offset (bytes from start of value) | Size      | Type     | Name        | Description                                                  |
|------------------------------------|-----------|----------|-------------|--------------------------------------------------------------|
| 0                                  | 1         | `int8`   | Scale       | Signed scale in the range -128 to 127.                       |
| 1                                  | Length-1  | `[]byte` | Coefficient | BIGINT-encoded coefficient.                                  |

The decimal value is `coefficient * 10^(-scale)`. So `(scale=2, coefficient=12345)` is `123.45`; `(scale=-3, coefficient=42)` is `42000`. A DECIMAL record with `Length = 0` is rejected with `EmptyDecimalError`. A DECIMAL with `Length = 1` (scale only, no coefficient) encodes the zero decimal, regardless of the scale value.

### Header Extension Block

The Header Extension Block is an optional fixed-stride block that may appear immediately after the trailer when the EXTENDED-HEADER bit is set in a future Header.Flags revision (currently reserved). The current spec does not define a layout for this block; readers must reject any file whose trailer is followed by additional bytes (other than the optional signature) with `UnexpectedTrailingBytesError{Offset, Length}`.

### Trailer

The trailer is fixed-size: 16 bytes (without the optional 64-byte signature) or 80 bytes (with the signature, when the SIGNED flag is set).

| Offset (bytes from start of trailer) | Size | Type      | Name        | Description                                                              |
|--------------------------------------|------|-----------|-------------|--------------------------------------------------------------------------|
| 0                                    | 8    | `uint64`  | BodyChecksum | Checksum (algorithm per Header.ChecksumAlg) of every byte from offset 0 up to (but not including) the trailer. |
| 8                                    | 1    | `uint8`   | TrailerVersion | Trailer format version. Must be `1`.                                  |
| 9                                    | 1    | `uint8`   | Reserved2   | Reserved for future use. Must be `0`.                                    |
| 10                                   | 2    | `uint16`  | RecordCount | Total number of records in the file (excluding extension entries).      |
| 12                                   | 4    | `[4]byte` | EofMagic    | ASCII `"XEND"` (`0x58 0x45 0x4E 0x44`).                                  |

If the SIGNED flag is set in `Header.Flags`, an additional 64-byte signature follows the trailer:

| Offset (bytes from start of trailer + 16) | Size | Type      | Name      | Description                                  |
|-------------------------------------------|------|-----------|-----------|----------------------------------------------|
| 0                                         | 64   | `[64]byte`| Signature | Caller-defined signature over the body bytes. |

The decoder verifies the signature when SIGNED is set; the encoder computes and writes it. The format of the signature itself (which key, which algorithm) is out of scope for TLVX — it's a pass-through field.

## Encoding Tables

### Record.Type values

| Value | Name      | Meaning                                                                       |
|-------|-----------|-------------------------------------------------------------------------------|
| 0x01  | STRING    | UTF-8 string (no null terminator).                                            |
| 0x02  | INT       | Big-endian signed 64-bit integer (Length must be 8).                          |
| 0x03  | BLOB      | Opaque byte payload.                                                          |
| 0x04  | NESTED    | Value is itself a TLVX file (header + records + trailer).                     |
| 0x05  | FLOAT     | Big-endian IEEE 754 64-bit double (Length must be 8).                         |
| 0x06  | TIMESTAMP | Big-endian Unix-millisecond timestamp (Length must be 8).                     |
| 0x07  | UUID      | RFC 4122 UUID (Length must be 16).                                            |
| 0x08  | LIST      | Sequence of records-without-frame: see "LIST encoding" below.                 |
| 0x09  | MAP       | Sequence of `(key_record, value_record)` pairs: see "MAP encoding" below.     |
| 0x0A  | REFERENCE | 4-byte uint32 record id, pointing at another record in the same file.         |
| 0x0B  | NULL      | Length must be 0; carries the format-defined null value.                      |
| 0x0C  | BIGINT    | Variable-length two's complement big-endian integer; Length is the byte count. |
| 0x0D  | DECIMAL   | Big-endian signed scale (1 byte) + variable-length two's complement coefficient. |
| 0x0E  | SYMBOL    | Caller-defined small symbol; Subtype field disambiguates flavours.            |
| 0x0F  | RESERVED  | Reserved. Readers must surface with `UnknownRecordTypeError`.                 |

Unknown record types must surface as a typed error so the caller can choose to skip or fail. The STRICT flag in `Header.Flags` controls the behaviour: when set, the decoder fails immediately on `UnknownRecordTypeError`; when unset, the decoder skips over the record (using its `Length`) and continues.

### Record.Subtype values

Subtype values are namespaced to their record type. The defined subtypes are:

| Type     | Subtype | Name        | Meaning                                            |
|----------|---------|-------------|----------------------------------------------------|
| STRING   | 0x00    | UTF8        | UTF-8 (the default).                               |
| STRING   | 0x01    | ASCII7      | 7-bit ASCII, validated.                            |
| STRING   | 0x02    | LATIN1      | ISO-8859-1.                                        |
| STRING   | 0x03    | UTF16BE     | UTF-16 big-endian.                                 |
| INT      | 0x00    | I64         | 64-bit signed (the default).                       |
| INT      | 0x01    | I32         | 32-bit signed (Length must be 4).                  |
| INT      | 0x02    | U64         | 64-bit unsigned (semantic only — bytes identical). |
| BLOB     | 0x00    | OPAQUE      | Opaque (the default).                              |
| BLOB     | 0x01    | DICT_REF    | First 4 bytes are a dictionary reference id.       |
| FLOAT    | 0x00    | F64         | 64-bit double (the default).                       |
| TIMESTAMP| 0x00    | UNIX_MS     | Unix milliseconds (the default).                   |
| TIMESTAMP| 0x01    | UNIX_NS     | Unix nanoseconds (Length must be 8).               |
| TIMESTAMP| 0x02    | RFC3339     | UTF-8 RFC3339 string (variable Length).            |
| LIST     | 0x00    | HOMOGENEOUS | All elements share Type and Subtype.               |
| LIST     | 0x01    | HETEROGENEOUS | Elements may differ.                             |
| MAP      | 0x00    | UNORDERED   | Entries in arbitrary order.                        |
| MAP      | 0x01    | ORDERED     | Entries in key-sort order.                         |

A subtype value not listed above is reserved; the decoder must surface it with `UnknownRecordSubtypeError{Type, Subtype, Offset}` regardless of the STRICT flag.

### Checksum algorithms

| Tag   | Name       | Length (bytes) | Polynomial / digest                              |
|-------|------------|----------------|--------------------------------------------------|
| 0x01  | CRC32_IEEE | 4 (right-aligned in the 8-byte BodyChecksum field) | x^32 + x^26 + x^23 + … (standard IEEE 802.3) |
| 0x02  | CRC64_ECMA | 8                                                | x^64 + x^62 + x^57 + … (ECMA-182)            |
| 0x03  | SHA256_T32 | 32 (truncated to first 8 bytes for BodyChecksum) | SHA-256, leftmost 8 bytes                    |
| 0x04  | XXH64      | 8                                                | xxHash 64-bit, seed 0                        |
| 0x05  | BLAKE3_T32 | 32 (truncated to first 8 bytes for BodyChecksum) | BLAKE3, leftmost 8 bytes                     |

A `ChecksumAlg` value not listed above is reserved; readers must reject the file with `UnknownChecksumAlgError{Alg}`.

### Compression algorithms

When the COMPRESSED flag is set (in Header.Flags or per-record Record.Flags), the record's value is compressed. The compression algorithm is determined by the high bits of `Record.Flags` reserved for that purpose (currently always zlib in version 1). Future versions will allocate distinct bits or a new field for algorithm selection.

| Tag   | Name        | Notes                                                              |
|-------|-------------|--------------------------------------------------------------------|
| 0x01  | ZLIB        | RFC 1950, default for the COMPRESSED flag in version 1.            |
| 0x02  | ZSTD        | Zstandard, reserved for future versions.                            |
| 0x03  | LZ4         | LZ4 frame format, reserved for future versions.                     |
| 0x04  | SNAPPY      | Snappy framing format, reserved for future versions.                |
| 0x05  | BROTLI      | Brotli, reserved for future versions.                               |

A reader that encounters a compression algorithm not listed above must surface `UnknownCompressionAlgError{Alg}`.

### Encryption algorithms

When the ENCRYPTED flag is set, the record's value is encrypted. The encryption algorithm is determined by the high bits of `Record.Flags` reserved for that purpose (currently always AES-256-GCM in version 1).

| Tag   | Name             | Notes                                                              |
|-------|------------------|--------------------------------------------------------------------|
| 0x01  | AES_256_GCM      | AES-256 in GCM mode, default for the ENCRYPTED flag in version 1.  |
| 0x02  | AES_128_GCM      | AES-128 in GCM mode, reserved for future versions.                  |
| 0x03  | CHACHA20_POLY1305| ChaCha20-Poly1305, reserved for future versions.                    |
| 0x04  | XCHACHA20        | XChaCha20-Poly1305, reserved for future versions.                   |

A reader that encounters an encryption algorithm not listed above must surface `UnknownEncryptionAlgError{Alg}`.

### FieldType values

Used in SectionDescriptor entries to describe the type of each named field.

| Tag   | Name      | Underlying TLVX representation                  |
|-------|-----------|-------------------------------------------------|
| 0x01  | STRING    | UTF-8 string with leading uint32 length         |
| 0x02  | INT       | Big-endian signed 64-bit integer                |
| 0x03  | UINT      | Big-endian unsigned 64-bit integer              |
| 0x04  | FLOAT     | Big-endian IEEE 754 double                      |
| 0x05  | BOOL      | Single byte (`0x00` false, `0x01` true)         |
| 0x06  | BYTES     | Variable-length blob with leading uint32 length |
| 0x07  | TIMESTAMP | Big-endian Unix milliseconds (uint64)           |
| 0x08  | UUID      | 16-byte RFC 4122 UUID                           |
| 0x09  | NESTED    | Variable-length nested SectionDescriptor + payload |
| 0x0A  | LIST      | uint32 count + LIST element body                |
| 0x0B  | MAP       | uint32 count + key/value entry pairs            |
| 0x0C  | NULL      | Zero bytes; presence bit elsewhere indicates set/unset |

A descriptor whose entry has a `FieldType` not listed above must be rejected with `UnknownFieldTypeError{FieldType, Subtype, Offset}`.

### Symbol kinds

Used in the `Subtype` field of records of `Type = SYMBOL` (0x0E).

| Subtype | Name           | Meaning                                                |
|---------|----------------|--------------------------------------------------------|
| 0x00    | INTERNED       | Caller-defined symbol, interned by id.                 |
| 0x01    | EXTERNAL       | Caller-defined symbol, looked up in an external table. |
| 0x02    | KEYWORD        | Reserved-keyword symbol (caller-defined keyword set).  |
| 0x03    | NAMESPACE      | Namespaced symbol (`Value` is `<namespace>:<name>`).   |

A symbol record's `Subtype` not listed above must be rejected with `UnknownSymbolKindError{Subtype}`.

### Extension tags

| Tag   | Name              | Value format                                     |
|-------|-------------------|--------------------------------------------------|
| 0x0001 | CREATED_AT_UNIX_MS | uint64 big-endian Unix milliseconds            |
| 0x0002 | WRITER_ID         | UTF-8 application identifier                     |
| 0x0003 | WRITER_VERSION    | UTF-8 semver string                              |
| 0x0004 | CHUNK_BUFFER_SIZE | uint32 big-endian; suggested chunk-buffer size for streaming readers |
| 0x0005 | DICTIONARY        | Caller-defined dictionary blob                   |
| 0x0006 | SCHEMA_URI        | UTF-8 URI                                        |
| 0x0007 | DESCRIPTION       | UTF-8 description string                         |
| 0x0008 | TAGS              | UTF-8 newline-separated list of caller tags      |

A tag value not listed above is reserved; the decoder must skip over the extension entry without surfacing an error (extensions are explicitly forward-compatible — unknown tags are not failures).

## Conditional and Optional Fields

### Empty files

A file with zero records is legal: header followed immediately (or after an empty index) by the trailer. `RecordCount = 0` in the trailer. An empty file's `BodyChecksum` is computed over the bytes from offset 0 up to the start of the trailer — for a no-record file with no index, that's exactly 16 bytes (the header). For a no-record file with an index but `IndexCount = 0`, it's still 16 bytes (the index occupies zero bytes when `IndexCount = 0`).

### Reserved fields

The `Reserved1` and `Reserved2` fields in the header and trailer must be zero on write; readers must fail with `ReservedFieldNonZeroError{Field, Got}` if non-zero. Reserved bits in `Record.Flags` (bits 12-15) and `Header.Flags` (bit 7) are similarly enforced; an `Record.Flags` value where any reserved bit is set is rejected with `ReservedFlagBitError{Field, Bit}`.

### SIGNED flag

The SIGNED flag (bit 2 in `Header.Flags`) controls whether a 64-byte signature follows the trailer. Readers must verify the signature when SIGNED is set; encoders must compute and append it. The format of the signature itself is out of scope for TLVX (hash algorithm and key are caller-provided). The signature bytes are computed over every byte from offset 0 up to (but not including) the signature itself — that is, including the BodyChecksum, TrailerVersion, RecordCount, and EofMagic fields.

A SIGNED file whose signature does not verify must surface `ErrSignatureMismatch` wrapped through `wrapErr`. The decoder must report this error before the BodyChecksum mismatch — signature verification supersedes checksum verification because a forged file with a recomputed checksum would otherwise pass.

### INDEXED and EXTENDED consistency

The INDEXED flag must be set iff `IndexCount > 0`. A reader that sees INDEXED set with `IndexCount = 0` (or vice versa) must reject the file with `InconsistentFlagError{Flag: "INDEXED", Reason: "IndexCount mismatch"}`.

The EXTENDED flag must be set iff `ExtCount > 0`. Same enforcement as INDEXED, with `Flag: "EXTENDED"`.

### CHUNKED records

The CHUNKED flag (in Record.Flags) marks a record whose value is a sequence of `(uint32 length, value bytes)` chunks terminated by a chunk with `length = 0`. Chunked records' total length is the sum of their chunk sizes plus the per-chunk length prefixes plus 4 bytes for the terminator; the outer `Length` field still describes the on-disk byte count.

A chunked record cannot be both CHUNKED and COMPRESSED at the per-record level — the spec rejects the combination with `InvalidFlagComboError{Combo: "CHUNKED+COMPRESSED"}`. The reasoning: CHUNKED is for streaming reads of large records; COMPRESSED is for compact storage. A reader cannot stream-decompress without knowing the total compressed length up front, which CHUNKED hides. (The combination is fine at the file level, where the whole file is compressed before chunking.)

### TIMESTAMPED records

The TIMESTAMPED flag (in Record.Flags) marks a record whose first 8 bytes of Value are a big-endian Unix-millisecond timestamp. The remaining `Length - 8` bytes are the actual value. A TIMESTAMPED record with `Length < 8` is rejected with `InsufficientLengthError{Field: "Value", Need: 8, Got: Length}`.

The TIMESTAMPED flag is independent of the TIMESTAMP record type (Type = 0x06): a TIMESTAMP-typed record has its own value-is-a-timestamp semantics, so the flag is **redundant** there. A TIMESTAMP record with TIMESTAMPED set is rejected with `RedundantFlagError{Type: "TIMESTAMP", Flag: "TIMESTAMPED"}`. The flag is meaningful for non-timestamp record types where the value happens to carry a creation/update time alongside the actual data.

### LIST encoding

The LIST encoding (Record.Type = 0x08) is: a 4-byte big-endian count, followed by that many child records (each with full frame header). The outer record's `Length` field is the byte count of the count + child records.

A LIST record with `Subtype = HOMOGENEOUS` requires all child records to share `Type` and `Subtype`; mismatch is reported as `HeterogeneousListError{ExpectedType, ExpectedSubtype, Got, Offset}`. A LIST record with `Subtype = HETEROGENEOUS` skips this check.

### MAP encoding

The MAP encoding (Record.Type = 0x09) is: a 4-byte big-endian count, followed by that many `(key_record, value_record)` pairs. The outer record's `Length` is the byte count of the entire encoding.

A MAP record with `Subtype = ORDERED` requires keys to be in sort order (lexicographic byte comparison); out-of-order keys are reported as `UnsortedMapError{Index, Offset}`. A MAP record with `Subtype = UNORDERED` skips this check.

Duplicate keys in a MAP record are not legal regardless of subtype; the decoder must surface `DuplicateMapKeyError{Index, Offset}` for the first duplicate.

### REFERENCE encoding

The REFERENCE encoding (Record.Type = 0x0A) is a 4-byte big-endian record id pointing at another record in the same file. The decoder does not eagerly resolve references — the AST stores the id and a separate `Resolve` API walks the index. A REFERENCE record whose id does not appear in the file's index is reported lazily as `UnresolvedReferenceError{ID}` when the caller invokes `Resolve`.

A reference cycle is detected by `Resolve`: if following references from id A eventually reaches id A again without intermediate resolution, the error is `ReferenceCycleError{Stack []uint32}`. The detection is best-effort — the resolver maintains a small visited set bounded at 64 entries; cycles longer than 64 hops are not detected and instead surface as a stack-depth failure (`ReferenceDepthExceededError{Limit: 64}`).

### NESTED record validation

The NESTED record encoding (Record.Type = 0x04) wraps a complete TLVX file (header + records + trailer) inside the value bytes. The nested file's `Header.Magic` must be `"TLVX"`; if it isn't, the decoder surfaces `InvalidNestedFileError{Reason: "magic"}`.

The nested file's `Header.Version` must match the outer file's version — mixing versions in nesting is not legal in version 1. The decoder surfaces `MixedVersionsError{Outer, Inner}` on mismatch.

The nested file's `Header.ChecksumAlg` may differ from the outer file's — each nested file is independently checksummed.

### SEALED files

The SEALED flag (bit 6 in `Header.Flags`) marks a file as immutable — no further records may be appended. The flag is purely informational at the format level; the decoder accepts SEALED files identically to unsealed ones. The flag is intended for tools that perform append operations on TLVX files; those tools must refuse to append to a file with SEALED set.

A writer that wishes to convert an unsealed file to a sealed one must rewrite the entire file (re-checksum, re-sign if SIGNED is set). Toggling the SEALED bit alone would invalidate the BodyChecksum.

### STRICT vs. lenient decoding

The STRICT flag (bit 5 in `Header.Flags`) controls how the decoder responds to recoverable issues:

- Unknown record types: STRICT → fail with `UnknownRecordTypeError`; lenient → skip the record (using its `Length`) and continue.
- Unknown extension tags: STRICT → fail with `UnknownExtensionTagError`; lenient → skip the entry.
- Non-canonical BIGINT/DECIMAL: STRICT → fail with `NonCanonicalBigIntError`; lenient → accept and silently re-canonicalise on encode.
- Reserved-flag-bits set: STRICT and lenient — both reject. Reserved bits are not "unknown" — they are explicitly defined as zero.

The STRICT flag does **not** affect signature, checksum, version, or magic verification — those are always strict regardless of the flag.

### Maximum sizes and overflow

A TLVX file's maximum size is 4 GiB - 1 (constrained by the 32-bit `TrailerOffset` field). Writers that produce files approaching this limit must split into multiple files; the format does not define a multi-file extension.

A record's maximum value length is 4 GiB - 1 (constrained by the 32-bit `Length` field). Records approaching this limit are unusual and likely indicate the caller should be using NESTED or LIST encodings.

The Index supports up to 65535 entries (`uint16 IndexCount`). Files needing more index entries must structure their records into NESTED groupings, or the writer must omit the Index entirely (linear scan).

The Extension Table supports up to 65535 entries (`uint16 ExtCount`).

### NULL records

A NULL record (Record.Type = 0x0B) must have `Length = 0`. A NULL record with `Length != 0` is rejected with `NullRecordHasValueError{Length}`. The Subtype field for NULL records must also be `0`; nonzero Subtype is rejected with `NullRecordHasSubtypeError{Subtype}`.

## Checksums and Integrity

The trailer's `BodyChecksum` covers every byte from offset 0 (the start of the header) up to (but not including) the BodyChecksum field itself. Decoders must compute the running checksum as they read, then compare it to the value in the trailer; on mismatch, return a typed error wrapping the leaf sentinel `ErrChecksumMismatch`.

For algorithms whose digest is shorter than 8 bytes (CRC32_IEEE), the digest is right-aligned in the 8-byte field with the high bytes zero. For algorithms whose digest is longer than 8 bytes (SHA256, BLAKE3), the digest is truncated to its leftmost 8 bytes.

The `RecordCount` field in the trailer is informational; it must be consistent with the number of records actually present in the body. A reader that finds a mismatch must surface it with `RecordCountMismatchError{Header: <count>, Actual: <count>}`.

### Body checksum input bytes

The "body bytes" used as input to the checksum cover every byte from offset 0 (the start of the header's `Magic` field) up to (but not including) the trailer's `BodyChecksum` field. This means:

- The 16-byte header is included in the body.
- The Index (if present) is included in the body.
- The records are included.
- The Extension Table (if present) is included.
- The trailer is **not** included in the body checksum (the checksum cannot include itself).

A decoder that miscomputes the body byte range — e.g. by including the trailer or by stopping at the trailer's start one byte too early — will produce a checksum mismatch on otherwise-valid files. Tests should construct a known-good file with a hand-computed checksum and assert byte-for-byte equality before relying on round-trip equality.

### Streaming verification

For files large enough that holding the entire body in memory is impractical, decoders should compute the running checksum incrementally as bytes are consumed. The standard library exposes `tlvx.NewStreamingDecoder(r io.Reader) *StreamingDecoder` which yields records one at a time and surfaces a `Verify()` call at the end to compare against the trailer's BodyChecksum.

Streaming decoders must still validate the trailer's `EofMagic` and `TrailerVersion` before returning success — a streaming reader that returns a fully-read file but skips trailer validation is incorrect.

### Algorithm pinning per file

The `ChecksumAlg` field is per-file, not per-record. A writer that wishes to use different algorithms for different records must split the records across multiple files. This is intentional — per-record algorithm selection would force the trailer to carry a per-record table, which would defeat the trailer's fixed size.

The `ChecksumAlg` is read out of the header before any record bytes are consumed, so the decoder can choose the appropriate hash implementation up front. A header with `ChecksumAlg = 0` is reserved and rejected with `UnknownChecksumAlgError{Alg: 0}`.

## Padding and Alignment

There is no padding between header and index, between index and records, between records, between the last record and the extension table, or between the extension table and the trailer. Implementations must not assume natural alignment of multi-byte fields.

Decoders that read multi-byte fields by casting unaligned pointers (a common pattern in C/C++) must use `memcpy` or equivalent unaligned-load primitives rather than direct pointer dereference. Go implementations using `binary.BigEndian.Uint32` and friends are aligned-safe by construction; tests that read into byte slices and parse via the `binary` package are aligned-safe.

Encoders must write field bytes contiguously without padding bytes; an encoder that emits inadvertent padding (e.g. by mis-sizing a buffer) produces a file the decoder will reject because the offsets won't line up. The standard library's `tlvx.Encode` validates output by re-reading its own bytes through `tlvx.Decode` and asserting `require.Equal` on round-trip — implementations should follow the same self-test pattern.

## Versioning

Only Version 1 is defined. Future revisions will increment `Header.Version`. A decoder must reject files with an unrecognized version using a typed error `UnknownVersionError{Version}`. The trailer also has a `TrailerVersion` field for future-compatibility; a decoder that sees `TrailerVersion != 1` must reject with `UnknownTrailerVersionError{Version}`.

When a future TLVX version is defined, this spec will document which Version values continue to be accepted as a "compatibility window" — typically the previous and current versions. Readers that encounter an unsupported version must not silently fall back; they must return the typed error so the caller can choose recovery (e.g. fetch a newer reader, request the writer downgrade).

## Examples

### Minimal: header + trailer (no records, no index, no extensions)

```
54 4C 56 58 01 00 01 00          Header.Magic + Version + Flags=0 + ChecksumAlg=CRC32 + Reserved1=0
00 00 00 00 00 14 00 00          IndexCount=0 + ExtCount=0 + TrailerOffset=0x14 (=20 decimal)
A1 B2 C3 D4 00 00 00 00          BodyChecksum (right-aligned CRC32; high bytes zero)
01 00 00 00                      TrailerVersion=1 + Reserved2=0 + RecordCount=0
58 45 4E 44                      EofMagic = "XEND"
```

Total: 16 (header) + 16 (trailer) = 32 bytes.

### Typical: one STRING record, no flags

Header (16) + one record with Type=STRING, Subtype=UTF8, Length=5, Value="hello" + trailer (16) = 45 bytes total.

```
54 4C 56 58 01 00 01 00          Header
00 00 00 00 00 1D 00 00          IndexCount=0 + ExtCount=0 + TrailerOffset=0x1D (=29 decimal)
01 00 00 00 00 00 00 05          Record: Type=0x01 + Subtype=0x00 + Flags=0x0000 + Length=5
68 65 6C 6C 6F                   Value = "hello"
.. .. .. .. .. .. .. ..          BodyChecksum (computed over the 29 preceding bytes)
01 00 00 01                      TrailerVersion=1 + Reserved2=0 + RecordCount=1
58 45 4E 44                      EofMagic
```

### Complex: COMPRESSED flag set, two records, one INT and one TIMESTAMP

```
54 4C 56 58 01 01 02 00          Header (Flags = COMPRESSED, ChecksumAlg = CRC64)
00 00 00 00 00 30 00 00          IndexCount=0 + ExtCount=0 + TrailerOffset=0x30
02 00 00 00 00 00 00 08          Record 1: Type=INT, Subtype=I64, Flags=0, Length=8
00 00 00 00 00 00 00 2A          Value = 42 (big-endian int64)
06 00 00 00 00 00 00 08          Record 2: Type=TIMESTAMP, Subtype=UNIX_MS, Flags=0, Length=8
00 00 01 8B 4F 25 00 00          Value = 1700000000000 (Unix ms)
.. .. .. .. .. .. .. ..          BodyChecksum (CRC64 over the 48 preceding bytes)
01 00 00 02                      TrailerVersion=1 + Reserved2=0 + RecordCount=2
58 45 4E 44                      EofMagic
```

### With index: random-access ready

```
54 4C 56 58 01 08 01 00          Header (Flags = INDEXED, ChecksumAlg = CRC32)
00 02 00 00 00 30 00 00          IndexCount=2 + ExtCount=0 + TrailerOffset=0x30
00 00 00 01 00 00 00 20          Index entry 1: RecordID=1, Offset=0x20
00 00 00 02 00 00 00 28          Index entry 2: RecordID=2, Offset=0x28

01 00 00 08 00 00 00 05          Record 1 (offset 0x20): STRING, INDEXED, Length=5
68 65 6C 6C 6F 00 00 00          Value = "hello" + 3 bytes padding (illustrative — TLVX has no padding,
                                  this layout assumes lengths align)
```

### With extension table

```
54 4C 56 58 01 10 01 00          Header (Flags = EXTENDED)
00 00 00 02 00 30 00 00          IndexCount=0 + ExtCount=2
01 00 00 00 00 00 00 02          Record: Type=INT, Subtype=I64, Length=2... (truncated for example)

00 02 00 04 64 65 6D 6F          Ext entry: Tag=WRITER_ID + Length=4 + Value="demo"
00 03 00 05 31 2E 32 2E 33      Ext entry: Tag=WRITER_VERSION + Length=5 + Value="1.2.3"
```

The hex-dump layouts above are illustrative — actual byte counts depend on the record content. Implementers should compute lengths, offsets, and checksums dynamically when wiring up tests; the examples are oriented at communicating layout, not exact bytes.

### Nested record carrying a TLVX file

```
54 4C 56 58 01 00 01 00          Outer Header
00 00 00 00 00 32 00 00          IndexCount=0 + ExtCount=0 + TrailerOffset=0x32
04 00 00 00 00 00 00 1A          Outer record: Type=NESTED, Length=0x1A (=26 decimal)
54 4C 56 58 01 00 01 00          Inner Header (start of nested file)
00 00 00 00 00 12 00 00          Inner: IndexCount=0 + ExtCount=0 + TrailerOffset=0x12
.. .. .. .. .. .. .. ..          Inner BodyChecksum
01 00 00 00                      Inner: TrailerVersion + Reserved2 + RecordCount=0
58 45 4E 44                      Inner EofMagic
.. .. .. .. .. .. .. ..          Outer BodyChecksum
01 00 00 01                      Outer: TrailerVersion + Reserved2 + RecordCount=1
58 45 4E 44                      Outer EofMagic
```

The decoder reads the nested file by recursively invoking the decode pipeline on the value bytes — the inner header begins at offset 8 inside the outer record's value, and the inner trailer ends at the outer record's value boundary.

### LIST record with three INT children

```
54 4C 56 58 01 00 01 00          Header
00 00 00 00 00 30 00 00          IndexCount=0 + ExtCount=0 + TrailerOffset=0x30
08 00 00 00 00 00 00 24          Outer record: Type=LIST, Length=0x24 (=36 decimal)
00 00 00 03                      Child count = 3
02 00 00 00 00 00 00 08          Child 1: Type=INT, Subtype=I64, Length=8
00 00 00 00 00 00 00 01          Value = 1
02 00 00 00 00 00 00 08          Child 2: Type=INT, Subtype=I64, Length=8
00 00 00 00 00 00 00 02          Value = 2
02 00 00 00 00 00 00 08          Child 3: Type=INT, Subtype=I64, Length=8
00 00 00 00 00 00 00 03          Value = 3
.. .. .. .. .. .. .. ..          BodyChecksum
01 00 00 01                      Trailer
58 45 4E 44                      EofMagic
```

A HOMOGENEOUS LIST (Subtype = 0x00) requires every child's `(Type, Subtype)` to match the first child's; the decoder snapshots `(Type, Subtype)` after reading the first child and rejects any deviation.

### MAP record with two STRING-to-INT entries

```
04 09 00 00 00 00 00 30          Outer record: Type=MAP, Subtype=UNORDERED, Length=0x30
00 00 00 02                      Entry count = 2
01 00 00 00 00 00 00 03          Key 1: Type=STRING, Length=3
6B 65 79                         Value = "key"
02 00 00 00 00 00 00 08          Value 1: Type=INT, Length=8
00 00 00 00 00 00 00 0A          Value = 10
01 00 00 00 00 00 00 03          Key 2: Type=STRING, Length=3
74 6E 33                         Value = "tn3"
02 00 00 00 00 00 00 08          Value 2: Type=INT, Length=8
00 00 00 00 00 00 00 14          Value = 20
```

The hex above is the outer record's value; the surrounding header/trailer follow the same pattern as previous examples.

### Schema-validated record

```
0E 00 02 00 00 00 00 30          Outer record: Type=SYMBOL, Subtype=KEYWORD, Flags=SCHEMA, Length=0x30
D1 D1 01 02                      DescriptorMagic=0xD1D1 + DescriptorVersion=1 + FieldCount=2
04 4E 41 4D 45 01 00             Field 1: NameLen=4 + Name="NAME" + FieldType=STRING + Subtype=0x00
03 41 47 45 02 00                Field 2: NameLen=3 + Name="AGE" + FieldType=INT + Subtype=0x00
00 00 00 1B                      DescriptorLength=27
00 00 00 05 41 4C 49 43 45      NAME field: STRING(5) = "ALICE"
00 00 00 00 00 00 00 1E          AGE field: INT64 = 30
```

The decoder reads the descriptor first, then walks the field table and decodes each field according to its declared type.
