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

The generator takes **no arguments**. Connection details come from the standard libpq environment variables:

| Variable | Purpose |
|---|---|
| `PGHOST` | Database host. |
| `PGPORT` | Database port. |
| `PGUSER` | Database user. |
| `PGDATABASE` | Database name. Also determines the generated skill's path (`./.claude/skills/pg-$PGDATABASE/`). |
| `PGPASSWORD` | Database password. Never seen by the model — read directly from the environment by `psql` inside the container. |

All five must be set before invocation. If any is missing, the skill stops and lists the missing variables; it will not prompt for them, accept them inline, or fall back to a default.

Beyond the five required vars, **any other libpq env var** is forwarded to `psql` inside the container — so `PGSSLMODE`, `PGSSLROOTCERT`, `PGSERVICE`, `PGOPTIONS`, `PGTARGETSESSIONATTRS`, `PGAPPNAME`, `PGCONNECT_TIMEOUT`, etc. all work as you'd expect for a host-side `psql`. The forwarding rule is "names matching `PG[A-Z]…`", which keeps libpq vars (no underscore after `PG`) flowing through while excluding this plugin's own configuration vars (`PG_DOCKER_ARGS`, `PG_CONTAINER_RUNTIME`, `PG_ENV_FILE`, `PSQL_IMAGE`).

```
/postgres-skill-creator
```

The generated `query.sh` does **not** retain the password — see the [`.env` workflow](#env-per-environment-workflow) below for how the *runtime* picks credentials up per environment.

### Pairing with a credential helper

Because the skill is agnostic to where the env vars came from, it composes naturally with any pre-authenticated tool you already use to manage secrets — 1Password CLI, HashiCorp Vault, `gcloud`, `direnv`-loaded `.env` files, and so on. For example, with the 1Password CLI:

```
op run --env-file=postgres.env -- claude
```

…where `postgres.env` declares `PGHOST=op://Vault/Item/host`, `PGPASSWORD=op://Vault/Item/password`, etc. `op run` resolves the secrets, exports them into the subprocess environment, and tears them down on exit — the model invoking `/postgres-skill-creator` only ever sees that the variables are *set*, never the values themselves. The same shape works with `vault read … | …`, `direnv exec`, `gcloud auth print-access-token`, or any helper that exposes secrets to a subprocess via env vars.

This is the recommended pattern. Exporting `PGPASSWORD` globally in your shell rc works too, but loses the boundary that a per-invocation helper provides.

### Other requirements

- `docker` or `podman` on `PATH` (the plugin runs `psql` from a container — no host-side `psql` install needed).
- The container can reach `PGHOST`. On Linux when the DB is on `localhost`, set `PG_DOCKER_ARGS=--network=host` so the container shares the host network (otherwise `localhost` resolves *inside* the container).

## Regenerating / updating an installed skill

Schemas drift. To pull in changes, re-export the same env vars and re-run the generator:

```
/postgres-skill-creator
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

## Development

Lightweight evals live under `skills/postgres-skill-creator/evals/`:

- `evals/test_introspect.sh` — shell-level tests for `introspect.sh`'s env-var validation and argument signature. Run with `bash evals/test_introspect.sh` from the skill directory; no Postgres or container runtime needed (the script exits before reaching `psql` when env vars are missing or the arg shape is wrong).
- `evals/evals.json` — skill-level eval that exercises the "refuse and instruct when env vars are missing" behavior of the SKILL.md instructions themselves.

A full end-to-end eval loop with a containerized Postgres fixture is tracked in [#61](https://github.com/z5labs/ai/issues/61).
