Generated the `pg-warehouse` skill at `./.claude/skills/pg-warehouse/` from the introspected schema.

## What was captured

- **4 tables**, **0 views**, **0 enums**
- Schemas: `public`, `analytics`, `crm`
- Hub table: `analytics."UserSessions"` (12 cols, 3 inbound FKs from `public.events.session_id`, `public.events.parent_session_id`, `crm.touchpoints.session_id`)
- Other tables: `analytics.cohort_assignments` (6 cols), `crm.touchpoints` (8 cols), `public.events` (4 cols)

## Top-table ranking (deterministic)

Sorted by FK reference row count DESC, then column count DESC, then `(schema, table)` ASC:

1. `analytics."UserSessions"` — 3 inbound FKs, 12 cols
2. `crm.touchpoints` — 0 inbound FKs, 8 cols
3. `analytics.cohort_assignments` — 0 inbound FKs, 6 cols
4. `public.events` — 0 inbound FKs, 4 cols

`<top-table>` = `"analytics"."UserSessions"` (schema-qualified, double-quoted so the README's `count(*)` sample is runnable copy-paste even though `UserSessions` is mixed-case). The README's bash samples use single-quoted outer wrappers around the SQL so the embedded double quotes survive intact.

## Files written

- `.claude/skills/pg-warehouse/SKILL.md` — frontmatter `description` lists the four ranked tables; `disable-model-invocation` is omitted so the skill is model-invocable
- `.claude/skills/pg-warehouse/README.md` — human-facing onboarding doc, all `<...>` placeholders substituted; enum mentions stripped (3 places) since `enums.tsv` would be empty; `views.md` reference bullet dropped since there are no views
- `.claude/skills/pg-warehouse/scripts/query.sh` — wrapper with `<host>`, `<port>`, `<user>`, `<dbname>` substituted (PGPASSWORD intentionally not baked in); `chmod +x` set
- `.claude/skills/pg-warehouse/scripts/.env.example` — commented template with the generation-time defaults inline
- `.claude/skills/pg-warehouse/references/tables.md` — per-schema, per-table column listings with FK arrows in notes
- `.claude/skills/pg-warehouse/references/relationships.md` — flat FK list (3 edges, all pointing into `analytics."UserSessions".id`)
- `.claude/skills/pg-warehouse/references/indexes.md` — one line per index, schema-qualified

`references/views.md` and `references/enums.md` were intentionally not written — both topics are empty for this database.

## Note on smoke test

The instructions ask for `bash .claude/skills/pg-warehouse/scripts/query.sh "SELECT 1"` as a smoke test, but the docker/podman call would actually attempt to reach the (stipulated) database, which doesn't exist in this environment — skipping the live smoke test. All static checks pass:

- SKILL.md / README.md non-empty
- README.md has no unsubstituted `<...>` placeholders
- query.sh has no unsubstituted `<...>` placeholders and is executable
- `.env.example`, `references/tables.md`, `references/relationships.md` exist

## Next steps for the user

1. From the project root, `cp .claude/skills/pg-warehouse/scripts/.env.example .env` and fill in real `PGHOST` / `PGPORT` / `PGUSER` / `PGPASSWORD` (or use a credential helper like `op run --env-file=postgres.env -- claude`).
2. Add `/.env` and `/.env.*` to `.gitignore` so secrets aren't committed.
3. Re-run `/postgres-skill-creator` whenever the schema drifts — the generator overwrites `pg-warehouse/` in place.
