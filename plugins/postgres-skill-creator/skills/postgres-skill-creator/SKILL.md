---
name: postgres-skill-creator
description: Introspect a Postgres database and generate a project-level skill (`pg-<dbname>`) that bakes its schema into a reference the model can consult for ad-hoc query/exploration work
disable-model-invocation: true
---

Generate a project-level skill at `./.claude/skills/pg-<dbname>/` that captures the schema of the Postgres database identified by the `PGDATABASE` environment variable. The generated skill is meant for an analyst or developer doing ad-hoc reads, one-off DML, and manual exploration — so its job is to give the model enough schema context to write correct SQL without re-introspecting on every invocation.

## Preconditions

This skill takes **no arguments**. All connection details are read from the standard libpq environment variables:

- `PGHOST`, `PGPORT`, `PGUSER`, `PGDATABASE` — routing
- `PGPASSWORD` — credential

If **any** of the five is unset, **stop immediately**. Do not prompt the user for the missing value, do not accept it inline, do not invent a placeholder. Print the full list of missing variables and tell the user to export them — or to load them from a pre-authenticated credential helper they already use (`op run --env-file=…`, `vault`, `gcloud`, a direnv-loaded `.env`, etc.) — and re-invoke the skill. Example refusal:

> The following environment variables are required but unset: `PGPASSWORD`, `PGHOST`. Export them (directly, or via a credential helper such as `op run --env-file=secrets.env -- claude …`) before re-invoking this skill.

The reason is non-negotiable: secrets must reach tools out-of-band, never through model context. If the password lands in your context — even briefly, even just to "export and forget" — it is captured in transcript, in any future compaction, and in any logs the harness keeps. The libpq env vars are how `psql` already expects credentials, so honoring them is both safer and more idiomatic than passing a connection string.

The generated `query.sh` does **not** retain the password — at runtime users supply credentials (host/port/user/password) via a `.env` file (see Step 2).

## Why this skill exists

Postgres schemas are often too large to keep in working memory, but most ad-hoc work depends on knowing column names, types, foreign keys, and which views are pre-joined. Re-introspecting at the start of every session wastes turns. The generated skill solves this by writing the schema once into a project-local reference the model loads on demand. Overwriting on regeneration is the right default — schemas drift, and the user has stated they only work with one DB at a time.

## High-level workflow

1. Detect a container runtime (`docker` or `podman`) — error out with a clear message if neither is on PATH.
2. Run introspection queries via the `alpine/psql` container against the user's database.
3. Write the generated skill to `./.claude/skills/pg-<dbname>/`, overwriting any prior version.
4. Verify the generated skill by sanity-checking that `SKILL.md` and at least one reference file are non-empty.
5. Tell the user the skill is installed and how to invoke it.

The bundled `scripts/introspect.sh` does step 1 and 2. Read it before running so you understand what queries it issues — you may need to adapt if a query fails (e.g. older Postgres versions lack a column).

If the host can't reach the database via `PGHOST` as-is (common on Linux when the DB is on `localhost` and `PGHOST=localhost` would resolve inside the container), set `PG_DOCKER_ARGS=--network=host` before invoking the script.

Override `PSQL_IMAGE` when the default `docker.io/alpine/psql` won't work — either to pin a psql major version that matches an older server, or to pull from a private registry in environments where docker.io is blocked (e.g. `PSQL_IMAGE=registry.internal.example.com/alpine/psql:17.7`). Authentication to the private registry (`docker login` / `podman login`) is the user's responsibility; the script just passes the image reference through. The generated skill's `query.sh` reads `PSQL_IMAGE` too, so users in locked-down environments should export it persistently (e.g. in their shell profile) rather than only for the generation run.

## Step 1: Run introspection

```bash
bash <skill-dir>/scripts/introspect.sh /tmp/pg-introspect-<dbname>
```

`<skill-dir>` is wherever this skill is installed (the directory containing the `SKILL.md` you are reading). Use an absolute path — don't assume a relative path resolves from the user's current working directory.

The script reads `PGHOST`/`PGPORT`/`PGUSER`/`PGDATABASE`/`PGPASSWORD` from the environment and validates them itself; it will exit non-zero with the same kind of refusal described in Preconditions if any are missing. You should still check Preconditions first so the user gets the refusal *before* the script runs, not as a script error.

The script writes one TSV per topic into the output directory:

