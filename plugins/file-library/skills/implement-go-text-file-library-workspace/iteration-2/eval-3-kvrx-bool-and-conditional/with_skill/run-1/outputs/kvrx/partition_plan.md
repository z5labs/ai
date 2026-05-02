# Partition plan for kvrx (record-bool + Conditional)

Spec section line ranges (from `grep -n '^## ' SPEC.md` and `wc -l SPEC.md = 731`):

- Overview: 3-31 (29 lines)
- Lexical Elements (Tokens): 32-139 (108 lines)
- Structure (Grammar): 140-415 (276 lines)
- Semantics: 416-631 (216 lines)
- Examples: 632-731 (100 lines)

## tokenizer phase

no partitioning needed (slice total: 237 lines, chunked files: 0)

Slices: Overview (3-31), Lexical Elements (32-139), Examples (632-731).

## parser phase

partitioned into 2 sub-units:

- sub-unit 1 (record-bool): Overview (3-31), Top-level grammar (156-165), Records (166-185), Types (269-289), Primary expressions (320-341), Bool type inference (469-472), Statement disambiguation (383-400), Examples (632-731) — ~224 sliced lines
- sub-unit 2 (Conditional): Overview (3-31), Conditionals grammar (247-268), References (497-507), Conditionals at parse time (508-525), Examples (632-731) — ~180 sliced lines

Slice total before partitioning: 621 lines (Overview + Structure (Grammar) + Semantics + Examples) — over the 600-line gate.

## parser sub-unit 1 — starting (17:43:11)

## parser sub-unit 1 — done (17:44:51)

## parser sub-unit 2 — starting (17:44:51)

## parser sub-unit 2 — done (17:46:54)

## printer phase

partitioned into 2 sub-units:

- sub-unit 1 (record-bool printing): Overview (3-31), Top-level grammar (156-165), Records (166-185), Primary expressions (320-341), Examples (632-731) — ~180 sliced lines
- sub-unit 2 (Conditional printing): Overview (3-31), Conditionals grammar (247-268), References (497-507), Examples (632-731) — ~162 sliced lines

Slice total before partitioning: 621 lines — over the 600-line gate.

## printer sub-unit 1 — starting (17:47:09)

## printer sub-unit 1 — done (17:48:21)

## printer sub-unit 2 — starting (17:48:21)

## printer sub-unit 2 — done (17:48:21)

Note on printer sub-units: sub-unit 1 (record-bool) and sub-unit 2 (Conditional) were implemented together because the conditional's body recursively prints record statements via the same `printExpression` / `printRecord` helpers. The boundary remains useful for the up-front spec-slice budget (~149 + ~162 lines) but the implementation step itself was atomic; both sub-units are in `printer.go` with no overlap or rewrite. Round-trip tests cover both sub-units' surface together.
