# pg-evaldb

A project-local Claude Code skill that knows the schema of the `evaldb` Postgres database and provides a thin wrapper for running ad-hoc SQL against it via a containerized `psql`.

Captured at generation time:
- 5 tables (across schemas `public` and `analytics`)
- 1 view (`public.active_users`)
- 1 enum (`public.order_status`)

Top tables (by FK reference count, then column count): `users`, `orders`, `order_items`, `products`, `events`.

## Layout

```
pg-evaldb/
  SKILL.md                  # what Claude reads when this skill triggers
  README.md                 # this file (for humans)
  scripts/
    query.sh                # containerized psql wrapper
    .env.example            # template for per-environment credentials
  references/
    tables.md               # column listings + per-table notes
    relationships.md        # FK graph
    views.md                # view definitions
    indexes.md              # index definitions
    enums.md                # enum types and values
```

## One-time setup

1. Copy the example env file and fill in real credentials. Do this once per environment you connect to:

   ```
   cp .claude/skills/pg-evaldb/scripts/.env.example .env.dev
   # edit .env.dev — set PGHOST, PGPORT, PGUSER, PGPASSWORD
   ```

   Repeat for `.env.staging`, `.env.prod`, etc., as needed.

2. Add the chosen filename(s) to your repo's `.gitignore` so secrets don't get committed:

   ```
   /.env
   /.env.*
   ```

   If you'd rather not have credentials on disk at all, skip the `.env` and instead launch Claude Code from a credential helper that exports the libpq env vars for you, e.g. `op run --env-file=secrets.env -- claude`, `vault`, `gcloud`, or a direnv-loaded `.env` outside the repo. `query.sh` reads `PGHOST` / `PGPORT` / `PGUSER` / `PGPASSWORD` from the process environment when no `.env` file resolves.

## Using `query.sh`

`query.sh` resolves the `.env` file in this order, first match wins:

1. `--env-file PATH` CLI flag
2. `PG_ENV_FILE` environment variable
3. `./.env` in the current working directory

If `--env-file` or `PG_ENV_FILE` is set, the path must exist or the script exits with an error. If neither is set and `./.env` doesn't exist, the generation-time defaults baked into `query.sh` are used. `PGDATABASE` is always pinned to `evaldb` — this skill is schema-specific by design.

Smoke test:

```
bash .claude/skills/pg-evaldb/scripts/query.sh "SELECT 1"
```

Real-table count:

```
bash .claude/skills/pg-evaldb/scripts/query.sh 'SELECT count(*) FROM "public"."users"'
```

Per-environment via `--env-file`:

```
bash .claude/skills/pg-evaldb/scripts/query.sh --env-file .env.prod "SELECT now()"
```

Multi-statement script via stdin:

```
PG_ENV_FILE=.env.staging bash .claude/skills/pg-evaldb/scripts/query.sh < script.sql
```

## Networking notes

If `PGHOST` is set to `localhost` or `127.0.0.1` and the Postgres server is on the host (not inside another container), set:

```
export PG_DOCKER_ARGS=--network=host
```

so the containerized `psql` can reach the host loopback. This was the configuration in use when this skill was generated.

## Pinning the psql image

`query.sh` defaults to `docker.io/alpine/psql:17.7`. To use a private registry mirror or pin a different psql major version, export `PSQL_IMAGE` before invoking the script:

```
export PSQL_IMAGE=registry.internal.example.com/alpine/psql:17.7
```

Set it persistently (in your shell profile) if you're in an environment that always blocks docker.io.

## Regenerating

Schemas drift. When `evaldb` changes shape, re-export the libpq env vars and re-invoke the meta-generator:

```
/postgres-skill-creator
```

The generator overwrites this directory in place — that is intentional. The skill is meant to track the database, not accumulate stale references.
