# Partition plan

SPEC.md section line ranges (from `grep -n '^## ' SPEC.md` and `wc -l`):

| Section                | Range    | Lines |
|------------------------|----------|-------|
| Overview               | 3-18     | 16    |
| Lexical Elements       | 19-80    | 62    |
| Structure (Grammar)    | 81-111   | 31    |
| Semantics              | 112-119  | 8     |
| Examples               | 120-158  | 39    |

## tokenizer phase

no partitioning needed (slice total: 117 lines, chunked files: 0)

Slices: Overview (3-18), Lexical Elements (19-80), Examples (120-158).

## parser phase

no partitioning needed (slice total: 94 lines, chunked files: 0)

Slices: Overview (3-18), Structure (Grammar) (81-111), Semantics (112-119), Examples (120-158).

## printer phase

no partitioning needed (slice total: 94 lines, chunked files: 0)

Slices: Overview (3-18), Structure (Grammar) (81-111), Semantics (112-119), Examples (120-158).
