# Postgres fixture

Containerized Postgres seeded by `init.sql`, used by the postgres-skill-creator e2e eval suite.

## Quick use

```bash
# Start the fixture and capture its connection details.
eval "$(bash up.sh)"

# Run anything you'd run with libpq env vars set; PG_DOCKER_ARGS=--network=host
# is set for you so containerized psql resolves localhost to the host loopback
# rather than its own.
bash ../../../scripts/introspect.sh /tmp/introspect-out

# Tear down when finished.
bash down.sh
```

`up.sh` writes `KEY=VALUE` lines to stdout (`PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`, `PG_FIXTURE_NAME`, `PG_DOCKER_ARGS`) so it composes with `eval`. It picks a free ephemeral port by default; override with `PG_FIXTURE_PORT=…` if you want a stable one for repeated runs.

## What's in the schema

- Two schemas: `public` and `analytics`.
- Tables: `public.users`, `public.products`, `public.orders`, `public.order_items`, `analytics.events`.
- A composite foreign key (`order_items` → `orders` on `(id, status)`) so the introspection's composite-FK code path is exercised.
- A cross-schema FK (`analytics.events.user_id` → `public.users.id`).
- A view (`public.active_users`).
- An enum (`order_status`) used as a column type on both `orders` and `order_items`.
- Non-PK indexes on `orders` and `analytics.events`.
- Comments on `public.users` (table-level) and `public.users.email` (column-level).
- Seeded rows in known counts so a grader can assert on `SELECT count(*)`.

The shape is deliberately small but covers every introspection query in `scripts/introspect.sh` — adding more tables wouldn't add coverage, only runtime.

## Reusing across runs

`up.sh` uses the stable container name `pg-skill-eval` by default and skips the bring-up when a container with that name is already running — discovering its host port via `<runtime> port` so the emitted env points at the live mapping. So a developer iterating on the suite can `bash up.sh` once, run the suite many times, then `bash down.sh` at the end.

For parallel runs, set `PG_FIXTURE_NAME` to a distinct name per invocation (e.g. `PG_FIXTURE_NAME=pg-skill-eval-a` and `…-b`); each gets its own ephemeral host port and `down.sh`'s default sweep cleans them all up.

## Image and version

Default image: `docker.io/library/postgres:17-alpine`. Override via `POSTGRES_IMAGE=…`. The `init.sql` only uses features available in PG 12+; cross-version testing is tracked in [#75](https://github.com/z5labs/ai/issues/75).