- `tables.tsv` — schema, table, column, type, nullable, default
- `primary_keys.tsv` — schema, table, column
- `foreign_keys.tsv` — schema, table, column, ref_schema, ref_table, ref_column
- `indexes.tsv` — schema, table, index, definition
- `enums.tsv` — schema, type, label
- `views.tsv` — schema, view (one row per view; the SQL definition is written separately to `views/<schema>.<view>.sql` because multi-line SQL would corrupt the TSV)
- `comments.tsv` — schema, relation, column (may be NULL), comment

If any query fails, look at the script's error output and either fix the query inline or skip that section in the generated skill — partial coverage is better than aborting the whole generation.

## Step 2: Write the generated skill

Create these files under `./.claude/skills/pg-<dbname>/` (where `<dbname>` is the value of `PGDATABASE`). Before using `<dbname>` in the path or deleting anything, validate that it is non-empty and matches `^[A-Za-z0-9_-]+$`. If validation fails, stop and ask the user to either re-export `PGDATABASE` with a path-safe value or supply a safe override; do not delete any directory. If the validated target directory already exists, **delete it first** — overwrite is intentional, schemas drift and stale references mislead.

### `SKILL.md`

Use this skeleton; substitute the `<...>` placeholders with real content. The frontmatter `description` is what determines triggering, so make it specific to this database — name the database, mention that it's for ad-hoc reads/writes, and list the **top 3–5 prominent tables** (computed once from introspection by the deterministic rule below).

**Top-table ranking (deterministic).** Both the SKILL.md `<top tables>` list above and the README's `<top tables>` / `<top-table>` substitutions use the same ranked list. Compute it once per generation by sorting all tables in `tables.tsv` by:

1. **FK in-degree DESC** — count, from `foreign_keys.tsv`'s `(ref_schema, ref_table)` columns, how many distinct FKs point AT this table. Hub tables (the ones lots of other tables reference) rank highest because they're the most useful orientation signal for an engineer skimming the README.
2. **Column count DESC** — break in-degree ties by table width.
3. **`(schema, table_name)` ASC** — lexicographic, final tie-breaker.

Take the top 5 (or fewer if the schema has fewer than 5 tables) for `<top tables>`. The single first entry is `<top-table>`. Allowing "either column count or FK references" would let two runs over the same schema produce different output, which makes the generated skill's diff/audit story brittle — pick this rule and stick to it.

**Do NOT set `disable-model-invocation: true` on the generated skill.** The generated `pg-<dbname>` skill is meant to fire automatically when its description matches the user's prose ("how many rows in `orders` last week?", "which users haven't logged in in 30 days?"). Disabling model invocation would force the user to type `/pg-<dbname>` explicitly to use it, which defeats the point of having a per-database skill that the model recognizes by table-name context. The meta-generator (this `postgres-skill-creator`) is slash-only because *it* requires deliberate invocation; the *generated* skill should not inherit that property. Omit the field — Claude Code's default is "model-invocable" — rather than setting it to `false` (the field's name is a footgun; explicit `false` reads like an extra knob even though it's just the default).

```markdown
---
name: pg-<dbname>
description: Run ad-hoc SQL against the <dbname> Postgres database (tables: <top tables>). Use whenever the user asks to query, inspect, update, or insert data in <dbname>, even if they don't name the DB explicitly.
---

This skill knows the schema of the `<dbname>` Postgres database and provides a wrapper for running queries against it via a containerized `psql`.

## Running queries

Use `scripts/query.sh` to execute SQL. It loads connection details (host, port, user, password) from a `.env` file at runtime. The dbname is hardcoded — this skill is schema-specific by design, so swapping `.env` files swaps the *environment*, not the database.

### One-time setup

Copy the example env file and fill in real credentials for whichever environment you connect to:

    cp .claude/skills/pg-<dbname>/scripts/.env.example .env.dev
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

    bash .claude/skills/pg-<dbname>/scripts/query.sh "SELECT ..."

    bash .claude/skills/pg-<dbname>/scripts/query.sh --env-file .env.prod "SELECT ..."

For multi-statement scripts, pipe via stdin:

    PG_ENV_FILE=.env.staging bash .claude/skills/pg-<dbname>/scripts/query.sh < script.sql

## Schema overview

<One-paragraph human summary: how many tables, the dominant subject areas (inferred from table-name clusters), notable views, anything striking like a single table with 80 columns or a heavy enum-driven design.>

### Tables (<count>)

<A compact bulleted list grouped by schema. For each table, one line: `table_name (n cols, PK: ...)`. Don't dump every column here — that's what references/ is for.>

## When you need details

- **Full column listing** — read `references/tables.md`
- **Foreign-key graph** — read `references/relationships.md`
- **Views and their definitions** — read `references/views.md`
- **Indexes** — read `references/indexes.md`
- **Enums** — read `references/enums.md`

(Drop the enums bullet entirely if `enums.tsv` was empty — don't tell the model to read a file that doesn't exist.)

Before writing a non-trivial query, read the relevant reference file. The schema overview above is a map, not a substitute.

## Conventions for ad-hoc work

- For destructive statements (UPDATE, DELETE, DROP), wrap in a transaction and show the user the plan before executing: run the SELECT that identifies affected rows first, then propose the UPDATE/DELETE in a `BEGIN; ... ROLLBACK;` block so they can review.
- Prefer parameterized queries when values come from the user's prose to avoid quoting mistakes — use `psql`'s `\set` and `:'var'` interpolation.
- When a query returns more than ~50 rows, suggest a `LIMIT` or aggregation rather than dumping everything.
```

