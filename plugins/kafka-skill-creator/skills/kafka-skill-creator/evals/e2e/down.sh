#!/usr/bin/env bash
# Tear down the e2e fixture and wipe its volumes so the next `up.sh`
# starts clean.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
elif command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then
  COMPOSE=(podman compose)
else
  echo "error: need either 'docker compose' or 'podman compose' on PATH" >&2
  exit 1
fi

echo "==> bringing down ${COMPOSE[*]} (with volumes)"
"${COMPOSE[@]}" down -v
