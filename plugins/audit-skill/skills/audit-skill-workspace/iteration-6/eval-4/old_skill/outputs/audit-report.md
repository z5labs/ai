# Audit: implement-graphql-resolver

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-4/old_skill/work/skills/implement-graphql-resolver/`
- Date: 2026-04-30
- Findings: 3  (idempotency: 0, reproducibility: 2, context-management: 0, strict-definitions: 1, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

- `SKILL.md:33` — "any relevant context the resolver phase will need to do its work well" gives no objective criterion; reproducibility requires a stated test (a list of fields, a regex, or a named condition). Two runs will write different `_context_schema.md` summaries.
- `SKILL.md:39` — "in enough detail that the test phase can write meaningful tests" leaves "enough" and "meaningful" undefined; name what the summary must contain (e.g. "for each resolver: argument types, return type, data-source calls, error cases").

### Context management

No findings.

### Strict definitions

- `SKILL.md:22` — output `src/resolvers.test.ts` is described as "created or appended" without stating which condition selects which behavior. State the rule explicitly (e.g. "created if absent, appended if present; existing tests are preserved").

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:12`): re-runs append only missing symbols, edits files in place rather than recreating.
- Description names what / when / when-to-skip, including a concrete skip clause for schema-only changes and missing service skeletons (`SKILL.md:3`).
- Inputs are declared with source, required-ness, and validation, including a refuse-and-instruct path when `src/schema.ts` or `src/resolvers.ts` is missing (`SKILL.md:14-17`).
- Phase handoffs use scratch files with explicit cleanup, so the main thread does not accumulate per-phase context (`SKILL.md:23`, `SKILL.md:47-49`).
- Heavy reads (full `src/schema.ts`, full `src/resolvers.ts`, full `SPEC.md`) are delegated to subagents rather than loaded into the orchestrator's context (`SKILL.md:31`, `SKILL.md:37`, `SKILL.md:43`).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-4/old_skill/work/audit-implement-graphql-resolver-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
