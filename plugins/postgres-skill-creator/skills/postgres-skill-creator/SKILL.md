---
name: postgres-skill-creator
description: Introspect a Postgres database and generate a project-level skill (`pg-<dbname>`) that bakes its schema into a reference the model can consult for ad-hoc query/exploration work
disable-model-invocation: true
argument-hint: "[connection-string]"
---

Generate a project-level skill at `./.claude/skills/pg-<dbname>/` that captures the schema of the Postgres database identified by `$ARGUMENTS[0]`. The generated skill is meant for an analyst or developer doing ad-hoc reads, one-off DML, and manual exploration — so its job is to give the model enough schema context to write correct SQL without re-introspecting on every invocation.

The connection string is the first argument. If it is missing, ask the user for it. The connection string identifies host/port/database/user, but **not the password** — the password is read from the `PGPASSWORD` environment variable. If `PGPASSWORD` is unset, prompt the user for it and export it for the duration of the run; do not write it to disk.

## Why this skill exists

Postgres schemas are often too large to keep in working memory, but most ad-hoc work depends on knowing column names, types, foreign keys, and which views are pre-joined. Re-introspecting at the start of every session wastes turns. The generated skill solves this by writing the schema once into a project-local reference the model loads on demand. Overwriting on regeneration is the right default — schemas drift, and the user has stated they only work with one DB at a time.

## High-level workflow

1. Detect a container runtime (`docker` or `podman`) — error out with a clear message if neither is on PATH.
2. Run introspection queries via the `alpine/psql` container against the user's database.
3. Write the generated skill to `./.claude/skills/pg-<dbname>/`, overwriting any prior version.
4. Verify the generated skill by sanity-checking that `SKILL.md` and at least one reference file are non-empty.
5. Tell the user the skill is installed and how to invoke it.

The bundled `scripts/introspect.sh` does step 1 and 2. Read it before running so you understand what queries it issues — you may need to adapt if a query fails (e.g. older Postgres versions lack a column).

If the host can't reach the database via the connection string as-is (common on Linux when the DB is on `localhost`), set `PG_DOCKER_ARGS=--network=host` before invoking the script.

Override `PSQL_IMAGE` when the default `docker.io/alpine/psql` won't work — either to pin a psql major version that matches an older server, or to pull from a private registry in environments where docker.io is blocked (e.g. `PSQL_IMAGE=registry.internal.example.com/alpine/psql:17.7`). Authentication to the private registry (`docker login` / `podman login`) is the user's responsibility; the script just passes the image reference through. The generated skill's `query.sh` reads `PSQL_IMAGE` too, so users in locked-down environments should export it persistently (e.g. in their shell profile) rather than only for the generation run.

## Step 1: Run introspection

```bash
bash <skill-dir>/scripts/introspect.sh "$CONN_STRING" /tmp/pg-introspect-<dbname>
```

`<skill-dir>` is wherever this skill is installed (the directory containing the `SKILL.md` you are reading). Use an absolute path — don't assume a relative path resolves from the user's current working directory.

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

Create these files under `./.claude/skills/pg-<dbname>/` (where `<dbname>` is parsed from the connection string). Before using `<dbname>` in the path or deleting anything, validate that it is non-empty and matches `^[A-Za-z0-9_-]+$`. If validation fails, stop and ask the user to confirm the database name or supply a safe override; do not delete any directory. If the validated target directory already exists, **delete it first** — overwrite is intentional, schemas drift and stale references mislead.

### `SKILL.md`

Use this skeleton; substitute the `<...>` placeholders with real content. The frontmatter `description` is what determines triggering, so make it specific to this database — name the database, mention that it's for ad-hoc reads/writes, and list a few of the most prominent table names (top 3–5 by column count or by appearing in the most foreign keys).

```markdown
---
name: pg-<dbname>
description: Run ad-hoc SQL against the <dbname> Postgres database (tables: <top tables>). Use whenever the user asks to query, inspect, update, or insert data in <dbname>, even if they don't name the DB explicitly.
---

This skill knows the schema of the `<dbname>` Postgres database and provides a wrapper for running queries against it via a containerized `psql`.

## Running queries

Use `scripts/query.sh` to execute SQL. It expects `PGPASSWORD` in the environment.

    PGPASSWORD=... bash .claude/skills/pg-<dbname>/scripts/query.sh "SELECT ..."

For multi-statement scripts, pipe via stdin:

    PGPASSWORD=... bash .claude/skills/pg-<dbname>/scripts/query.sh < script.sql

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

A thin wrapper. Bake the non-secret connection bits in; password comes from env.

```bash
#!/usr/bin/env bash
set -euo pipefail
: "${PGPASSWORD:?PGPASSWORD must be exported}"
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
CONN="<connection-string-without-password>"
if [ $# -eq 0 ]; then
  exec "$RUNTIME" run --rm -i -e PGPASSWORD "$PSQL_IMAGE" "$CONN"
else
  exec "$RUNTIME" run --rm -i -e PGPASSWORD "$PSQL_IMAGE" "$CONN" -c "$1"
fi
```

Keep the `PSQL_IMAGE` default in `query.sh` aligned with the default in `introspect.sh` so the generated skill works out of the box without the env var set.

If the user-supplied connection string had a password embedded (`postgresql://user:pw@host/db`), strip it before substituting into `CONN` — passwords belong in `PGPASSWORD`, not in a file the user might commit. `chmod +x` the script after writing.

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
- `.claude/skills/pg-<dbname>/scripts/query.sh` is executable
- At least `references/tables.md` and `references/relationships.md` exist

Run a smoke test: `bash .claude/skills/pg-<dbname>/scripts/query.sh "SELECT 1"`. If this fails, the generated skill is broken — surface the error to the user instead of claiming success.

## Step 4: Report

Tell the user:
- The path the skill was written to
- The number of tables, views, and enums captured
- That `PGPASSWORD` must be exported in any session that uses the generated skill
- If `PSQL_IMAGE` was set during generation (private registry / pinned version), that the same value must be exported in any session that uses the generated skill
- That re-running this generator will overwrite the skill in place when the schema drifts