### `scripts/query.sh`

A thin wrapper. The dbname is hardcoded; host/port/user/password come from a `.env` file at runtime, with the generation-time values baked in only as fallbacks.

```bash
#!/usr/bin/env bash
set -euo pipefail

# Resolution order for the .env file, first match wins:
#   1. --env-file PATH CLI flag
#   2. PG_ENV_FILE environment variable
#   3. ./.env in the current working directory
# If --env-file or PG_ENV_FILE is set, the path must exist or the script
# exits with an error. If neither is set and ./.env doesn't exist, the
# baked-in defaults below are used.
# See .env.example for the keys this skill reads.

ENV_FILE=""
SQL_ARGS=()
while [ $# -gt 0 ]; do
  case "$1" in
    --env-file)
      if [ $# -lt 2 ] || [ -z "${2:-}" ] || [[ "${2:-}" == -* ]]; then
        echo "error: --env-file requires a non-empty PATH argument" >&2
        exit 1
      fi
      ENV_FILE="$2"
      shift 2
      ;;
    --env-file=*)
      ENV_FILE="${1#--env-file=}"
      if [ -z "$ENV_FILE" ]; then
        echo "error: --env-file requires a non-empty PATH argument" >&2
        exit 1
      fi
      shift
      ;;
    --)
      shift
      while [ $# -gt 0 ]; do SQL_ARGS+=("$1"); shift; done
      ;;
    *)
      SQL_ARGS+=("$1")
      shift
      ;;
  esac
done

if [ -z "$ENV_FILE" ]; then
  ENV_FILE="${PG_ENV_FILE:-}"
fi
if [ -z "$ENV_FILE" ] && [ -f ./.env ]; then
  ENV_FILE="./.env"
fi

# Parse KEY=VALUE lines from $ENV_FILE without sourcing it. Sourcing would
# execute arbitrary shell, which is unsafe given that ./.env is loaded
# automatically when present.
load_env_file() {
  local file="$1" line key value
  while IFS= read -r line || [ -n "$line" ]; do
    line="${line%$'\r'}"
    if [[ "$line" =~ ^[[:space:]]*(#.*)?$ ]]; then continue; fi
    if [[ "$line" =~ ^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
      key="${BASH_REMATCH[2]}"
      value="${BASH_REMATCH[3]}"
      if [[ "$value" =~ ^\"(.*)\"$ ]] || [[ "$value" =~ ^\'(.*)\'$ ]]; then
        value="${BASH_REMATCH[1]}"
      fi
      export "$key=$value"
    else
      echo "invalid env assignment in $file: $line" >&2
      exit 1
    fi
  done < "$file"
}

if [ -n "$ENV_FILE" ]; then
  if [ ! -f "$ENV_FILE" ]; then
    echo "env file not found: $ENV_FILE" >&2
    exit 1
  fi
  load_env_file "$ENV_FILE"
fi

# Fall back to generation-time values for any key the .env didn't supply.
: "${PGHOST:=<host>}"
: "${PGPORT:=<port>}"
: "${PGUSER:=<user>}"
: "${PGPASSWORD:=}"
# PGDATABASE is hardcoded — this skill is schema-specific by design.
PGDATABASE="<dbname>"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

# Forward every libpq env var (PGHOST/PGPORT/PGUSER/PGDATABASE/PGPASSWORD plus
# PGSSLMODE, PGSSLROOTCERT, PGSERVICE, PGOPTIONS, PGTARGETSESSIONATTRS,
# PGAPPNAME, PGCONNECT_TIMEOUT, …) into the container. Filter `^PG[A-Z]` so
# libpq-style names (PGFOO) pass while our internal config (PG_DOCKER_ARGS,
# PG_CONTAINER_RUNTIME, PG_ENV_FILE) does not — those start with PG_ and would
# confuse psql if forwarded.
LIBPQ_ENV_ARGS=()
while IFS= read -r var; do
  [ -z "$var" ] && continue
  LIBPQ_ENV_ARGS+=(-e "$var")
done < <(compgen -e | grep -E '^PG[A-Z]' || true)

if [ -n "${PG_CONTAINER_RUNTIME:-}" ]; then
  RUNTIME="$PG_CONTAINER_RUNTIME"
elif command -v docker >/dev/null 2>&1; then
  RUNTIME=docker
elif command -v podman >/dev/null 2>&1; then
  RUNTIME=podman
else
  echo "neither docker nor podman found on PATH" >&2
  exit 1
fi
PSQL_IMAGE="${PSQL_IMAGE:-docker.io/alpine/psql:17.7}"

read -r -a EXTRA_ARGS <<< "${PG_DOCKER_ARGS:-}"

if [ ${#SQL_ARGS[@]} -eq 0 ]; then
  exec "$RUNTIME" run --rm -i \
    "${LIBPQ_ENV_ARGS[@]}" \
    "${EXTRA_ARGS[@]}" "$PSQL_IMAGE"
else
  # Join all positional args with spaces so unquoted SQL like
  # `query.sh SELECT 1` is preserved instead of silently truncated to "SELECT".
  exec "$RUNTIME" run --rm -i \
    "${LIBPQ_ENV_ARGS[@]}" \
    "${EXTRA_ARGS[@]}" "$PSQL_IMAGE" -c "${SQL_ARGS[*]}"
fi
```

