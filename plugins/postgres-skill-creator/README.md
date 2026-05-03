# postgres-skill-creator

A Claude Code plugin that introspects a Postgres database and generates a project-level skill (`pg-<dbname>`) capturing the schema as reference docs the model can consult during ad-hoc query and exploration work.

## What it produces

After running, you get a skill at `./.claude/skills/pg-<dbname>/` containing:

- `SKILL.md` — what Claude reads when the skill triggers. Model-invocable by default (no `disable-model-invocation`) so it fires on natural-language prompts like "how many rows in `orders` last week?", not just on `/pg-<dbname>` slash invocations.
- `README.md` — human-facing quickstart with copy-paste samples for `query.sh` (smoke test, real-table count, multi-statement via stdin, per-environment via `--env-file`), the `.env`-per-environment workflow, the credential-helper pointer, and the regeneration command. This is the file an engineer reads when onboarding to a project that uses the skill.
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

All paths in this section are relative to this plugin's directory (`plugins/postgres-skill-creator/`). From a repo-root checkout, `cd plugins/postgres-skill-creator` first.

The eval suite lives under `skills/postgres-skill-creator/evals/`. It has two layers — cheap shell tests that don't need a database, and a fixture-driven suite that brings up a real Postgres in a container.

### Layer 1 — `introspect.sh` shell tests (no fixture)

```bash
bash skills/postgres-skill-creator/evals/test_introspect.sh
```

Covers the credential-routing contract from [#42](https://github.com/z5labs/ai/issues/42): `introspect.sh` accepts exactly one argument (the output dir), refuses with a single complete missing-list when any of the five required libpq env vars is unset, points users at credential helpers, and forwards the right `-e` env-var set to the runtime (libpq vars in, internal config vars out). Exits before reaching `psql`, so neither Postgres nor a container runtime is required.

### Layer 2 — fixture-driven skill evals (`evals/evals.json`)

`evals/evals.json` holds seven evals:

- **Refusal contracts** (id 0–2): the skill refuses cleanly when `PGPASSWORD` is unset, when several env vars are unset, or when a connection string is passed positionally (the `hunter2` URL test). These run cheaply against a model — no fixture needed.
- **Lightweight happy-path** (id 3–4): the skill produces a complete `pg-<dbname>/` skill from a stipulated schema, with a non-placeholder README, model-invocable frontmatter, and a deterministically-ranked schema-qualified-and-quoted top table. These also run without a fixture (the schema is described in the prompt).
- **End-to-end against real Postgres** (id 5): the skill runs against the live containerized fixture, and the grader actually executes the *generated* `query.sh` against the same fixture and verifies a known-row-count smoke test. This is the assertion that closes the manual-smoke gap deferred in #42.
- **Regeneration** (id 6): a stale `pg-evaldb/` directory with a sentinel file is pre-seeded; the skill is run again; the grader verifies the directory was wiped and rewritten, not patched.

#### Bringing the fixture up

```bash
eval "$(bash skills/postgres-skill-creator/evals/fixtures/postgres/up.sh)"
```

`up.sh` starts a `postgres:17-alpine` container seeded by `init.sql` (two schemas, five tables, a composite FK, a cross-schema FK, a view, an enum, indexes, comments) on a free ephemeral port and prints `KEY=VALUE` lines covering `PGHOST`/`PGPORT`/`PGUSER`/`PGPASSWORD`/`PGDATABASE`/`PG_FIXTURE_NAME`/`PG_DOCKER_ARGS`. `eval` exports them into the current shell. `bash down.sh` (or `bash down.sh <name>`) tears down. See `evals/fixtures/postgres/README.md` for the schema details.

#### Grading run output

`evals/grade.py` walks an iteration directory (`postgres-skill-creator-workspace/iteration-N/`) and writes `grading.json` next to each subagent's `outputs/`. Pass `--fixture-env <env-file>` so the e2e and regeneration graders can run the generated `query.sh` against the live fixture:

```bash
bash skills/postgres-skill-creator/evals/fixtures/postgres/up.sh > /tmp/pg-fixture.env
python skills/postgres-skill-creator/evals/grade.py \
  skills/postgres-skill-creator-workspace/iteration-1 \
  --fixture-env /tmp/pg-fixture.env
bash skills/postgres-skill-creator/evals/fixtures/postgres/down.sh
```

#### Running a full skill-creator iteration

The end-to-end iteration loop (spawning subagents, capturing timing, grading, generating the eval viewer for human review) runs through the [skill-creator](https://github.com/anthropics/agent-skills/tree/main/example-skills/skill-creator) skill itself rather than a single shell command — start a Claude Code session in this repo and invoke the skill-creator with a pointer to `skills/postgres-skill-creator/`. The most recent iteration workspace is committed under `skills/postgres-skill-creator-workspace/iteration-N/` for the audit trail.

Out-of-scope follow-ups: negative network-path evals ([#74](https://github.com/z5labs/ai/issues/74)), cross-PG-major-version coverage ([#75](https://github.com/z5labs/ai/issues/75)), description-optimization loop for slash-only skills ([#76](https://github.com/z5labs/ai/issues/76)).
