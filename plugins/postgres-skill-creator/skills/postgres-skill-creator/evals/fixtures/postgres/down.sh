#!/usr/bin/env bash
# Tear down a fixture container brought up by ./up.sh.
#
# Usage:
#   bash down.sh                    # tears down PG_FIXTURE_NAME if set, else
#                                   # every container matching pg-skill-eval-*
#   bash down.sh <container-name>   # tears down the named container

set -euo pipefail

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

if [ $# -ge 1 ]; then
  TARGETS=("$1")
elif [ -n "${PG_FIXTURE_NAME:-}" ]; then
  TARGETS=("$PG_FIXTURE_NAME")
else
  # Sweep every fixture container started by up.sh. Matches the
  # pg-skill-eval-<port> name pattern so we don't touch unrelated containers.
  mapfile -t TARGETS < <("$RUNTIME" ps -a --format '{{.Names}}' | grep -E '^pg-skill-eval-[0-9]+$' || true)
fi

if [ ${#TARGETS[@]} -eq 0 ]; then
  echo "no fixture containers to remove" >&2
  exit 0
fi

for name in "${TARGETS[@]}"; do
  "$RUNTIME" rm -f "$name" >/dev/null 2>&1 && echo "removed $name" >&2 || \
    echo "warning: failed to remove $name (already gone?)" >&2
done
