# Audit: implement-graphql-resolver

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-8/eval-4/with_skill/work/skills/implement-graphql-resolver/SKILL.md`
- Date: 2026-05-01
- Findings: 9  (idempotency: 0, reproducibility: 0, context-management: 9, strict-definitions: 0, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

- `SKILL.md:31` — Phase 1 spawns one subagent covering every type, query, and mutation in `SPEC.md` regardless of spec size; per-call edit output to `src/schema.ts` grows unboundedly with the spec. Document a partitioning rule (e.g. one sub-call per spec section, or split when type+query+mutation count exceeds N).
- `SKILL.md:37` — Phase 2 spawns one subagent covering every new query and mutation regardless of count; per-call edit output to `src/resolvers.ts` grows unboundedly with the spec. Document a partitioning rule (e.g. one sub-call per resolver group, or split when resolver count exceeds N).
- `SKILL.md:43` — Phase 3 spawns one subagent that writes one test per resolver with no upper bound on resolver count; the test file's per-call output grows 1:1 with Phase 2's output. Document a partitioning rule symmetric to Phase 2's.
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` is described in open-ended prose ("capturing the new schema additions", "any relevant context the resolver phase will need to do its work well"); specify a strict format (e.g. one type per line with `name: <T>; fields: <f1:T1, f2:T2…>; nullable: <list>; deprecated: <list>`) so Phase 2 can rely on it without re-reading `src/schema.ts`.
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` has no line cap; specify a hard cap (e.g. ≤400 lines) so growth above the cap signals the spec was sized too large for one Phase 1 call and must be chunked.
- `SKILL.md:39` — inter-phase artifact `_context_resolvers.md` is described in open-ended prose ("describing the new resolver functions and their data-source dependencies in enough detail that the test phase can write meaningful tests"); specify a strict format (e.g. one resolver per line with `name; signature; data-sources; side-effects`) so Phase 3 can rely on it without re-reading `src/resolvers.ts`.
- `SKILL.md:39` — inter-phase artifact `_context_resolvers.md` has no line cap; specify a hard cap (e.g. ≤400 lines) so growth above the cap signals Phase 2's work-unit was sized too large and must be chunked.
- `SKILL.md:37` — Phase 2 instructs its subagent to "read `src/schema.ts` in full to ground its work in the current schema state (including the additions Phase 1 just made)"; the file Phase 1 just `Edit`-ed is the same file Phase 2 reads whole, so `_context_schema.md` is doing no work and the subagent's input grows with every prior Phase 1 run. Rely on `_context_schema.md` (once strictly formatted, per the check above) as the cross-reference of record, or pass a sliced `Read(path, offset, limit)` range covering only the new symbols.
- `SKILL.md:43` — Phase 3 instructs its subagent to take "`src/resolvers.ts` (re-read in full so the subagent sees every resolver, including those just added)"; the file Phase 2 just `Edit`-ed is the same file Phase 3 reads whole, so `_context_resolvers.md` is doing no work and Phase 3's input grows with every prior Phase 2 run. Rely on `_context_resolvers.md` (once strictly formatted) as the cross-reference of record, or pass a sliced `Read(path, offset, limit)` range covering only the new resolvers.

### Strict definitions

No findings.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:10-12` — names exactly what re-running does: append-only by exported name, conflicts surface as TypeScript errors and stop the run).
- Outputs section names path, format, and pre-existing-file behavior for each artifact (`SKILL.md:21-23` — `Edit`-ed never recreated, test file appended on subsequent runs, scratch files overwritten then deleted).
- Cleanup phase deletes both scratch files (`SKILL.md:49`), so the next run starts from a clean inter-phase state.
- Description names what / when / when-to-skip in three concrete clauses (`SKILL.md:3` — "add resolvers, extend a GraphQL API, or wire up new query/mutation endpoints" for triggering; "schema-only changes" and "no service skeleton exists yet" for skipping).
- Inputs section declares source, required-ness, and a validation rule for each input (`SKILL.md:16-17` — `SPEC.md` at `<service>/SPEC.md`; service directory must contain `src/schema.ts` and `src/resolvers.ts`, with refusal-to-proceed if missing).
- Workflow states "Run the three phases in order. Do not skip ahead." (`SKILL.md:27`) so the phase ordering is explicit rather than implied by section order.
- Security: skill consumes no credentials at all (no env vars, no argument-hints, no prompted secrets, no generated `.env` files); the workflow operates entirely on local filesystem paths the user provides.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-8/eval-4/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