Substitute `<host>`, `<port>`, `<user>`, and `<dbname>` with the values of `PGHOST`, `PGPORT`, `PGUSER`, and `PGDATABASE` at generation time. **Never substitute `PGPASSWORD`** — the generated `query.sh` reads it from `.env` at runtime, and writing it into a file alongside the skill would re-create the secret-on-disk problem this skill is designed to avoid. `chmod +x` the script after writing.

The `LIBPQ_ENV_ARGS` block is what gives users access to the rest of libpq's connection surface (TLS settings, service files, target_session_attrs, etc.) — anything they export with a `PG[A-Z]…` name flows through to `psql` inside the container, the same way it would for a host-side `psql`.

Keep the `PSQL_IMAGE` default in `query.sh` aligned with the default in `introspect.sh` so the generated skill works out of the box without the env var set.

### `scripts/.env.example`

A commented template the user copies to a real `.env` (or per-environment `.env.dev` / `.env.prod`). Pre-fill the comments with the generation-time values so users see what the defaults are without having to inspect `query.sh`:

```
# Override per environment. Copy this file to .env (or .env.dev / .env.prod),
# uncomment the keys you want to override, and add the chosen filename(s) to
# .gitignore — these contain secrets.
#
# PGDATABASE is intentionally not listed: the dbname is hardcoded in query.sh
# and any value set here is overwritten at runtime.

# PGHOST=<host>
# PGPORT=<port>
# PGUSER=<user>
# PGPASSWORD=

# Any other libpq env var (PGSSLMODE, PGSSLROOTCERT, PGSERVICE, PGOPTIONS,
# PGTARGETSESSIONATTRS, PGAPPNAME, PGCONNECT_TIMEOUT, …) set here is also
# forwarded to psql inside the container.
```

### `README.md`

A human-facing README the engineer reads when they open the `pg-<dbname>/` directory or onboard a new teammate. SKILL.md is written for Claude when triggering; the README is for the human who's about to use the skill.

Read `references/generated-readme-skeleton.md` for the verbatim template. Substitute the `<...>` placeholders with real values from introspection at generation time:

