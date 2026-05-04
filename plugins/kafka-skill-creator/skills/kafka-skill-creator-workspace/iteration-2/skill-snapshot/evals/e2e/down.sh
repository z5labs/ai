#!/usr/bin/env bash
# Tear down the e2e fixture and wipe its volumes so the next `up.sh`
# starts clean.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Find the runtime that actually holds the fixture's containers. up.sh
# exports KAFKA_CONTAINER_RUNTIME for its own subshell, but that value is
# ephemeral — by the time the user runs `./down.sh` in a separate
# invocation it's gone, and a naive auto-detect on a host with both
# docker and podman would always prefer docker and target the wrong
# engine, leaving podman-managed containers and volumes in place.
#
# Ask each available runtime whether it owns the fixture's named
# container; use the one that says yes. Falls back to up.sh's preference
# order if neither holds it (fixture is already down — `down -v` is a
# no-op there, so picking either is harmless).
COMPOSE=()
for r in docker podman; do
  command -v "$r" >/dev/null 2>&1 || continue
  if "$r" ps -a --format '{{.Names}}' 2>/dev/null \
      | grep -qx kafka-skill-creator-e2e-kafka; then
    if "$r" compose version >/dev/null 2>&1; then
      COMPOSE=("$r" compose)
      echo "==> detected fixture under $r (matched container kafka-skill-creator-e2e-kafka)"
      break
    fi
  fi
done

if [ ${#COMPOSE[@]} -eq 0 ]; then
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    COMPOSE=(docker compose)
  elif command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then
    COMPOSE=(podman compose)
  else
    echo "error: need either 'docker compose' or 'podman compose' on PATH" >&2
    exit 1
  fi
fi

echo "==> bringing down ${COMPOSE[*]} (with volumes)"
"${COMPOSE[@]}" down -v
