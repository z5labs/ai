#!/usr/bin/env bash
# Bring up the e2e fixture: docker compose up + wait-for-healthy + seed.
#
# Use either docker or podman compose, whichever is on PATH.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# `jq` is used to read healthcheck status out of `compose ps --format json`.
# Without it the polling loops silently see an empty status until they time
# out 2-3 minutes later, which surfaces as a misleading "did not become
# healthy" failure instead of a missing-dependency error. Fail fast.
if ! command -v jq >/dev/null 2>&1; then
  echo "error: 'jq' is required to parse compose healthcheck status" >&2
  echo "       install jq (Fedora: dnf install jq, Debian: apt install jq, macOS: brew install jq)" >&2
  exit 1
fi

# Detect compose runtime AND the underlying CLI in a single pass. Export
# KAFKA_CONTAINER_RUNTIME so seed.sh (and any downstream tool that honors it)
# uses the same one — without this, a host with both docker and podman
# installed but only `podman compose` working would have up.sh start the
# fixture under podman and seed.sh then try to talk to docker. seed.sh's
# auto-detect prefers docker when both binaries exist.
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
  export KAFKA_CONTAINER_RUNTIME=docker
elif command -v podman >/dev/null 2>&1 && podman compose version >/dev/null 2>&1; then
  COMPOSE=(podman compose)
  export KAFKA_CONTAINER_RUNTIME=podman
else
  echo "error: need either 'docker compose' or 'podman compose' on PATH" >&2
  exit 1
fi

echo "==> starting Kafka + Karapace via ${COMPOSE[*]}"
"${COMPOSE[@]}" up -d

echo "==> waiting for kafka to be healthy"
# `compose up -d` returns immediately. Poll the broker's healthcheck via
# its container status instead of duplicating the probe here.
deadline=$(( $(date +%s) + 180 ))
while :; do
  status=$(
    "${COMPOSE[@]}" ps --format json kafka 2>/dev/null \
      | (jq -r '.[0].Health // .[0].State // ""' 2>/dev/null \
         || jq -r '.Health // .State // ""' 2>/dev/null) \
      || true
  )
  case "$status" in
    healthy) echo "  kafka: healthy"; break ;;
  esac
  if [ "$(date +%s)" -gt "$deadline" ]; then
    echo "error: kafka did not become healthy within 180s" >&2
    "${COMPOSE[@]}" logs --tail=80 kafka >&2 || true
    exit 1
  fi
  sleep 2
done

echo "==> waiting for karapace to be healthy"
deadline=$(( $(date +%s) + 120 ))
while :; do
  status=$(
    "${COMPOSE[@]}" ps --format json karapace 2>/dev/null \
      | (jq -r '.[0].Health // .[0].State // ""' 2>/dev/null \
         || jq -r '.Health // .State // ""' 2>/dev/null) \
      || true
  )
  case "$status" in
    healthy) echo "  karapace: healthy"; break ;;
  esac
  if [ "$(date +%s)" -gt "$deadline" ]; then
    echo "error: karapace did not become healthy within 120s" >&2
    "${COMPOSE[@]}" logs --tail=80 karapace >&2 || true
    exit 1
  fi
  sleep 2
done

echo "==> seeding"
bash "$SCRIPT_DIR/seed.sh"

echo
echo "fixture is up."
echo "next steps:"
echo "  source $SCRIPT_DIR/env.sh"
echo "  # then run the skill against the fixture."
