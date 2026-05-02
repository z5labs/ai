# Partition plan

## types phase

no partitioning needed (slice total: 342 lines, chunked files: 0)

Sections sliced: Overview (3-17, 15 lines), Conventions (18-29, 12 lines),
Field Definitions (30-200, 171 lines), Encoding Tables (201-338, 138 lines),
Versioning (472-477, 6 lines).

## decoder phase

partitioned into 3 sub-units:

- sub-unit 1: Overview + Conventions + Field Definitions (198 lines: 15+12+171)
- sub-unit 2: Overview + Conventions + Encoding Tables + Conditional/Optional Fields (258 lines: 15+12+138+93)
- sub-unit 3: Overview + Conventions + Checksums and Integrity + Padding and Alignment + Examples (203 lines: 15+12+32+8+136)

decoder slice total before partitioning: 605 lines (Overview 15 + Conventions 12 +
Field Definitions 171 + Encoding Tables 138 + Conditional/Optional 93 +
Checksums 32 + Padding 8 + Examples 136). Trips the 600-line gate. Each sub-unit
stays under the 300-line per-sub-unit cap and always carries Overview + Conventions.

## encoder phase

no partitioning needed (slice total: 512 lines, chunked files: 0)

Sections sliced: Overview (3-17, 15 lines), Conventions (18-29, 12 lines),
Field Definitions (30-200, 171 lines), Encoding Tables (201-338, 138 lines),
Checksums and Integrity (432-463, 32 lines), Padding and Alignment (464-471, 8 lines),
Examples (478-613, 136 lines).

## decoder sub-unit 1 — starting (17:43:31)

## decoder sub-unit 1 — done (17:45:12)

## decoder sub-unit 2 — starting (17:45:12)

## decoder sub-unit 2 — done (17:45:32)

Note: this sub-unit's spec slices (Encoding Tables + Conditional/Optional Fields)
cover Record.Type/Subtype, Compression/Encryption algorithms, FieldType, Symbol
kinds, Extension tags, and conditional-field rules for SIGNED / INDEXED-EXTENDED
consistency / CHUNKED / TIMESTAMPED / LIST / MAP / REFERENCE / NESTED / SEALED /
STRICT / NULL. The user's request scopes the implementation to the Header
fields only (Magic / Version / Flags / ChecksumAlg / Reserved1 / IndexCount /
ExtCount / TrailerOffset) plus the seven defined flags and the Checksum
algorithms enum, all of which were already wired in sub-unit 1. The only
header-relevant rule from this slice — INDEXED iff IndexCount>0 / EXTENDED iff
ExtCount>0 — was not in the user prompt and is deferred until records and
trailer come online (without those, the consistency check is half a check). No
new decoder symbols added.

## decoder sub-unit 3 — starting (17:45:32)

## decoder sub-unit 3 — done (17:45:56)

Note: this sub-unit's spec slices (Checksums and Integrity, Padding and
Alignment, Examples) confirm the byte layout already implemented in sub-unit 1.
The Checksums and Integrity slice notes ChecksumAlg=0 is reserved and must be
rejected with UnknownChecksumAlgError{Alg: 0}; sub-unit 1's switch-default
already covers that. The Examples section's minimal-file hex dump matches the
TestDecodeHeaderHappyPath fixtures byte-for-byte. No new decoder symbols added.