- `<dbname>` — the value of `PGDATABASE`
- `<top tables>` — the same 3–5 ranked tables you put in the SKILL.md frontmatter description, computed once by the **Top-table ranking** rule above. Render each as a bare `table` name (the description is prose, not SQL).
- `<top-table>` — the **first** entry of `<top tables>`, formatted for SQL as a **schema-qualified, double-quoted identifier**: `"<schema>"."<table>"`. The README's row-count sample is supposed to be runnable copy-paste, so it must work for any introspected name — including non-`public` schemas, identifiers that need quoting (mixed case, spaces, reserved words like `Order`), and duplicate table names across schemas. A bare `users` would fail on `analytics."UserSessions"`-shaped tables; always-quote is safe because Postgres treats `"users"` and `users` as the same object when the stored name is lowercase.
- `<table count>`, `<view count>`, `<enum count>` — totals from introspection (drop the enums clause entirely if `enums.tsv` was empty)

The README must include working samples for `query.sh` (smoke test, real-table count, multi-statement via stdin, per-environment via `--env-file`), the one-time `.env` setup, the credential-helper pointer, and the regeneration command — those are the things engineers ask first when they open the skill directory.

### `references/tables.md`

For each table, a section like:

```markdown
## <schema>.<table>

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval(...) | PK |
| user_id | bigint | NO |  | FK → users.id |
| ... |

<Table-level COMMENT from comments.tsv, if any. Column-level comments belong in the row's `notes` cell, joined to other notes (PK, FK, unique, enum) with `; `.>
```

Group by schema. Sort tables alphabetically within a schema. Include FK arrows in the `notes` column so the model can navigate without opening `relationships.md` for simple lookups.

### `references/relationships.md`

A flat list of FKs in arrow form:

```
orders.user_id → users.id
orders.product_id → products.id
order_items.order_id → orders.id
```

This is deliberately compact — it's the file the model loads when it needs to figure out a join path.

### `references/views.md`

For each view, a section with the schema-qualified name and the SQL definition in a fenced block (read it from `views/<schema>.<view>.sql` in the introspection output). Views are often pre-joined query patterns, so their definitions are reusable templates worth showing the model verbatim.

```markdown
## <schema>.<view>

<View-level COMMENT from comments.tsv, if any.>

\`\`\`sql
<contents of views/<schema>.<view>.sql>
\`\`\`
```

### `references/indexes.md`

One line per index: `schema.tablename: indexdef` (always schema-qualified — bare table names collide across schemas). The model uses this to judge whether a WHERE clause is going to be cheap.

### `references/enums.md`

Only create this file if `enums.tsv` is non-empty.

```markdown
## <schema>.<enum_type>
- value_one
- value_two
```

## Step 3: Verify

After writing files, check:
- `.claude/skills/pg-<dbname>/SKILL.md` exists and is non-empty
- `.claude/skills/pg-<dbname>/SKILL.md`'s frontmatter does **not** contain `disable-model-invocation: true` (the generated skill must be model-invocable so it fires on natural-language prompts)
- `.claude/skills/pg-<dbname>/README.md` exists and is non-empty
- `.claude/skills/pg-<dbname>/README.md` has no unsubstituted `<...>` placeholders — `grep -E '<dbname>|<top tables?>|<top-table>|<table count>|<view count>|<enum count>'` should return nothing. The README is assembled from a placeholder-heavy template, so a generation bug can easily leave a literal like `<table count>` in the opening paragraph; verifying README existence alone won't catch that
- `.claude/skills/pg-<dbname>/scripts/query.sh` is executable
- `.claude/skills/pg-<dbname>/scripts/.env.example` exists
- At least `references/tables.md` and `references/relationships.md` exist
- `query.sh` has no unsubstituted `<...>` placeholders (the generator must have filled in `<host>`, `<port>`, `<user>`, and `<dbname>`)

Run a smoke test: `bash .claude/skills/pg-<dbname>/scripts/query.sh "SELECT 1"`. The libpq env vars (`PGHOST`/`PGPORT`/`PGUSER`/`PGDATABASE`/`PGPASSWORD`) are still exported from the invocation that triggered this skill, so the script picks them up from the environment without needing a `.env` file. If this fails, the generated skill is broken — surface the error to the user instead of claiming success.

## Step 4: Report

Tell the user:
- The path the skill was written to
- The number of tables, views, and enums captured
- To copy `scripts/.env.example` to `.env` (or per-environment `.env.dev` / `.env.prod`), fill in real credentials, and add the chosen filename(s) to `.gitignore`
- If `PSQL_IMAGE` was set during generation (private registry / pinned version), that the same value must be exported in any session that uses the generated skill
- That re-running this generator will overwrite the skill in place when the schema drifts
