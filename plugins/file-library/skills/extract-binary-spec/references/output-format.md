# Output format

The `extract-binary-spec` skill produces a directory tree, not a single file. Each file follows a strict template so the consumer (`implement-binary-file-library` agent) can partition and load only what it needs — important when a format defines hundreds of record types.

## Directory layout

```
<format-name>/
├── SPEC.md
├── structures/
│   ├── <name>.md
│   └── ...
├── encoding-tables/
│   ├── <name>.md
│   └── ...
└── examples/
    ├── minimal.md
    ├── typical.md
    └── complex.md
```

Filenames are kebab-case of the structure / table / example name (e.g. `resource-record.md`, `chunk-ihdr.md`, `record-types.md`).

## File: `SPEC.md` (the index)

The format-wide overview plus an index that points to detail files. Keep it small — the index should stay navigable in one screen even for a 1000-record format.

```markdown
# <Format Name> Binary Specification Reference

## Overview
One paragraph: what the format is for, version(s) covered, and the governing
standard (RFC number, ISO number, vendor doc URL).

## Conventions
- **Byte order**: default endianness for the format (big-endian / little-endian / network byte order)
- **Bit numbering**: MSB-0 or LSB-0 convention used in bit field tables
- **Size units**: bytes / octets / words (and word size if not 8 bits)
- **Notation**: any diagram or notation conventions used in detail files

## Top-level structure
High-level description of how a complete file or message decomposes into
sub-structures. Optional summary diagram if the spec provides one.

## Structures index
One bullet per structure file with a one-line purpose:

- [`structures/header.md`](structures/header.md) — fixed 12-byte header on every message
- [`structures/question.md`](structures/question.md) — query record
- ...

## Encoding tables index
- [`encoding-tables/opcodes.md`](encoding-tables/opcodes.md) — operation codes
- ...

## Examples index
- [`examples/minimal.md`](examples/minimal.md)
- [`examples/typical.md`](examples/typical.md)
- [`examples/complex.md`](examples/complex.md)

## Appendix
- Maximum sizes and implementation limits
- Related RFCs or standards
- IANA registry references
- Version history summary
```

## File: `structures/<name>.md` (one per structure)

For every structure (header, record, frame, chunk, etc.), produce one file using this template. Omit any subsection the structure genuinely does not have.

### Structure variants

Most structures have a fixed byte-offset layout and use both the byte diagram and the field table. But some don't fit that mold — pick the right presentation rather than forcing a byte-offset table that doesn't apply:

- **Bit-only structures** (a single byte's bit packing — e.g. gzip's FLG byte, an 8-bit flags register). State `Layout: 1 byte, bit-packed` at the top, omit the byte diagram and field table, and put the full description in the **Bit fields** section. The structure file still needs a field table consumer downstream — use a one-row table whose Type column is `uint8` (or `uint16` if packed across two bytes) and link from its Description column to the Bit fields section below.
- **Recursive or algorithmic encodings** (DNS domain-name labels with compression pointers, ASN.1 BER length encoding, length-prefixed chains, TLV walks). Replace the byte-offset field table with an **Encoding** section that describes the algorithm prosaically and gives a per-element wire diagram for each branch (e.g. label vs. pointer). Include a short decode-pseudocode block when the algorithm has loops, recursion, or pointer-chasing — it'll be lifted directly into the Go decoder.
- **Single-field primitive payloads** (e.g. DNS RDATA-A is one 4-byte IPv4 address; RDATA-NS is one domain-name reference). Use the field table with a single row. Don't paraphrase the type ("a 32-bit IPv4 address") — name the Go type directly (`[4]byte`).

````markdown
# <Structure Name>

<One-line purpose: where this structure appears in the format and what it carries.>

**Byte order:** <state explicitly only when different from `SPEC.md#Conventions`.>

## Byte diagram

ASCII wire diagram. RFC-style bit diagram for fixed-size structures, simple
"field A | field B" diagram for variable-length sequences.

```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Field A              |          Field B              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

## Field table

| Offset (bytes) | Size | Type | Name | Description |
|---|---|---|---|---|
| 0 | 2 | uint16 | Length | Total message length in bytes |
| 2 | 1 | uint8 | Flags | See [Bit fields](#bit-fields) |
| 3 | 1 | uint8 | Reserved | Must be 0 |

For each column:
- **Offset**: byte offset from the start of this structure
- **Size**: in bytes (or `N bits` if the field is sub-byte and packed)
- **Type**: Go-friendly — `uint8`, `uint16`, `int32`, `[4]byte`, `[]byte`, ...
- **Name**: PascalCase, matching the spec's name where reasonable
- **Description**: one short line; richer detail belongs in the subsections below

## Bit fields

For fields that pack multiple values into one or more bytes:

| Bit(s) | Name | Description |
|---|---|---|
| 7 | QR | Query (0) or Response (1) |
| 6-3 | Opcode | Operation type — see [`../encoding-tables/opcodes.md`](../encoding-tables/opcodes.md) |
| 2 | AA | Authoritative Answer |

Use the bit-numbering convention from `SPEC.md#Conventions`. State it at the top of the file when it differs.

