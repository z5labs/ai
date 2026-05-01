---
name: implement-graphql-resolver
description: Multi-phase orchestrator that adds GraphQL resolvers to a TypeScript service. Phases run in order — schema definitions, resolver functions, tests — each delegated to a subagent. Use whenever the user asks to add resolvers, extend a GraphQL API, or wire up new query/mutation endpoints. Skip when the user wants schema-only changes or when no service skeleton exists yet.
---

# implement-graphql-resolver

Orchestrator that adds GraphQL resolvers to an existing TypeScript service. You delegate each phase (schema → resolvers → tests) to a focused subagent and pass small handoff summaries between them.

## Idempotency

This skill is idempotent — re-running on the same `SPEC.md` and service tree appends only the symbols that aren't already present. Existing resolver code is preserved by `Edit`-ing the source files rather than recreating them.

## Inputs

- **SPEC.md** (required) — Source: filesystem at `<service>/SPEC.md`. Lists the types, queries, and mutations to add.
- **Service directory** (required) — Source: filesystem path the user provides. Must contain `src/schema.ts` and `src/resolvers.ts`. If either file is missing, stop and ask the user to scaffold the service first.

## Outputs

- **Edits** to `<service>/src/schema.ts` and `<service>/src/resolvers.ts` — appended via `Edit`, never recreated wholesale.
- **Tests** at `<service>/src/resolvers.test.ts` — created or appended.
- **Scratch files** `<service>/_context_schema.md` (after Phase 1) and `<service>/_context_resolvers.md` (after Phase 2) — overwritten each run, deleted in Cleanup.

## Workflow

Run the three phases in order. Do not skip ahead.

### Phase 1: Schema

Spawn a subagent with the entire `SPEC.md` and the service directory. The subagent reads `SPEC.md`, then edits `src/schema.ts` to add every new type, query, and mutation listed in the spec.

When the subagent returns, run `npm run typecheck` to verify the schema compiles, then write `_context_schema.md` capturing the new schema additions. The summary should include every type the spec mentions, along with notes on field types, nullability, deprecations, and any relevant context the resolver phase will need to do its work well.

### Phase 2: Resolvers

Spawn a subagent with the entire `SPEC.md`, the service directory, and `_context_schema.md`. The subagent reads `src/schema.ts` in full to ground its work in the current schema state (including the additions Phase 1 just made), then edits `src/resolvers.ts` to add resolver functions for every new query and mutation.

When the subagent returns, run `npm run typecheck`, then write `_context_resolvers.md` describing the new resolver functions and their data-source dependencies in enough detail that the test phase can write meaningful tests.

### Phase 3: Tests

Spawn a subagent with the entire `SPEC.md`, `src/resolvers.ts` (re-read in full so the subagent sees every resolver, including those just added), and both `_context_schema.md` and `_context_resolvers.md`. The subagent writes `src/resolvers.test.ts` with one test per resolver.

When the subagent returns, run `npm test`.

### Cleanup

Delete `_context_schema.md` and `_context_resolvers.md`. Don't leave scratch files in the service tree.
