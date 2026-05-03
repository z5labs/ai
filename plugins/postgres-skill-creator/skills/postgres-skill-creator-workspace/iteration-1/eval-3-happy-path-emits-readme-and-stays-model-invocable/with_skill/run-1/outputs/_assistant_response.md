Generated the project-level skill `pg-orders` from the `orders` Postgres schema.

**Path:** `.claude/skills/pg-orders/`

**Captured:** 4 tables, 1 view, 0 enums.

**Files written:**

- `.claude/skills/pg-orders/SKILL.md` — model-invocable (no `disable-model-invocation`); description names the top tables (orders, products, users, order_items) so the skill fires on natural-language prompts about the orders DB.
- `.claude/skills/pg-orders/README.md` — human-facing onboarding, with the one-time `.env` setup, smoke test, row-count sample (`SELECT count(*) FROM "public"."orders"`), per-environment usage, credential-helper pointer (`op run`), and regeneration command.
- `.claude/skills/pg-orders/scripts/query.sh` — executable (`chmod +x`); dbname `orders` hardcoded, host/port/user baked in as fallbacks, password read at runtime from a `.env` file. Forwards every `PG[A-Z]…` libpq env var into the container.
- `.claude/skills/pg-orders/scripts/.env.example` — commented template with the generation-time `PGHOST`/`PGPORT`/`PGUSER` values pre-filled.
- `.claude/skills/pg-orders/references/tables.md` — full column listing per table with PK/FK annotations.
- `.claude/skills/pg-orders/references/relationships.md` — flat FK arrow list.
- `.claude/skills/pg-orders/references/views.md` — `public.order_summary` definition verbatim.
- `.claude/skills/pg-orders/references/indexes.md` — one schema-qualified index line per row.

No `references/enums.md` was written because the schema has no enum types. The README's schema-summary line, schema-lookup section, and "where things live" bullet were all stripped of enum references per the skeleton instructions.

**Next steps for you:**

1. Copy `scripts/.env.example` to `.env` (or per-environment `.env.dev` / `.env.prod`) at your project root and fill in real `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`. Add `/.env` and `/.env.*` to `.gitignore`.
2. Smoke-test with `bash .claude/skills/pg-orders/scripts/query.sh "SELECT 1"`.
3. Re-run `/postgres-skill-creator` whenever the schema drifts — it overwrites everything under `pg-orders/` in place.

**Caveats:**

- The container-runtime smoke test (Step 3 of the generator's own SKILL.md) was **not** run here because no Docker/Podman daemon is reachable in this sandbox and the introspection would have hit the same wall — the schema (4 tables, 1 view, no enums) was treated as the introspection result per the eval prompt's stipulation. Run the smoke test locally before relying on the generated skill in real work.
- Generation-time fallbacks in `query.sh` use plausible substituted values (`PGHOST=db.internal.example.com`, `PGPORT=5432`, `PGUSER=orders_app`) since the real libpq env vars weren't readable in this sandbox. Replace them by populating a `.env` at your project root — the runtime values from the `.env` will override the baked-in fallbacks anyway, so the substitutes only matter as a last-resort default.
