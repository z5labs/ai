#!/usr/bin/env bash
# Bring up a Postgres container seeded by ./init.sql for the e2e eval suite.
#
# Prints `KEY=VALUE` lines on stdout that the caller can `eval`. Example:
#
#   eval "$(bash up.sh)"
#   bash ../../scripts/introspect.sh /tmp/introspect-out
#
# The container is named after the host port (`pg-skill-eval-<port>`) so
# parallel invocations on different ports don't collide. The port is chosen
# from a free ephemeral one if the caller doesn't override it.
#
# Environment overrides:
#   POSTGRES_IMAGE   — image to run (default: docker.io/library/postgres:17-alpine).
#   PG_FIXTURE_PORT  — host port to bind. Default: an ephemeral free port.
#   PG_FIXTURE_NAME  — container name. Default: pg-skill-eval-<port>.
#   PG_CONTAINER_RUNTIME — `docker` or `podman`; auto-detected by default.
#
# Constants:
#   user/password/dbname are baked at fixture-creation time. They're test-only
#   credentials, never reused outside this fixture, so naming them here keeps
#   the up/down scripts hermetic — the caller doesn't need to plumb them.

set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

POSTGRES_IMAGE="${POSTGRES_IMAGE:-docker.io/library/postgres:17-alpine}"
PG_USER="${PG_FIXTURE_USER:-evaluser}"
PG_PASSWORD="${PG_FIXTURE_PASSWORD:-evalpass}"
PG_DATABASE="${PG_FIXTURE_DATABASE:-evaldb}"

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

# Pick a free ephemeral port if the caller didn't pin one. Python's the most
# portable way that's likely to be installed alongside docker/podman; if it's
# not, we fall back to a fixed port (the container name is still derived so
# parallel runs on the same port collide loudly rather than silently).
if [ -z "${PG_FIXTURE_PORT:-}" ]; then
  if command -v python3 >/dev/null 2>&1; then
    PG_FIXTURE_PORT="$(python3 -c 'import socket;s=socket.socket();s.bind(("",0));print(s.getsockname()[1]);s.close()')"
  else
    PG_FIXTURE_PORT=55432
  fi
fi
PG_FIXTURE_NAME="${PG_FIXTURE_NAME:-pg-skill-eval-${PG_FIXTURE_PORT}}"

# If a container with this name is already running, reuse it. Lets a developer
# `bash up.sh` once and re-run the eval suite without paying the bring-up cost
# every time. Tear-down is explicit via down.sh.
if "$RUNTIME" inspect "$PG_FIXTURE_NAME" >/dev/null 2>&1; then
  echo "# reusing existing container $PG_FIXTURE_NAME" >&2
else
  "$RUNTIME" run -d --rm \
    --name "$PG_FIXTURE_NAME" \
    -e POSTGRES_USER="$PG_USER" \
    -e POSTGRES_PASSWORD="$PG_PASSWORD" \
    -e POSTGRES_DB="$PG_DATABASE" \
    -p "${PG_FIXTURE_PORT}:5432" \
    -v "${HERE}/init.sql:/docker-entrypoint-initdb.d/init.sql:ro,Z" \
    "$POSTGRES_IMAGE" \
    >/dev/null
fi

# Wait for readiness via `pg_isready` inside the container — no host-side psql
# required. The init.sql runs inline with container start; readiness is
# reported only after it completes successfully.
deadline=$(( $(date +%s) + 60 ))
while true; do
  if "$RUNTIME" exec "$PG_FIXTURE_NAME" pg_isready -U "$PG_USER" -d "$PG_DATABASE" >/dev/null 2>&1; then
    break
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    echo "fixture: pg_isready timed out after 60s" >&2
    "$RUNTIME" logs "$PG_FIXTURE_NAME" >&2 || true
    exit 1
  fi
  sleep 0.5
done

# Connect from outside the container (which is where introspect.sh's container
# also lives — both containers are on the host network, not on a shared
# user-defined network), so localhost is correct from the host's perspective.
# The generator's own container will need PG_DOCKER_ARGS=--network=host to
# resolve "localhost" to the host loopback rather than its own.
cat <<EOF
PGHOST=127.0.0.1
PGPORT=${PG_FIXTURE_PORT}
PGUSER=${PG_USER}
PGPASSWORD=${PG_PASSWORD}
PGDATABASE=${PG_DATABASE}
PG_FIXTURE_NAME=${PG_FIXTURE_NAME}
PG_DOCKER_ARGS=--network=host
EOF
