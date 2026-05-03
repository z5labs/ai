Regenerated `pg-evaldb` from scratch — the stale directory (with `references/_stale_marker` and the `STALE-FROM-PREVIOUS-RUN` line in `references/tables.md`) was wiped before anything was rewritten.

## What happened

1. Re-introspected the live `evaldb` Postgres via `alpine/psql` over `--network=host` (5 PG env vars + `PG_DOCKER_ARGS`).
2. Validated `PGDATABASE=evaldb` matches `^[A-Za-z0-9_-]+$`.
3. **Deleted the prior `./.claude/skills/pg-evaldb/` directory in full** — including `references/_stale_marker` and the old `references/tables.md` ending in `STALE-FROM-PREVIOUS-RUN`.
4. Wrote a fresh skill from current introspection.

## Captured

- 6 tables across 2 schemas (`public.users`, `public.products`, `public.orders`, `public.order_items`, `analytics.events`)
- 1 view (`public.active_users`) with its full SQL definition
- 1 enum (`public.order_status` — pending/paid/shipped/cancelled)
- 5 foreign keys (including the composite `order_items → orders` FK on `(order_id, order_status)`)
- 11 indexes

## Where it landed

`./.claude/skills/pg-evaldb/`:

- `SKILL.md` — frontmatter description names the database and the top-5 tables (`orders, users, products, events, order_items`), no `disable-model-invocation` so the skill fires automatically on natural-language prompts
- `README.md` — engineer-facing docs with the row-count sample using the schema-qualified, double-quoted top-table identifier `"public"."orders"`
- `scripts/query.sh` (executable) — `.env`-driven wrapper, dbname hardcoded as `evaldb`, no `PGPASSWORD` baked in
- `scripts/.env.example` — commented template with `PGHOST=127.0.0.1`, `PGPORT=46065`, `PGUSER=evaluser`
- `references/tables.md`, `relationships.md`, `views.md`, `indexes.md`, `enums.md`

## Verification

- `SKILL.md` non-empty, no `disable-model-invocation: true`
- `README.md` non-empty with no unsubstituted `<...>` placeholders
- `query.sh` is executable, no unsubstituted placeholders
- `references/_stale_marker` is gone; `references/tables.md` no longer contains `STALE-FROM-PREVIOUS-RUN`

## Next steps for you

- `cp .claude/skills/pg-evaldb/scripts/.env.example .env` (or `.env.dev`/`.env.prod`), fill in real credentials, and add `/.env` and `/.env.*` to `.gitignore`.
- Re-run `/postgres-skill-creator` again whenever the schema drifts — it will overwrite this directory in place.
