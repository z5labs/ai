# Partition plan

User request: implement string-typed records (`record string KEY = "value"`) end-to-end.

SPEC.md sections (line ranges):
- Overview: 3-18 (16 lines)
- Lexical Elements (Tokens): 19-80 (62 lines)
- Structure (Grammar): 81-111 (31 lines)
- Semantics: 112-119 (8 lines)
- Examples: 120-158 (39 lines)

Per-phase slice totals:
- tokenizer: Overview + Lexical Elements + Examples = 16 + 62 + 39 = 117 lines
- parser: Overview + Structure + Semantics + Examples = 16 + 31 + 8 + 39 = 94 lines
- printer: Overview + Structure + Semantics + Examples = 16 + 31 + 8 + 39 = 94 lines

Chunked input layout (`tokens/`, `grammar/`): not present (chunked file count = 0 for every phase).

Scope-gate threshold is 600 lines / 8 chunked files per phase. No phase trips the threshold.

## tokenizer phase

no partitioning needed (slice total: 117 lines, chunked files: 0)

## parser phase

no partitioning needed (slice total: 94 lines, chunked files: 0)

## printer phase

no partitioning needed (slice total: 94 lines, chunked files: 0)
