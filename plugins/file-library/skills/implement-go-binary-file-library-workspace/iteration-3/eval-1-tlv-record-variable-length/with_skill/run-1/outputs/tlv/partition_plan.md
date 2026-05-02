# Partition Plan

Scope: implement the `Record` structure for TLV1 (Type uint8, Length uint16 BE, Value []byte), add a `RecordType` enum (STRING, INT, BLOB, NESTED), update `File` to hold `[]Record`, and wire up decode/encode + tests (including round-trip per record type, an empty record, an INT record with Length=8, and a typical STRING record).

`SPEC.md` present; chunked layout (`structures/`, `encoding-tables/`) absent.

## types phase

no partitioning needed (slice total: 81 lines, chunked files: 0)

- Overview (lines 3-15, 13 lines)
- Conventions (lines 16-22, 7 lines)
- Field Definitions (lines 23-66, 44 lines)
- Encoding Tables (lines 67-79, 13 lines)
- Versioning (lines 94-97, 4 lines)

## decoder phase

no partitioning needed (slice total: 121 lines, chunked files: 0)

- Overview (lines 3-15, 13 lines)
- Conventions (lines 16-22, 7 lines)
- Field Definitions (lines 23-66, 44 lines)
- Encoding Tables (lines 67-79, 13 lines)
- Conditional and Optional Fields (lines 80-85, 6 lines)
- Checksums and Integrity (lines 86-89, 4 lines)
- Padding and Alignment (lines 90-93, 4 lines)
- Examples (lines 98-127, 30 lines)

## encoder phase

no partitioning needed (slice total: 115 lines, chunked files: 0)

- Overview (lines 3-15, 13 lines)
- Conventions (lines 16-22, 7 lines)
- Field Definitions (lines 23-66, 44 lines)
- Encoding Tables (lines 67-79, 13 lines)
- Checksums and Integrity (lines 86-89, 4 lines)
- Padding and Alignment (lines 90-93, 4 lines)
- Examples (lines 98-127, 30 lines)
