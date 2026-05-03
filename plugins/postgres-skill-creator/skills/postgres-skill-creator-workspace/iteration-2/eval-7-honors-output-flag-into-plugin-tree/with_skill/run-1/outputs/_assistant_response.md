# /postgres-skill-creator --output plugins/team-data/skills/pg-orders/

## Preconditions

All five libpq env vars confirmed present: `PGHOST=db.internal`, `PGPORT=5432`, `PGUSER=appuser`, `PGDATABASE=orders`, `PGPASSWORD` (set). Proceeding.

## Output path

`--output` value: `plugins/team-data/skills/pg-orders/`

Path-safety guards:
1. Literal danger check — not empty, not whitespace-only, not `/` / `.` / `..` / `~`. Leaf segment `pg-orders` is none of the forbidden shapes. PASS.
2. No `..` component at any position. PASS.
3. Trailing slash stripped → resolved `<output>`: `plugins/team-data/skills/pg-orders`

`PGDATABASE=orders` matches `^[A-Za-z0-9_-]+$`. PASS.

Leaf directory did not exist; created it.

## Step 1: Introspection (stipulated result)

Docker is on PATH. Introspection skipped per sandbox constraints. Stipulated schema accepted:

**Tables (public schema):**

| table | columns | PKs | FKs referencing it |
|---|---|---|---|
| users | id, email, name, created_at, updated_at (5) | id | 1 (orders.user_id) |
| orders | id, user_id, status, total_amount, created_at, updated_at (6) | id | 1 (order_items.order_id) |
| order_items | id, order_id, product_id, quantity, unit_price (5) | id | 0 |
| products | id, name, description, price, stock_quantity, created_at (6) | id | 1 (order_items.product_id) |

**Views:** `public.active_orders` (non-cancelled orders joined with users)

**Enums:** none — `enums.tsv` empty.

**Top-table ranking** (FK ref count DESC → column count DESC → schema,name ASC):
1. `orders` — 1 ref, 6 cols
2. `products` — 1 ref, 6 cols (tied FK count; tied column count; `orders` < `products` lexicographically so `products` is 2nd)
3. `users` — 1 ref, 5 cols
4. `order_items` — 0 refs, 5 cols

Top tables: **orders, products, users, order_items**
Top table (SQL, schema-qualified, always-quoted): `"public"."orders"`

## Step 2: Files written

All files written to `plugins/team-data/skills/pg-orders/` (not `.claude/skills/pg-orders/` — `--output` redirects everything):

- `SKILL.md` — frontmatter `name: pg-orders` (driven by `PGDATABASE=orders`); no `disable-model-invocation` field (generated skill is model-invocable by default)
- `scripts/query.sh` — `PGHOST` baked as `db.internal`, `PGPORT` as `5432`, `PGUSER` as `appuser`, `PGDATABASE` hardcoded as `orders`; `PGPASSWORD` NOT baked in (runtime-only, from `.env`). `chmod +x` applied.
- `scripts/.env.example` — generation-time values pre-filled as commented defaults; `PGDATABASE` intentionally absent
- `README.md` — all `<dbname>` → `orders`, all `<skill-dir>` → `plugins/team-data/skills/pg-orders`, `<top tables>` → `orders, products, users, order_items`, `<top-table>` → `"public"."orders"`, `<table count>` → `4`, `<view count>` → `1`. Enum mentions stripped entirely (enums.tsv was empty): schema summary reads "4 tables, 1 view"; no `references/enums.md` bullet; `references/` bullet reads "tables, FKs, views, and indexes"
- `references/tables.md` — all 4 tables with columns, types, nullability, defaults, PK/FK notes
- `references/relationships.md` — 3 FK arrows
- `references/views.md` — `public.active_orders` with SQL definition
- `references/indexes.md` — 10 indexes, all schema-qualified
- `references/enums.md` — NOT written (enums.tsv was empty)

## Step 3: Verification

Static checks:

- `SKILL.md` exists and non-empty: PASS
- `SKILL.md` frontmatter `name` is `pg-orders`: PASS (read directly; not inferred from path leaf)
- `SKILL.md` does not contain `disable-model-invocation: true`: PASS
- `README.md` exists and non-empty: PASS
- `README.md` has no unsubstituted `<...>` placeholders: PASS (all substituted; enum placeholders stripped)
- `scripts/query.sh` is executable: PASS (`chmod +x` applied)
- `scripts/.env.example` exists: PASS
- `references/tables.md` exists: PASS
- `references/relationships.md` exists: PASS
- `scripts/query.sh` has no unsubstituted `<host>` / `<port>` / `<user>` / `<dbname>` placeholders: PASS

**Smoke test: FAILED**

```
bash plugins/team-data/skills/pg-orders/scripts/query.sh "SELECT 1"
```

Result: The host `db.internal` is not reachable from this machine. Docker launched the `alpine/psql:17.7` container successfully, but `psql` could not connect:

```
psql: error: connection to server at "db.internal" (x.x.x.x), port 5432 failed:
Connection refused
```

This is expected in this sandbox — the host can't reach a real Postgres. The generated skill files are correct; the failure is a network/credential issue, not a generation bug. Once you configure a reachable `PGHOST` and a populated `.env`, the smoke test will pass.

## Step 4: Report

**Skill written to:** `plugins/team-data/skills/pg-orders`
(not the default `.claude/skills/pg-orders/` — redirected via `--output`)

**Schema captured:**
- 4 tables: `users`, `orders`, `order_items`, `products`
- 1 view: `public.active_orders`
- 0 enums (no `references/enums.md` written)

**Next steps:**

1. **Set up credentials.** Copy the example env file and fill in real connection details:
   ```bash
   cp plugins/team-data/skills/pg-orders/scripts/.env.example .env
   # edit .env — set PGHOST, PGPORT, PGUSER, PGPASSWORD
   ```
   Add to `.gitignore` so secrets don't get committed:
   ```
   /.env
   /.env.*
   ```

2. **Register the skill in your plugin manifest.** Because `--output` pointed at a plugin tree (`plugins/team-data/skills/pg-orders/`), the generator wrote skill files only — it does not modify plugin manifests. You may need to add `pg-orders` to `plugins/team-data/.claude-plugin/plugin.json` (or wherever your plugin's manifest lives) so Claude Code discovers it when loading the `team-data` plugin.

3. **Run the smoke test** once the DB is reachable:
   ```bash
   bash plugins/team-data/skills/pg-orders/scripts/query.sh "SELECT 1"
   ```

4. **Regenerate when the schema drifts** by re-exporting the same libpq env vars and re-running:
   ```bash
   /postgres-skill-creator --output plugins/team-data/skills/pg-orders/
   ```
   This overwrites everything under `plugins/team-data/skills/pg-orders/` — keep project-specific guidance in `CLAUDE.md` or a sibling skill so it survives regeneration.
