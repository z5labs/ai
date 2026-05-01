# Audit: implement-graphql-resolver

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-8/eval-4/old_skill/work/skills/implement-graphql-resolver/SKILL.md`
- Date: 2026-05-01
- Findings: 9  (idempotency: 0, reproducibility: 1, context-management: 8, strict-definitions: 0, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

- `SKILL.md:33` — "any relevant context the resolver phase will need to do its work well" gives no objective criterion; reproducibility requires a stated test (a named field set, a fixed schema, or "exactly the symbols listed in `_context_schema.md`'s table"). Two runs will diverge on what the summary includes.

### Context management

- `SKILL.md:31` — Phase 1 spawns one subagent covering the entire `SPEC.md` regardless of size; document a partitioning rule (e.g. one sub-call per spec section, split when type/query/mutation count > N) so per-call output stays bounded.
- `SKILL.md:37` — Phase 2 spawns one subagent covering all queries and mutations regardless of size; document a partitioning rule (e.g. one sub-call per resolver group) so per-call output stays bounded.
- `SKILL.md:43` — Phase 3 spawns one subagent that writes tests for every new resolver regardless of count; document a partitioning rule (e.g. one sub-call per N resolvers) so per-call output stays bounded.
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` is described in open-ended prose ("every type the spec mentions, along with notes on field types, nullability, deprecations, and any relevant context"); specify a strict format (e.g. one type per line: `TypeName: field1: T1, field2: T2?` plus a fixed deprecation column) so Phase 2 can rely on it without re-reading `src/schema.ts`.
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` has no line cap; specify a hard cap (e.g. ≤400 lines) so growth above the cap signals the spec was sized too large and must be chunked.
- `SKILL.md:39` — inter-phase artifact `_context_resolvers.md` is described in open-ended prose ("describing the new resolver functions and their data-source dependencies in enough detail that the test phase can write meaningful tests"); specify a strict format (e.g. one resolver per line: `resolverName(args) -> ReturnType; deps: [dataSource1, dataSource2]`) so Phase 3 can rely on it without re-reading `src/resolvers.ts`.
- `SKILL.md:39` — inter-phase artifact `_context_resolvers.md` has no line cap; specify a hard cap (e.g. ≤400 lines) so growth above the cap signals the work-unit was sized too large and must be chunked.
- `SKILL.md:37` — Phase 2 instructs the subagent to "read `src/schema.ts` in full" — that file was just mutated by Phase 1; rely on `_context_schema.md` from Phase 1 as the cross-reference of record, or pass a sliced (path, offset, limit) range covering only the appended region.
- `SKILL.md:43` — Phase 3 spawns the subagent with "`src/resolvers.ts` (re-read in full so the subagent sees every resolver, including those just added)" — that file was just mutated by Phase 2; rely on `_context_resolvers.md` from Phase 2 as the cross-reference of record, or pass a sliced (path, offset, limit) range covering only the appended region.

### Strict definitions

No findings.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:10–12`) — names the convergence behavior (append-only by exported name, conflicts surface as TS errors).
- Description names what / when / when-to-skip (`SKILL.md:3`) — all three triggering elements present, including the rare when-to-skip clause.
- Inputs section declares both inputs with source, required-ness, and a concrete validation rule (`SKILL.md:14–17`).
- Outputs section enumerates every artifact with its path and overwrite/append behavior (`SKILL.md:19–23`).
- Workflow delegates each phase to a subagent rather than doing the heavy reads in the main thread (`SKILL.md:29–45`) — the partitioning gap above is a refinement of an already-correct delegation discipline.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-8/eval-4/old_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
