# Skill Benchmark: audit-skill

**Iteration**: 6
**Date**: 2026-05-01T00:34:39Z
**Evals**: 0, 1, 2, 3, 4 (1 run each per configuration)

## Summary

| Metric | Old Skill | With Skill | Delta |
|--------|-----------|------------|-------|
| Pass Rate | 90% | 95% | +0.04 |
| Time | 116.6s | 121.9s | +5.3s |
| Tokens | 40934 | 42900 | +1966 |

## Per-eval pass rates

| Eval | Name | Old Skill | With Skill | Δ |
|------|------|-----------|------------|---|
| 0 | bad-skill-finds-violations | 100% (17/17) | 100% (17/17) | +0.00 |
| 1 | good-skill-stays-quiet | 100% (9/9) | 100% (9/9) | +0.00 |
| 2 | name-resolution-from-claude-skills | 100% (3/3) | 100% (3/3) | +0.00 |
| 3 | insecure-skill-finds-security-violations | 82% (9/11) | 73% (8/11) | -0.09 |
| 4 | orchestrator-skill-finds-output-context-violations | 69% (9/13) | 100% (13/13) | +0.31 |

## Observations

- **eval-4 (the new orchestrator-skill fixture) is the load-bearing test**: with_skill catches all four output-context patterns (unbounded per-phase scope, open-ended `_context_*` summaries, missing line cap, re-read of mutated source); old_skill does not, failing exactly the four new-check assertions. The +0.31 pass-rate delta on eval-4 is the new checks doing their job.
- **No regressions on existing evals 0–2**: both configs pass identical assertions. The new checks fire only on multi-phase orchestrator skills, so single-phase fixtures remain quiet.
- **eval-3 (insecure-skill) failures are pre-existing grader limitations, not audit regressions**:
    - `target-not-modified` false-negative — the grader compares against `SKILL.md` only, but the multi-file fixture's `target-after-audit.md` is a concatenation of SKILL.md + references/database.md (per the subagent's instructions). Audit was read-only.
    - `finds-references-side-violation` — no handler defined in iteration-5's grade.py; pre-existing missing handler.
    - `no-severity-tiers` (with_skill only) — false-positive: regex matched 'High-level' inside a quoted phrase from the target, not an actual severity label.
- The new context-management.md checks did NOT fire spuriously on any single-phase fixture (evals 0, 1, 2, 3) — both with_skill reports show `Context management — No findings.` (eval-3) or only the pre-existing inline-template finding (eval-0).