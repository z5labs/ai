---
name: pg-warehouse
description: Run ad-hoc SQL against the warehouse Postgres database (tables: UserSessions, touchpoints, cohort_assignments, events). Use whenever the user asks to query, inspect, update, or insert data in warehouse, even if they don't name the DB explicitly.
---

This skill knows the schema of the `warehouse` Postgres database and provides a wrapper for running queries against it via a containerized `psql`.

## Running queries

Use `scripts/query.sh` to execute SQL. It loads connection details (host, port, user, password) from a `.env` file at runtime. The dbname is hardcoded - this skill is schema-specific by design, so swapping `.env` files swaps the *environment*, not the database.

### One-time setup

Copy the example env file and fill in real credentials for whichever environment you connect to:

    cp .claude/skills/pg-warehouse/scripts/.env.example .env.dev
    # edit .env.dev - set PGHOST, PGPORT, PGUSER, PGPASSWORD

Repeat for additional environments (`.env.staging`, `.env.prod`, ...). Add the chosen filename(s) to `.gitignore` so secrets don't get committed:

    /.env
    /.env.*

### Per-invocation usage

`query.sh` resolves the `.env` file in this order, first match wins:

1. `--env-file PATH` CLI flag
2. `PG_ENV_FILE` environment variable
3. `./.env` in the current working directory

If `--env-file` or `PG_ENV_FILE` is set, it must point to an existing file or the script exits with an error. If neither is set and `./.env` does not exist, the generation-time defaults are used.

    bash .claude/skills/pg-warehouse/scripts/query.sh "SELECT ..."

    bash .claude/skills/pg-warehouse/scripts/query.sh --env-file .env.prod "SELECT ..."

For multi-statement scripts, pipe via stdin:

    PG_ENV_FILE=.env.staging bash .claude/skills/pg-warehouse/scripts/query.sh < script.sql

## Schema overview

The `warehouse` database holds 4 tables across three schemas (`public`, `analytics`, `crm`) with no views and no enums. The dominant subject area is session/event analytics: `analytics."UserSessions"` is the hub table - three foreign keys point at it from `public.events` (twice, via `session_id` and `parent_session_id`) and from `crm.touchpoints` - making it the natural starting point for join paths. `analytics.cohort_assignments` rounds out the analytics schema, and `crm.touchpoints` lives alongside `public.events` as a sibling event source.

### Tables (4)

**analytics**
- `UserSessions` (12 cols, PK: id) - referenced by `public.events.session_id`, `public.events.parent_session_id`, `crm.touchpoints.session_id`
- `cohort_assignments` (6 cols, PK: id)

**crm**
- `touchpoints` (8 cols, PK: id)

**public**
- `events` (4 cols, PK: id)

## When you need details

- **Full column listing** - read `references/tables.md`
- **Foreign-key graph** - read `references/relationships.md`
- **Indexes** - read `references/indexes.md`

Before writing a non-trivial query, read the relevant reference file. The schema overview above is a map, not a substitute.

## Conventions for ad-hoc work

- For destructive statements (UPDATE, DELETE, DROP), wrap in a transaction and show the user the plan before executing: run the SELECT that identifies affected rows first, then propose the UPDATE/DELETE in a `BEGIN; ... ROLLBACK;` block so they can review.
- Prefer parameterized queries when values come from the user's prose to avoid quoting mistakes - use `psql`'s `\set` and `:'var'` interpolation.
- When a query returns more than ~50 rows, suggest a `LIMIT` or aggregation rather than dumping everything.
- Identifiers like `analytics."UserSessions"` are mixed-case and **must be double-quoted** in SQL - bare `analytics.UserSessions` will be folded to lowercase by Postgres and fail.
