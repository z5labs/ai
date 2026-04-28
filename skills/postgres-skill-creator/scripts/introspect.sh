#!/usr/bin/env bash
# Introspect a Postgres database and write TSV files describing its schema.
#
# Usage: introspect.sh <connection-string> <output-dir>
#
# Reads PGPASSWORD from the environment. Requires `docker` or `podman` on PATH.
# Uses the alpine/psql container image for portability — the host need not have
# psql installed.

set -euo pipefail

if [ $# -ne 2 ]; then
  echo "usage: $0 <connection-string> <output-dir>" >&2
  exit 2
fi

CONN="$1"
OUT="$2"

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

mkdir -p "$OUT"

run_query() {
  local name="$1" sql="$2"
  # -A unaligned, -t tuples-only, -F $'\t' tab separator, -X no .psqlrc
  "$RUNTIME" run --rm -i -e PGPASSWORD \
    docker.io/alpine/psql "$CONN" -X -A -t -F $'\t' -c "$sql" \
    > "$OUT/$name.tsv"
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

run_query foreign_keys "
  SELECT tc.table_schema, tc.table_name, kcu.column_name,
         ccu.table_schema, ccu.table_name, ccu.column_name
  FROM information_schema.table_constraints tc
  JOIN information_schema.key_column_usage kcu
    ON tc.constraint_schema = kcu.constraint_schema
   AND tc.constraint_name = kcu.constraint_name
  JOIN information_schema.constraint_column_usage ccu
    ON tc.constraint_schema = ccu.constraint_schema
   AND tc.constraint_name = ccu.constraint_name
  WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema NOT IN ('pg_catalog', 'information_schema')
  ORDER BY tc.table_schema, tc.table_name;
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
  "$RUNTIME" run --rm -i -e PGPASSWORD \
    docker.io/alpine/psql "$CONN" -X -A -t -c \
    "SELECT pg_get_viewdef('${vschema}.${vname}'::regclass, true);" \
    > "$OUT/views/${vschema}.${vname}.sql"
done < "$OUT/views.tsv"

run_query comments "
  SELECT n.nspname, c.relname,
         COALESCE(a.attname, ''),
         d.description
  FROM pg_description d
  JOIN pg_class c ON d.objoid = c.oid
  JOIN pg_namespace n ON c.relnamespace = n.oid
  LEFT JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = d.objsubid
  WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
  ORDER BY n.nspname, c.relname;
"

echo "introspection written to $OUT" >&2
