# Audit: implement-graphql-resolver

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-4/old_skill/work/skills/implement-graphql-resolver/SKILL.md`
- Date: 2026-04-30
- Findings: 7  (idempotency: 0, reproducibility: 2, context-management: 4, strict-definitions: 1, security: 0)

## Findings

### Idempotency

No findings.

### Reproducibility

- `SKILL.md:33` — "any relevant context the resolver phase will need to do its work well" gives no objective criterion for what to include in `_context_schema.md`; two runs on the same spec will produce different summaries. Name the fields the summary must contain (e.g. "for each type: name, parent, fields with type+nullability, directives") and stop there.
- `SKILL.md:39` — "in enough detail that the test phase can write meaningful tests" leaves the size and shape of `_context_resolvers.md` to the model's judgment. State a concrete schema (e.g. "one row per resolver: name, arg types, return type, data-source calls") so two runs converge.

### Context management

- `SKILL.md:31` — Phase 1 spawns one subagent for the entire `SPEC.md` regardless of size; per-call output (schema edits) grows unboundedly with spec size. Document a partitioning rule (e.g. one sub-call per spec section, or split when type count > N).
- `SKILL.md:33` — inter-phase artifact `_context_schema.md` is described in open-ended prose ("The summary should include every type the spec mentions, along with notes on field types, nullability, deprecations, and any relevant context..."); specify a strict format (e.g. one type per line with a fixed column shape) and a hard line cap so Phase 2 can rely on it without re-reading source.
- `SKILL.md:37` — Phase 2 instructs the subagent to `Read src/schema.ts` in full after Phase 1 mutated it; the `_context_schema.md` summary should be the cross-reference of record. Either drop the full re-read or pass a sliced `(path, offset, limit)` range covering only the new additions.
- `SKILL.md:43` — Phase 3 instructs the subagent to re-read `src/resolvers.ts` in full after Phase 2 mutated it ("re-read in full so the subagent sees every resolver, including those just added"); rely on `_context_resolvers.md` from Phase 2 as the cross-reference of record, or pass a sliced range covering only the new resolvers.

### Strict definitions

- `SKILL.md:31` — Phase 1 step says "add every new type, query, and mutation listed in the spec" but doesn't state the precondition/test for detecting symbols that already exist; the idempotency claim on line 12 ("appends only the symbols that aren't already present") depends on this test being defined. State how the subagent decides whether a symbol already exists (name match, signature match, etc.).

### Security

No findings.

## Passing checks

- Idempotency declaration is present and specific (`SKILL.md:10-12`): names the re-run shape ("appends only the symbols that aren't already present") and the mechanism (`Edit`, not recreate).
- Description names what / when / when-to-skip (`SKILL.md:3`): all three triggering elements present, including a clear negative case ("schema-only changes" / "no service skeleton").
- Inputs declared with source and required-ness (`SKILL.md:15-17`): `SPEC.md` and service directory both have explicit filesystem sources and a refuse-and-instruct fallback for missing files.
- Outputs declared with overwrite/append behavior (`SKILL.md:20-23`): each output names its path and what happens to a pre-existing file (Edit/append/overwrite each run/created-then-appended).
- Cleanup step is explicit (`SKILL.md:47-49`): scratch files deleted at end, not left behind.
- No security surface — skill handles no credentials, no env-routed secrets, no URL connection strings.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration.
