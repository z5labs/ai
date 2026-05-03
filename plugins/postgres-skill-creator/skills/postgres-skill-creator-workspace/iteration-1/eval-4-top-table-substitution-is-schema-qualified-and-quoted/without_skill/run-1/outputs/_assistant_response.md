I don't have a registered `/postgres-skill-creator` slash command in this session, so I can't dispatch it as a skill. I'll proceed manually using the introspection results you provided and scaffold a Postgres-interaction skill for the `warehouse` database based on what was discovered.

## Introspection summary

Connection (from libpq env vars): `PGDATABASE=warehouse` plus the other four standard libpq variables.

Schemas, tables, and inbound foreign-key counts:

| Schema      | Table                | Cols | Inbound FKs | Notes                                                                                                       |
|-------------|----------------------|-----:|------------:|-------------------------------------------------------------------------------------------------------------|
| `public`    | `events`             | 4    | 0           | Two FK columns reference `analytics."UserSessions"` (`session_id`, `parent_session_id`).                    |
| `analytics` | `UserSessions`       | 12   | 3           | Mixed-case identifier — must be double-quoted. Referenced by `public.events` (x2) and `crm.touchpoints` (x1). |
| `analytics` | `cohort_assignments` | 6    | 0           |                                                                                                             |
| `crm`       | `touchpoints`        | 8    | 0           | `session_id` references `analytics."UserSessions"`.                                                         |

No views, no enums.

## Top table

Ranking by inbound FK count, the hub of this database is:

- **`analytics."UserSessions"`** — 3 inbound FKs

Because `UserSessions` is mixed-case, every reference to it from SQL (and from any generated docs, code, or query templates) must be **schema-qualified and double-quoted**:

```
analytics."UserSessions"
```

Unquoted (`analytics.UserSessions`) Postgres would fold to `analytics.usersessions` and the lookup would fail. The other tables (`public.events`, `analytics.cohort_assignments`, `crm.touchpoints`) are lowercase and don't strictly require quoting, but schema-qualifying them is still recommended for clarity since the database spans three schemas.

## Proposed skill scaffold

I'd create a skill at `plugins/postgres-skill-creator/skills/warehouse-db/` (or similar) with:

- `SKILL.md` — describes the `warehouse` database, lists the four tables with their schemas, column counts, and FK relationships, and instructs the model to always schema-qualify identifiers and double-quote `"UserSessions"`.
- `references/schema.md` — full table-by-table reference. Every mention of the top table rendered as `analytics."UserSessions"`.
- `references/relationships.md` — FK graph centered on `analytics."UserSessions"` showing the three inbound edges:
  - `public.events.session_id`         -> `analytics."UserSessions"`
  - `public.events.parent_session_id`  -> `analytics."UserSessions"`
  - `crm.touchpoints.session_id`       -> `analytics."UserSessions"`
- `scripts/` — small psql/SQL snippets for common joins, all using `analytics."UserSessions"` exactly.

## Example query templates (for the skill)

Top-table session lookup:

```sql
SELECT *
FROM analytics."UserSessions"
WHERE session_id = $1;
```

Join from `public.events` to the hub:

```sql
SELECT e.*, s.*
FROM public.events AS e
JOIN analytics."UserSessions" AS s
  ON s.session_id = e.session_id;
```

Three-way fan-in on the hub:

```sql
SELECT s.session_id,
       COUNT(DISTINCT e.id)  AS event_count,
       COUNT(DISTINCT t.id)  AS touchpoint_count
FROM analytics."UserSessions" AS s
LEFT JOIN public.events       AS e ON e.session_id     = s.session_id
                                   OR e.parent_session_id = s.session_id
LEFT JOIN crm.touchpoints     AS t ON t.session_id     = s.session_id
GROUP BY s.session_id;
```

Want me to go ahead and write the skill files into the `postgres-skill-creator` plugin tree, or would you prefer to review/tweak the structure first?
