# Audit: code-review-checklist

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-5/old_skill/work/skills/code-review-checklist/SKILL.md`
- Date: 2026-04-30
- Findings: 5  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 5, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:11` — idempotency declaration claims re-running "overwrites the previous review report", but the workflow (lines 39, 45, 51, 59) describes each phase appending its section to the report and the Final assembly step (line 63) prepending a summary; there is no truncate / overwrite-from-empty step before Phase 1, so a second run accumulates onto prior content. Either add an explicit "truncate `<report_path>` before Phase 1" step or restate the declaration as "appends across runs".
- `SKILL.md:4` — `argument-hint: "<path-to-diff>"` declares one positional, but Inputs at lines 17–18 declares two (`diff_path` required, `report_path` optional). Update to `argument-hint: "<path-to-diff> [report-path]"` so callers see the full calling shape.
- `SKILL.md:18` — input `report_path` has no validation rule (must the parent directory exist? is any path shape rejected? what if the default collides with an unrelated file?). State the validation explicitly or note "no validation; the Write tool's behavior governs".
- `SKILL.md:43` — Phase 2's "Phase 1's mechanical triggers carry over: if Phase 1 raised a finding on a hunk and no test was added for that hunk, raise a paired Tests finding" does not state where Phase 2 reads Phase 1's per-hunk findings from (in-memory list? re-parse the `## Correctness` section already written to the report?). Name the carry-over channel so Phase 2's precondition on Phase 1's output is explicit.
- `SKILL.md:63` — Final assembly is described as "prepend the line ... to the report file", but prepending is not a single-step operation (read-all + rewrite? a marker-line replacement?) and the surrounding workflow has each phase append. Spell out the prepend procedure so two runs produce the same byte sequence.

### Security

No findings.

## Passing checks

- Idempotency stance is declared explicitly at `SKILL.md:11` ("This skill is idempotent — re-running on the same diff overwrites the previous review report"); the declaration is present and specific even though the strict-definitions finding above flags an inconsistency between this declaration and the workflow steps.
- Description at `SKILL.md:3` names what (`Walk a single PR diff ... produce a markdown review report`), when to use (`when the user asks for a structured review on a specific diff or patch file`), and when to skip (`security audit, a performance review, or feedback on more than one PR at a time`) — all three triggering elements present, including a concrete negative-case list.
- Inputs section at `SKILL.md:17–18` names each input, its source (positional CLI arg), required-vs-optional with a default, and (for `diff_path`) a validation rule (`[ -f "$diff_path" ]`).
- Citation format is fixed at `SKILL.md:24–26` with a concrete derivation rule from the hunk header (`c + offset`) — removes a common reproducibility leak in review-style skills.
- Phase 1 and Phase 4 use mechanical triggers (specific syntactic patterns: `<=` against `len(...)`, unguarded dereferences, trailing whitespace, unreferenced imports) rather than subjective judgment, keeping the review reproducible across runs.
- The skill does not route any secret through the model's context — no prompts for credentials, no credential-shaped arguments, no URL-form connection strings, no generated files containing secrets.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-5/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
