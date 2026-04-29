# postgres-skill-creator

A Claude Code plugin that introspects a Postgres database and generates a project-level skill (`pg-<dbname>`) capturing the schema as reference docs the model can consult during ad-hoc query and exploration work.

## What it produces

After running, you get a skill at `./.claude/skills/pg-<dbname>/` containing:

- `SKILL.md` — schema-specific triggering description and a quickstart for running queries.
- `scripts/query.sh` — wrapper that runs `psql` inside a container, with the dbname baked in and host/port/user/password loaded from a `.env` file.
- `scripts/.env.example` — the keys `query.sh` understands, with the generation-time values pre-filled as commented defaults.
- `references/` — Markdown reference files describing tables, foreign-key relationships, views, indexes, and (when present) enums.

The skill is **schema-specific but environment-agnostic**: dbname is hardcoded so it always targets the same logical database, but the `.env` file determines which environment (dev / staging / prod / ...) the skill connects to at runtime.

## Installation

From a Claude Code session:

```
/plugin marketplace add z5labs/ai
/plugin install postgres-skill-creator@z5labs-ai
```

## Generating a skill

```
/postgres-skill-creator postgresql://user@host:5432/mydb
```

The connection string identifies host, port, user, and database. The **password** is read from `PGPASSWORD` for the duration of the generation run; export it before invoking the generator. The generated `query.sh` does **not** retain the password — see the [`.env` workflow](#env-per-environment-workflow) below.

Requirements:
- `docker` or `podman` on `PATH` (the plugin runs `psql` from a container — no host-side `psql` install needed).
- The host can reach the database via the connection string. On Linux when the DB is on `localhost`, set `PG_DOCKER_ARGS=--network=host` so the container shares the host network.

## Regenerating / updating an installed skill

Schemas drift. To pull in changes, re-run the generator with the same connection string:

```
/postgres-skill-creator postgresql://user@host:5432/mydb
```

This **overwrites** the existing `./.claude/skills/pg-<dbname>/` directory in place. That is intentional — stale references mislead the model. Any local edits you made to files under `pg-<dbname>/` will be lost, so keep project-specific guidance somewhere else (e.g. a top-level `CLAUDE.md` or a sibling skill).

You do **not** need to reinstall the plugin to regenerate the skill. The plugin is the generator; the generated `pg-<dbname>/` skill is the output.

## `.env` per-environment workflow

The same logical database is often deployed across multiple environments. Rather than regenerating the skill for each one (or exporting `PGHOST`/`PGUSER`/etc. by hand every session), you keep one `.env` file per environment and tell `query.sh` which one to use.

### Setup

1. Copy the example env file out of the skill, once per environment:

   ```
   cp .claude/skills/pg-mydb/scripts/.env.example .env.dev
   cp .claude/skills/pg-mydb/scripts/.env.example .env.prod
   ```

2. Fill in the real host/port/user/password for each environment. `PGDATABASE` is intentionally absent from `.env.example` — the dbname is hardcoded in `query.sh` and any value set in `.env` is overwritten at runtime.

3. Add the env files to `.gitignore` so secrets don't get committed:

   ```
   # in .gitignore
   /.env
   /.env.*
   ```

### Selecting an environment per session

`query.sh` resolves the `.env` file in this order, first match wins:

1. `--env-file PATH` CLI flag
2. `PG_ENV_FILE` environment variable
3. `./.env` in the current working directory

If none of these resolve to an existing file, `query.sh` falls back to the generation-time defaults (no error). If you explicitly point at a path that doesn't exist, `query.sh` errors out so a typo doesn't silently send you to the wrong environment.

```
# explicit per-invocation
bash .claude/skills/pg-mydb/scripts/query.sh --env-file .env.prod "SELECT 1"

# whole-shell
export PG_ENV_FILE=.env.staging
bash .claude/skills/pg-mydb/scripts/query.sh < script.sql

# default (./.env)
bash .claude/skills/pg-mydb/scripts/query.sh "SELECT 1"
```

## Runtime overrides

These environment variables apply to both the generator (`introspect.sh`) and the generated `query.sh`:

| Variable | Purpose |
|---|---|
| `PSQL_IMAGE` | Container image used to run `psql`. Override to pin a major version that matches an older server (e.g. `docker.io/alpine/psql:15`) or to pull from a private registry (e.g. `registry.internal.example.com/alpine/psql:17.7`) when `docker.io` is blocked. Default: `docker.io/alpine/psql:17.7`. |
| `PG_DOCKER_ARGS` | Extra args appended to `<runtime> run`. Common case on Linux when the DB is on `localhost`: `PG_DOCKER_ARGS=--network=host`. |
| `PG_CONTAINER_RUNTIME` | Force `docker` or `podman` rather than auto-detecting. |

If you use a private registry, authenticate (`docker login` / `podman login`) before invocation — the scripts pass the image reference through as-is.

If you set `PSQL_IMAGE` during generation, export the same value in any session that uses the generated skill so the runtime image matches the one introspection ran against.
