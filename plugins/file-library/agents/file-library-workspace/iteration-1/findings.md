# DSF dry-run — iteration-1 findings

**Issue:** [#88](https://github.com/z5labs/ai/issues/88)
**Date:** 2026-05-04
**Spec source:** <https://developer.x-plane.com/article/dsf-file-format-specification/>
**Outcome:** stuck-exit (spec scope larger than 15-iteration budget) — every acceptance criterion met.

## Acceptance criteria — status

| AC | Status | Evidence |
|---|---|---|
| Agent runs end-to-end OR documents friction in `_state_of_play.md` | ✓ | `eval-0-dsf-binary/outputs/dsf/_state_of_play.md` |
| At least one real-world `.dsf` round-trips byte-equal | ✓ | `TestRoundTripFromTestdata/dsf_real_world_dsf` PASS — 152,976 bytes in/out match exactly |
| Findings written up summarizing what worked, what stuck, what to change | ✓ | This file |

## What worked first try

- **Step 1 format detect** — user-supplied `binary` hint accepted; no ambiguity to escalate.
- **Step 2 scaffold** (`new-go-binary-file-library dsf`) — green on first invocation; all stub tests passed.
- **Step 4 iteration-1 implement** (top-level container: `FileHeader`, `Atom` envelope, MD5 footer) — landed 29 passing tests on a single implement-skill call. Scope sizing (~250 lines, 5 chunked files) was well under the 600-line / 8-file partition gate. The action-loop, `FieldError → OffsetError → leaf` chain, and round-trip discipline all materialized correctly with no rework.
- **Step 5 Gate (b)** (`review-file-library`) — produced a structurally-organized 50-blocker audit with zero false positives. The handoff artifact converted directly into a 6-step multi-run decomposition in `_state_of_play.md`.

## Where the agent stuck

After iteration 1, Gate (b) reported 50 blockers — all true positives, all reflecting deliberately-unimplemented surface area:

- per-atom payload typing for HEAD / PROP / DEFN / GEOD / DEMS / CMDS
- planar-numeric RLE + differencing over `uint16` (POOL) and `uint32` (PO32)
- 30+ command opcode payloads (state / object / network / polygon / mesh / comment families)
- encoding-table enums (kind codes)
- examples-bytes round-trip tests against `examples/{minimal,typical,complex}.md`

The orchestrator declared `stuck-scope-too-large` rather than spin the remaining 14-iteration budget on a target it could not converge in one run. The 6-step decomposition for completing DSF across multiple smaller autonomous runs is in `outputs/dsf/_state_of_play.md`.

## Friction observed → follow-up issues filed

Three concrete improvements to the agent or its underlying skills, each filed:

1. **[#91](https://github.com/z5labs/ai/issues/91) — `extract-binary-spec`: add `html2text` / `lynx` fallbacks to the Phase 1 HTML→text tool table.** Pandoc was unavailable on this host; `html2text -b 200` produced acceptable output for the X-Plane DSF page but was discovered empirically.

2. **[#92](https://github.com/z5labs/ai/issues/92) — `extract-binary-spec`: strip vendor HTML chrome before extraction.** ~180 of 1040 lines (~17%) of the X-Plane developer page were nav/sidebar/footer/auto-TOC. Inline tier-1 extraction absorbed the noise, but tier-3 fan-out would amplify it across every subagent's read window.

3. **[#93](https://github.com/z5labs/ai/issues/93) — `file-library`: early "spec too large; re-scope" signal based on audit-blocker count.** The current stuck conditions (3 no-progress iterations, 15-iteration hard cap) fire only after meaningful effort. A Gate (b)-time check on `audit_blockers / blockers_per_iteration_so_far` against remaining iteration budget would catch over-scoped runs in iteration 1 or 2, with the same multi-run decomposition output.

## Spec ambiguities (X-Plane spec, not z5labs/ai)

Twenty-three `> **Ambiguity:**` callouts surfaced across the spec artifacts. These are properties of the X-Plane spec (typos, undocumented edge cases) and live in `outputs/dsf/SPEC.md` and `outputs/dsf/structures/*.md` for human review before any future runs commit decisions about them. They are *not* filed as z5labs/ai issues. Highlights:

- DSF spec text says `sim/south` is the *northern* edge and vice versa (typo).
- DEMI flag value `4` is described as "bit 3" but is `0x0004` = bit 2 in LSB-0 numbering.
- `COMMENT 16` (opcode 33) is missing its `uint16 length` prefix line in the source HTML.
- Command opcodes 19–22 are a deliberate gap; opcode 255 is reserved with undefined behaviour.
- Planar-numeric RLE: count of *elements* vs *bytes* when element width > 1 — must be elements; spec is silent.
- Top-level atom ordering not mandated.
- Differencing endianness (logical vs on-disk).

Full list: see `outputs/dsf/_state_of_play.md` § "Ambiguities surfaced in spec artifacts".

## Real-world fixture note

`outputs/dsf/testdata/dsf-real-world.dsf` is a 152,976-byte tile (`+40-080.dsf`) extracted from an X-Plane scenery distribution. The file is X-Plane-licensed and **not redistributable**; it is excluded from commit by `iteration-1/.gitignore` (pattern: `**/testdata/*.dsf`). Only the test wiring (`TestRoundTripFromTestdata` in `outputs/dsf/encoder_test.go`) is committed.

A reviewer reproducing the round-trip locally needs to drop their own `.dsf` at `outputs/dsf/testdata/dsf-real-world.dsf` (or any name matching the table entry) before running `(cd outputs/dsf && go test -race -v ./...)`. The byte-equal assertion holds for any DSF whose layout the iteration-1 implementation handles — i.e., any DSF whose top-level Atom envelope is well-formed and whose payload bytes are stable under decode→encode at the opaque-`[]byte` level. That covers the entire X-Plane DSF corpus today, since iteration-1 does not interpret payload bytes.

## Suggested next runs (from `_state_of_play.md`)

Six smaller autonomous runs to complete the package, in dependency order:

1. HEAD/PROP — `StringTableAtom` walking helper + `PropAtom { (name, value) }`. ~1–2 iterations.
2. DEFN — `Atom-of-atoms` walker + `DefnAtom { Tert, Objt, Poly, Netw, Demn []string }`. ~1–2 iterations.
3. GEOD planar-numeric — `PlanarNumeric[T]` over `uint16` and `uint32` with RAW/DIFFERENCED/RLE/RLE_DIFFERENCED, plus `SCAL`/`SC32`. **Hardest sub-area.** ~3 iterations.
4. DEMS / DEMI / DEMD. ~1 iteration.
5. CMDS — 30+ opcodes split into 5 partitions (state, object, network, polygon, mesh). ~5 iterations.
6. Real-world `add-fixture` round-trip — keep last so any byte-level disagreement is unambiguously a logic bug.

Each is a candidate for a separate `/file-library` invocation; the `Resume` path in Step 0 picks up the existing `outputs/dsf/` package on each run.
