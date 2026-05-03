---
name: pg-evaldb
description: Run ad-hoc SQL against the evaldb Postgres database (tables: orders, users, products, events, order_items). Use whenever the user asks to query, inspect, update, or insert data in evaldb, even if they don't name the DB explicitly.
---

This skill knows the schema of the `evaldb` Postgres database and provides a wrapper for running queries against it via a containerized `psql`.

## Running queries

Use `scripts/query.sh` to execute SQL. It loads connection details (host, port, user, password) from a `.env` file at runtime. The dbname is hardcoded — this skill is schema-specific by design, so swapping `.env` files swaps the *environment*, not the database.

### One-time setup

Copy the example env file and fill in real credentials for whichever environment you connect to:

    cp .claude/skills/pg-evaldb/scripts/.env.example .env.dev
    # edit .env.dev — set PGHOST, PGPORT, PGUSER, PGPASSWORD

Repeat for additional environments (`.env.staging`, `.env.prod`, ...). Add the chosen filename(s) to `.gitignore` so secrets don't get committed:

    /.env
    /.env.*

### Per-invocation usage

`query.sh` resolves the `.env` file in this order, first match wins:

1. `--env-file PATH` CLI flag
2. `PG_ENV_FILE` environment variable
3. `./.env` in the current working directory

If `--env-file` or `PG_ENV_FILE` is set, it must point to an existing file or the script exits with an error. If neither is set and `./.env` does not exist, the generation-time defaults are used.

    bash .claude/skills/pg-evaldb/scripts/query.sh "SELECT ..."

    bash .claude/skills/pg-evaldb/scripts/query.sh --env-file .env.prod "SELECT ..."

For multi-statement scripts, pipe via stdin:

    PG_ENV_FILE=.env.staging bash .claude/skills/pg-evaldb/scripts/query.sh < script.sql

## Schema overview

A small storefront-style schema: 5 tables across two schemas (`public` for the customer-facing order pipeline — `users`, `products`, `orders`, `order_items` — and `analytics` for one event-stream table). Joins are anchored on `users` and `orders`: `orders.user_id` and `analytics.events.user_id` both reference `users.id`, and `order_items` carries a *composite* foreign key `(order_id, order_status) → orders(id, status)` that pins each line to the order's status at the time of fulfillment. Status itself is enum-driven via the `order_status` type (`pending`/`paid`/`shipped`/`cancelled`). One view, `public.active_users`, pre-aggregates `users` LEFT JOIN `orders` and is the natural starting point for "engaged user" questions.

### Tables (5)

**analytics**
- events (5 cols, PK: id)

**public**
- order_items (5 cols, PK: order_id, line_no)
- orders (4 cols, PK: id)
- products (4 cols, PK: id)
- users (4 cols, PK: id)

## When you need details

- **Full column listing** — read `references/tables.md`
- **Foreign-key graph** — read `references/relationships.md`
- **Views and their definitions** — read `references/views.md`
- **Indexes** — read `references/indexes.md`
- **Enums** — read `references/enums.md`

Before writing a non-trivial query, read the relevant reference file. The schema overview above is a map, not a substitute.

## Conventions for ad-hoc work

- For destructive statements (UPDATE, DELETE, DROP), wrap in a transaction and show the user the plan before executing: run the SELECT that identifies affected rows first, then propose the UPDATE/DELETE in a `BEGIN; ... ROLLBACK;` block so they can review.
- Prefer parameterized queries when values come from the user's prose to avoid quoting mistakes — use `psql`'s `\set` and `:'var'` interpolation.
- When a query returns more than ~50 rows, suggest a `LIMIT` or aggregation rather than dumping everything.
