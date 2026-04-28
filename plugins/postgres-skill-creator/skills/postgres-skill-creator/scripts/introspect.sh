#!/usr/bin/env bash
# Introspect a Postgres database and write TSV files describing its schema.
#
# Usage: introspect.sh <connection-string> <output-dir>
#
# Reads PGPASSWORD from the environment. Requires `docker` or `podman` on PATH.
# Uses the alpine/psql container image for portability — the host need not have
# psql installed.
#
# Networking note: the connection string is interpreted *inside* the container,
# so `localhost` / `127.0.0.1` refers to the container, not the host. On Linux,
# either point the connection string at a routable host (e.g. `host.docker.internal`
# with `--add-host`) or set `PG_DOCKER_ARGS=--network=host` so the container
# shares the host network namespace.
#
# Environment overrides:
#   PSQL_IMAGE           — container image to run psql from (default: docker.io/alpine/psql:17).
#                          Override to match your server's major version if needed.
#   PG_CONTAINER_RUNTIME — `docker` or `podman` (auto-detected by default).
#   PG_DOCKER_ARGS       — extra args appended to `<runtime> run` (e.g. `--network=host`).

set -euo pipefail

if [ $# -ne 2 ]; then
  echo "usage: $0 <connection-string> <output-dir>" >&2
  exit 2
fi

CONN="$1"
OUT="$2"

: "${PGPASSWORD:?PGPASSWORD must be exported}"

PSQL_IMAGE="${PSQL_IMAGE:-docker.io/alpine/psql:17.7}"

# Word-split PG_DOCKER_ARGS into an array so users can pass multiple flags.
read -r -a EXTRA_ARGS <<< "${PG_DOCKER_ARGS:-}"

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

mkdir -p "$OUT"

run_query() {
  local name="$1" sql="$2"
  local outfile="$OUT/$name.tsv"
  # -A unaligned, -t tuples-only, -F $'\t' tab separator, -X no .psqlrc
  if "$RUNTIME" run --rm -i -e PGPASSWORD "${EXTRA_ARGS[@]}" \
    "$PSQL_IMAGE" "$CONN" -X -A -t -F $'\t' -c "$sql" \
    > "$outfile"; then
    return 0
  fi

  echo "warning: failed to introspect $name; writing empty TSV and continuing" >&2
  : > "$outfile"
}

# Replace anything outside [A-Za-z0-9_-] so the result is safe to use as a
# single path segment. Identifiers stay intact inside the SQL itself.
sanitize_filename() {
  printf '%s' "$1" | sed 's/[^A-Za-z0-9_-]/_/g'
}

run_query tables "
  SELECT c.table_schema, c.table_name, c.column_name,
         CASE WHEN c.data_type = 'USER-DEFINED' THEN c.udt_name ELSE c.data_type END,
         c.is_nullable, COALESCE(c.column_default, '')
  FROM information_schema.columns c
  JOIN information_schema.tables t
    ON t.table_schema = c.table_schema AND t.table_name = c.table_name
  WHERE c.table_schema NOT IN ('pg_catalog', 'information_schema')
    AND t.table_type = 'BASE TABLE'
  ORDER BY c.table_schema, c.table_name, c.ordinal_position;
"

run_query primary_keys "
  SELECT tc.table_schema, tc.table_name, kcu.column_name
  FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage kcu
    ON tc.constraint_schema = kcu.constraint_schema
   AND tc.constraint_name = kcu.constraint_name
  WHERE tc.constraint_type = 'PRIMARY KEY'
    AND tc.table_schema NOT IN ('pg_catalog', 'information_schema')
  ORDER BY tc.table_schema, tc.table_name, kcu.ordinal_position;
"

# Use pg_constraint with unnest WITH ORDINALITY so composite FKs pair their
# referencing/referenced columns correctly. The information_schema variant
# joins by constraint name only and produces an N×N cross-product.
run_query foreign_keys "
  SELECT src_ns.nspname, src_tbl.relname, src_col.attname,
         ref_ns.nspname, ref_tbl.relname, ref_col.attname
  FROM pg_constraint con
  JOIN pg_class src_tbl
    ON src_tbl.oid = con.conrelid
  JOIN pg_namespace src_ns
    ON src_ns.oid = src_tbl.relnamespace
  JOIN pg_class ref_tbl
    ON ref_tbl.oid = con.confrelid
  JOIN pg_namespace ref_ns
    ON ref_ns.oid = ref_tbl.relnamespace
  JOIN LATERAL unnest(con.conkey) WITH ORDINALITY AS src_key(attnum, ord)
    ON TRUE
  JOIN LATERAL unnest(con.confkey) WITH ORDINALITY AS ref_key(attnum, ord)
    ON ref_key.ord = src_key.ord
  JOIN pg_attribute src_col
    ON src_col.attrelid = src_tbl.oid
   AND src_col.attnum = src_key.attnum
  JOIN pg_attribute ref_col
    ON ref_col.attrelid = ref_tbl.oid
   AND ref_col.attnum = ref_key.attnum
  WHERE con.contype = 'f'
    AND src_ns.nspname NOT IN ('pg_catalog', 'information_schema')
  ORDER BY src_ns.nspname, src_tbl.relname, src_key.ord;
"

run_query indexes "
  SELECT schemaname, tablename, indexname, indexdef
  FROM pg_indexes
  WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
  ORDER BY schemaname, tablename, indexname;
"

run_query enums "
  SELECT n.nspname, t.typname, e.enumlabel
  FROM pg_type t
  JOIN pg_enum e ON t.oid = e.enumtypid
  JOIN pg_namespace n ON n.oid = t.typnamespace
  WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
  ORDER BY n.nspname, t.typname, e.enumsortorder;
"

# Views: list each view's schema/name in views.tsv and dump each definition to
# views/<schema>.<name>.sql. View definitions are multi-line SQL, so embedding
# them in a TSV row corrupts downstream parsing.
run_query views "
  SELECT table_schema, table_name
  FROM information_schema.views
  WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
  ORDER BY table_schema, table_name;
"
mkdir -p "$OUT/views"
while IFS=$'\t' read -r vschema vname; do
  [ -z "$vschema" ] && continue
  safe_schema="$(sanitize_filename "$vschema")"
  safe_name="$(sanitize_filename "$vname")"
  "$RUNTIME" run --rm -i -e PGPASSWORD "${EXTRA_ARGS[@]}" \
    "$PSQL_IMAGE" "$CONN" -X -A -t \
    -v vschema="$vschema" -v vname="$vname" \
    -f - <<<"SELECT pg_get_viewdef(format('%I.%I', :'vschema', :'vname')::regclass, true);" \
    > "$OUT/views/${safe_schema}.${safe_name}.sql"
done < "$OUT/views.tsv"

# Strip tabs/newlines from comment text so each row stays on one TSV line.
run_query comments "
  SELECT n.nspname, c.relname,
         COALESCE(a.attname, ''),
         regexp_replace(d.description, '[\t\n\r]', ' ', 'g')
  FROM pg_description d
  JOIN pg_class c ON d.objoid = c.oid
  JOIN pg_namespace n ON c.relnamespace = n.oid
  LEFT JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = d.objsubid
  WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
  ORDER BY n.nspname, c.relname;
"

echo "introspection written to $OUT" >&2
