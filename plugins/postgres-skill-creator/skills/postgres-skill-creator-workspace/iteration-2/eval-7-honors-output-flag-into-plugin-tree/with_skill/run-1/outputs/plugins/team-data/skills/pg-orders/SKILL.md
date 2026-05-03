---
name: pg-orders
description: Run ad-hoc SQL against the orders Postgres database (tables: products, orders, users, order_items). Use whenever the user asks to query, inspect, update, or insert data in orders, even if they don't name the DB explicitly.
---

This skill knows the schema of the `orders` Postgres database and provides a wrapper for running queries against it via a containerized `psql`.

## Running queries

Use `scripts/query.sh` to execute SQL. It loads connection details (host, port, user, password) from a `.env` file at runtime. The dbname is hardcoded — this skill is schema-specific by design, so swapping `.env` files swaps the *environment*, not the database.

### One-time setup

Copy the example env file and fill in real credentials for whichever environment you connect to:

    cp plugins/team-data/skills/pg-orders/scripts/.env.example .env.dev
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

    bash plugins/team-data/skills/pg-orders/scripts/query.sh "SELECT ..."

    bash plugins/team-data/skills/pg-orders/scripts/query.sh --env-file .env.prod "SELECT ..."

For multi-statement scripts, pipe via stdin:

    PG_ENV_FILE=.env.staging bash plugins/team-data/skills/pg-orders/scripts/query.sh < script.sql

## Schema overview

The `orders` database is an e-commerce order-management schema comprising 4 tables and 1 view in the `public` schema. The design is centered on a classic order-line-item pattern: `users` place `orders`, each `order` contains one or more `order_items`, and each `order_item` references a `product` from the `products` catalog. The `order_summary` view provides a pre-joined roll-up of per-order totals, making it convenient for reporting queries. There are no enum types — status and similar categorical fields are modeled as text or smallint.

### Tables (4)

**public**

- `order_items` (6 cols, PK: id)
- `orders` (6 cols, PK: id)
- `products` (7 cols, PK: id)
- `users` (5 cols, PK: id)

## When you need details

- **Full column listing** — read `references/tables.md`
- **Foreign-key graph** — read `references/relationships.md`
- **Views and their definitions** — read `references/views.md`
- **Indexes** — read `references/indexes.md`

Before writing a non-trivial query, read the relevant reference file. The schema overview above is a map, not a substitute.

## Conventions for ad-hoc work

- For destructive statements (UPDATE, DELETE, DROP), wrap in a transaction and show the user the plan before executing: run the SELECT that identifies affected rows first, then propose the UPDATE/DELETE in a `BEGIN; ... ROLLBACK;` block so they can review.
- Prefer parameterized queries when values come from the user's prose to avoid quoting mistakes — use `psql`'s `\set` and `:'var'` interpolation.
- When a query returns more than ~50 rows, suggest a `LIMIT` or aggregation rather than dumping everything.
