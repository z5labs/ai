#!/usr/bin/env bash
# Bring up the e2e fixture: docker compose up + wait-for-healthy + seed.
#
# Use either docker or podman compose, whichever is on PATH.
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
