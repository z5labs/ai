# Audit: implement-graphql-resolver

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-4/with_skill/work/skills/implement-graphql-resolver/`
- Date: 2026-04-30
- Findings: 5  (idempotency: 0, reproducibility: 2, context-management: 3, strict-definitions: 0, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

- `SKILL.md:33` — "any relevant context the resolver phase will need to do its work well" gives no objective criterion; the Phase 1 subagent is left to guess what "relevant" and "do its work well" mean, so two runs produce two different summaries. State a fixed list of facts the summary must carry (e.g. for each new symbol: name, kind, signature, nullability, deprecation marker — and nothing else).
- `SKILL.md:39` — "in enough detail that the test phase can write meaningful tests" gives no objective criterion. Replace "enough detail" / "meaningful" with an enumerated schema (e.g. for each resolver: name, arg types, return type, data-source dependencies as a comma-separated list).

### Context management

- `SKILL.md:31` — Phase 1 spawns one subagent for the entire `SPEC.md` regardless of size; Phase 2 (`SKILL.md:37`) and Phase 3 (`SKILL.md:43`) do the same. Document a partitioning rule (e.g. "one sub-call per spec section / per type / per resolver group; if section count > N, split into sub-units") so per-call output stays bounded as the spec grows.
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` is described in open-ended prose ("summary should include every type the spec mentions, along with notes on field types, nullability, deprecations, and any relevant context"); same for `_context_resolvers.md` at `SKILL.md:39`. Specify a strict format (e.g. one symbol per line with a fixed column shape: `name | kind | signature | nullability | deprecation`) and a hard line cap (e.g. `≤400 lines; exceeding the cap is the signal the work-unit was sized too large and must be chunked`) so later phases can rely on the artifact instead of re-reading source.
- `SKILL.md:37` — Phase 2 instructs the subagent to "read `src/schema.ts` in full to ground its work in the current schema state (including the additions Phase 1 just made)"; Phase 3 (`SKILL.md:43`) likewise re-reads `src/resolvers.ts` "in full so the subagent sees every resolver, including those just added". Rely on `_context_schema.md` / `_context_resolvers.md` from the prior phase as the cross-reference of record, or pass a sliced `(path, offset, limit)` range — otherwise the summary's growth compounds with the source file's growth and the per-phase context budget is unbounded.

### Strict definitions

No findings.

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:12`) — re-runs append only missing symbols and existing code is preserved via `Edit`.
- Description names what (orchestrator that adds GraphQL resolvers), when-to-use ("Use whenever the user asks to add resolvers, extend a GraphQL API, or wire up new query/mutation endpoints"), and when-to-skip ("Skip when the user wants schema-only changes or when no service skeleton exists yet") (`SKILL.md:3`).
- Inputs section names each input, source, and required-ness with a refuse-and-instruct fallback for the missing-skeleton case (`SKILL.md:14-17`).
- Outputs section enumerates every artifact path, edit-vs-recreate behavior, and the scratch-file overwrite/delete lifecycle (`SKILL.md:19-23`).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-4/with_skill/work/audit-implement-graphql-resolver-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
