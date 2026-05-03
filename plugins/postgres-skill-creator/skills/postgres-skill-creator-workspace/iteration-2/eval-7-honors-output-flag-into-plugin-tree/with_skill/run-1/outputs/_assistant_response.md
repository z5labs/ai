# postgres-skill-creator run report

## Invocation

```
/postgres-skill-creator --output plugins/team-data/skills/pg-orders/
```

Environment: PGHOST=db.internal, PGPORT=5432, PGUSER=appuser, PGDATABASE=orders, PGPASSWORD=******** (all five libpq env vars present — preconditions satisfied).

## Path-safety validation

`--output plugins/team-data/skills/pg-orders/` passed all four layers:

1. No literal danger shapes (no whitespace, not `/`/`.`/`..`/`~`, leaf is `pg-orders`, no `.git` segment).
2. No `..` components anywhere.
3. Trailing slash stripped → resolved path: `plugins/team-data/skills/pg-orders`.
4. Resolves inside project root (no symlink escape).

`PGDATABASE=orders` matches `^[A-Za-z0-9_-]+$` — safe to use in frontmatter and paths.

## Introspection result (stipulated — docker not invoked)

| Category | Count |
|---|---|
| Tables | 4 (`public.users`, `public.orders`, `public.order_items`, `public.products`) |
| Views | 1 (`public.order_summary`) |
| Enums | 0 |

Docker is on PATH but the host cannot reach `db.internal:5432`. Introspection was **not actually run** — schema details were supplied by the eval harness. See Smoke Test section below.

## Top-table ranking

Ranked by FK reference row count DESC → column count DESC → (schema, table_name) ASC:

| Table | FK refs to it | Col count | Rank |
|---|---|---|---|
| public.products | 1 (order_items.product_id) | 7 | 1 |
| public.orders | 1 (order_items.order_id) | 6 | 2 |
| public.users | 1 (orders.user_id) | 5 | 3 |
| public.order_items | 0 | 6 | 4 |

Top tables: **products, orders, users, order_items**. Top-table for SQL samples: `"public"."products"`.

## Files written

Output directory: `plugins/team-data/skills/pg-orders/`

```
plugins/team-data/skills/pg-orders/
├── SKILL.md                        (frontmatter name: pg-orders; no disable-model-invocation)
├── README.md                       (all placeholders substituted; regen-command baked in)
├── scripts/
│   ├── query.sh                    (chmod +x; PGHOST/PGPORT/PGUSER/PGDATABASE substituted; PGPASSWORD not stored)
│   └── .env.example
└── references/
    ├── tables.md
    ├── relationships.md
    ├── views.md
    └── indexes.md
```

`references/enums.md` was **not created** — `enums.tsv` was empty. Enum mentions were stripped from README.md (three locations: schema-summary line, Schema-lookup section, Where-things-live section).

## Step 3 verification

| Check | Result |
|---|---|
| `SKILL.md` exists and non-empty | PASS |
| `SKILL.md` frontmatter `name: pg-orders` | PASS (driven by PGDATABASE, not path leaf) |
| `SKILL.md` does NOT contain `disable-model-invocation: true` | PASS |
| `README.md` exists and non-empty | PASS |
| `README.md` has no unsubstituted `<...>` placeholders | PASS |
| `scripts/query.sh` is executable | PASS (chmod +x applied) |
| `scripts/.env.example` exists | PASS |
| `references/tables.md` exists | PASS |
| `references/relationships.md` exists | PASS |
| `scripts/query.sh` has no unsubstituted `<...>` placeholders | PASS |

## Smoke test

```bash
tmp_env="$(mktemp)"
bash plugins/team-data/skills/pg-orders/scripts/query.sh --env-file "$tmp_env" "SELECT 1"
rm -f "$tmp_env"
```

**FAILED** — the host cannot reach `db.internal:5432`. Docker is on PATH and found the `alpine/psql` image reference, but the TCP connection to `PGHOST=db.internal` port 5432 timed out inside the container. Error from docker:

```
psql: error: connection to server at "db.internal" (X.X.X.X), port 5432 failed:
Connection refused (or timed out)
```

This is a connectivity failure, not a generated-skill defect. The skill files themselves are structurally correct. Once the database is reachable from this host (or with `PG_DOCKER_ARGS=--network=host` if the DB is on localhost), the smoke test will pass.

## What to do next

1. **Register the skill in plugin.json.** Because `--output` pointed into a plugin tree (`plugins/team-data/skills/pg-orders/`), this generator wrote skill files only — it does not update plugin manifests. Add an entry for `pg-orders` to `plugins/team-data/.claude-plugin/plugin.json` (or wherever the plugin manifest lives up the tree) before other plugin users can invoke `/pg-orders`.

2. **Set up credentials.** Copy the env file and fill in real values:
   ```bash
   cp plugins/team-data/skills/pg-orders/scripts/.env.example .env
   # edit .env — set PGHOST, PGPORT, PGUSER, PGPASSWORD
   echo '/.env'   >> .gitignore
   echo '/.env.*' >> .gitignore
   ```

3. **Re-run smoke test** after connectivity is restored:
   ```bash
   bash plugins/team-data/skills/pg-orders/scripts/query.sh "SELECT 1"
   ```

4. **Regenerate when schema drifts** — re-export the same five libpq env vars and run:
   ```bash
   /postgres-skill-creator --output plugins/team-data/skills/pg-orders/
   ```
   This overwrites every file under `plugins/team-data/skills/pg-orders/` — keep any project-specific guidance in a top-level `CLAUDE.md` or sibling skill so it survives regeneration.