## Variable-length fields

For each variable-length field:
- **Length determination**: length-prefix / sentinel / end-of-message / TLV
- **Length prefix format**: size and type of the prefix; whether it counts itself; bytes vs. some other unit
- **Maximum length**: if the spec states one
- **Encoding**: character encoding for string-like fields (UTF-8, ASCII, EBCDIC, ...)

## Conditional / optional fields

For fields whose presence depends on another field's value:
- **Condition**: which field/flag/version determines presence
- **When present**: layout
- **When absent**: how the decoder should behave (skip, default, error)

## Checksums and integrity

If this structure carries a checksum:
- **Algorithm**: CRC-32, Internet checksum, HMAC, ...
- **Scope**: which bytes are covered by the computation
- **Byte order**: of the checksum value itself
- **Pseudo-header**: any virtual header included in the computation
- **Computation**: step-by-step

## Padding and alignment

If this structure has alignment requirements:
- **Alignment**: byte / word / dword
- **Pad value**: 0x00, 0xFF, EBCDIC space, ...
- **Where**: between fields, end of structure, ...

## Nested structures

If this structure contains other structures, link to them:
- Body contains zero or more [`resource-record`](resource-record.md) entries
- Header is wrapped by [`tcp-segment`](tcp-segment.md) when over TCP

## Versioning notes

If this structure differs across versions:
- **Version field location**: where the decoder reads the version
- **Per-version differences**: layout diffs
- **Backward compatibility**: how a decoder handles unknown versions

## Ambiguities

> **Ambiguity:** <use this callout when the spec is unclear or contradictory in this structure's definition>
````

## File: `encoding-tables/<name>.md` (one per lookup table)

Lookup tables map numeric values to meanings — opcodes, record types, error codes, message types.

```markdown
# <Table Name>

<One-line description and which structure(s) reference this table.>

| Value | Name | Description | Reference |
|---|---|---|---|
| 0 | QUERY | Standard query | RFC 1035 §4.1.1 |
| 1 | IQUERY | Inverse query (obsolete) | RFC 3425 |

## Notes
- Reserved / unassigned ranges
- Vendor-specific ranges
- Registry pointers (IANA, etc.)
```

## File: `examples/<name>.md` (worked examples)

At least three: `minimal.md`, `typical.md`, `complex.md`. Each is an annotated hex dump.

```markdown
# Minimal valid <format-name> message

<Short description of what this example exercises.>

```
Offset    Hex                                                ASCII
00000000  xx xx xx xx xx xx xx xx  xx xx xx xx xx xx xx xx  ................
00000010  xx xx xx xx xx xx xx xx                           ........
```

## Annotation

| Bytes | Field | Value | Notes |
|---|---|---|---|
| 0–1 | Length | 0x0018 (24) | Total message length |
| 2 | Flags | 0x80 | QR=1 (response) |
| ... |
```

Generate original hex (do not copy hex dumps verbatim from the spec). The bytes must encode a real, valid message a decoder would accept.

## Important rules (apply to every file)

- **Be precise about byte layout.** "A 16-bit field" is not enough — specify byte order, signed vs. unsigned, valid range, behavior on overflow.
- **Capture every field.** A decoder must consume every byte. Every wire field needs a row in some field table.
- **Always state byte order explicitly.** Globally in `SPEC.md#Conventions`; per-structure when it differs.
- **Distinguish bytes from bits.** Column headers `Offset (bytes)` and `Bit(s)`. Never mix without labels.
- **Map to Go types.** The Type column must hold a Go type name the implement-binary-file-library agent can drop directly into a struct definition — never a paraphrase like "32-bit IPv4 address" or "domain name reference". Allowed forms:
  - Fixed integers: `uint8`, `uint16`, `uint32`, `uint64`, `int8`–`int64`
  - Fixed-size byte arrays: `[4]byte`, `[16]byte`
  - Variable-length byte arrays: `[]byte`
  - References to another structure file: the structure's PascalCase name (e.g. `DomainName` for a reference to `structures/domain-name.md`). The downstream agent defines that type when it implements `domain-name.md`.
- **Capture length-prefix semantics precisely.** Every variable-length field: does the prefix count itself? Bytes or some other unit? Inclusive or exclusive of padding?
- **Don't quote large excerpts verbatim.** Summarize in your own words; generate original hex examples that exercise the same rules. Quote only minimal fragments with attribution (section/page).
- **Don't invent.** If the spec is silent, mark it unspecified. If contradictory, add a `> **Ambiguity:**` callout so the implementer can decide.
- **Use stable relative paths as anchors.** Cross-references between files use `../encoding-tables/opcodes.md` form, not page numbers or section IDs that may not survive editing.
