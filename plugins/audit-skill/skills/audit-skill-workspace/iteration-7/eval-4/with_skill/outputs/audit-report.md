# Audit: implement-graphql-resolver

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-4/with_skill/work/skills/implement-graphql-resolver/SKILL.md`
- Date: 2026-04-30
- Findings: 7  (idempotency: 0, reproducibility: 2, context-management: 5, strict-definitions: 0, security: 0)

## Findings

### Idempotency

No findings.

(Idempotency declaration is present and specific at `SKILL.md:12`; remaining checks skipped per the audit-skill skip condition.)

### Reproducibility

- `SKILL.md:33` — "any relevant context the resolver phase will need to do its work well" gives no objective criterion; reproducibility requires a stated test (a fixed list of fields to capture, a regex, a named condition).
- `SKILL.md:39` — "in enough detail that the test phase can write meaningful tests" gives no objective criterion; specify exactly which facts the summary must include (e.g. for each resolver: name, argument types, return type, named data-source dependency).

### Context management

- `SKILL.md:31` — Phase 1 spawns one subagent covering the entire `SPEC.md` regardless of size; document a partitioning rule (e.g. one sub-call per spec type/section, split when count > N) so per-call edits to `src/schema.ts` stay bounded. Same gap applies to Phase 2 (`SKILL.md:37`) and Phase 3 (`SKILL.md:43`).
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` is described in open-ended prose ("capturing the new schema additions ... notes on field types, nullability, deprecations, and any relevant context"); specify a strict format (e.g. one type per line: `TypeName { field: Type! ... }`, one row per query/mutation with signature only, no prose) so Phase 2 can rely on it without re-reading `src/schema.ts`.
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` has no line cap; specify a hard cap (e.g. ≤400 lines) so growth above the cap signals the work-unit was sized too large and must be chunked.
- `SKILL.md:39` — inter-phase artifact `_context_resolvers.md` is described in open-ended prose ("describing the new resolver functions and their data-source dependencies in enough detail"); specify a strict format (e.g. one resolver per line with `name(args): ReturnType — uses <data-source>`) so Phase 3 can rely on it without re-reading `src/resolvers.ts`. Also lacks a line cap — specify a hard cap (e.g. ≤400 lines).
- `SKILL.md:37` — Phase 2 instructs subagent to Read `src/schema.ts` in full after Phase 1 mutated it; rely on `_context_schema.md` from Phase 1 as the cross-reference of record, or pass a sliced (path, offset, limit) range covering only the just-added section. Same pattern at `SKILL.md:43` — Phase 3 re-reads `src/resolvers.ts` in full after Phase 2 mutated it; rely on `_context_resolvers.md` instead.

### Strict definitions

No findings.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:12`) — names exactly what re-run does (appends only missing symbols, preserves existing code via `Edit`).
- Description names what / when / when-to-skip (`SKILL.md:3`) — "adds GraphQL resolvers", trigger phrasings ("add resolvers, extend a GraphQL API, wire up new query/mutation endpoints"), and a negative case ("Skip when the user wants schema-only changes or when no service skeleton exists yet").
- Inputs declared with source, required-ness, and a precondition check (`SKILL.md:15-17`) — `SPEC.md` and service directory both typed; missing `src/schema.ts` / `src/resolvers.ts` triggers a refuse-and-instruct.
- Outputs declared with paths and per-file overwrite/append behavior (`SKILL.md:19-23`).
- Cleanup phase is explicit (`SKILL.md:47-49`) — scratch files have a documented end-of-run disposition.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-4/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
