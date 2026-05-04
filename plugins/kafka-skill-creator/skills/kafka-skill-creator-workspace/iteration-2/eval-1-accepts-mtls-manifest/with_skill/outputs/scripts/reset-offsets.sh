#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$HERE/_common.sh"

USAGE="usage: reset-offsets.sh <group> --topic <T> --to-earliest|--to-latest|--to-offset N [--dry-run] --context <ctx> [--env-file PATH]"

GROUP=""
TOPIC=""
CONTEXT=""
ENV_FILE=""
DRY_RUN=0
TO_EARLIEST=0
TO_LATEST=0
TO_OFFSET=""
while [ $# -gt 0 ]; do
  case "$1" in
    --context)
      require_value --context "$#" "${2:-}"
      CONTEXT="$2"; shift 2 ;;
    --context=*)
      require_eq_value "$1"
      CONTEXT="${1#--context=}"; shift ;;
    --env-file)
      require_value --env-file "$#" "${2:-}"
      ENV_FILE="$2"; shift 2 ;;
    --env-file=*)
      require_eq_value "$1"
      ENV_FILE="${1#--env-file=}"; shift ;;
    --topic)
      require_value --topic "$#" "${2:-}"
      TOPIC="$2"; shift 2 ;;
    --topic=*)
      require_eq_value "$1"
      TOPIC="${1#--topic=}"; shift ;;
    --to-earliest) TO_EARLIEST=1; shift ;;
    --to-latest)   TO_LATEST=1;   shift ;;
    --to-offset)
      require_value --to-offset "$#" "${2:-}"
      TO_OFFSET="$2"; shift 2 ;;
    --to-offset=*)
      require_eq_value "$1"
      TO_OFFSET="${1#--to-offset=}"; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    --allow-active-members|--all-topics|--execute-yes|--force|--bypass)
      echo "error: $1 is not accepted by this wrapper (non-destructive posture)." >&2
      echo "$USAGE" >&2
      exit 2 ;;
    -*) echo "error: unknown flag: $1" >&2; echo "$USAGE" >&2; exit 2 ;;
    *) [ -z "$GROUP" ] || { echo "error: unexpected positional: $1" >&2; echo "$USAGE" >&2; exit 2; }
       GROUP="$1"; shift ;;
  esac
done

[ -n "$GROUP" ]   || { echo "$USAGE" >&2; exit 2; }
[ -n "$TOPIC" ]   || { echo "$USAGE" >&2; exit 2; }
[ -n "$CONTEXT" ] || { echo "$USAGE" >&2; exit 2; }

# Exactly one of --to-earliest / --to-latest / --to-offset must be set.
SELECTOR_COUNT=0
[ "$TO_EARLIEST" = 1 ] && SELECTOR_COUNT=$((SELECTOR_COUNT+1))
[ "$TO_LATEST" = 1 ]   && SELECTOR_COUNT=$((SELECTOR_COUNT+1))
[ -n "$TO_OFFSET" ]    && SELECTOR_COUNT=$((SELECTOR_COUNT+1))
if [ "$SELECTOR_COUNT" -ne 1 ]; then
  echo "error: exactly one of --to-earliest / --to-latest / --to-offset N must be specified" >&2
  echo "$USAGE" >&2
  exit 2
fi

require_allowed "consumer-group" "$GROUP" "${ALLOWED_GROUPS[@]}"
require_allowed "topic" "$TOPIC" "${ALLOWED_TOPICS[@]}"
resolve_env_file
validate_context_env "$CONTEXT"
pick_runtime
build_env_args
build_cert_mount_args "$CONTEXT"
prepare_kafkactl_config "$CONTEXT"

# kafkactl's `reset consumer-group-offset` takes the group as a positional
# argument, NOT --group. The default behavior is dry-run; pass --execute to
# actually apply. The wrapper inverts that: pass --execute unless --dry-run.
RESET_ARGS=(reset consumer-group-offset "$GROUP" --topic "$TOPIC" --output json)
[ "$TO_EARLIEST" = 1 ] && RESET_ARGS+=(--oldest)
[ "$TO_LATEST" = 1 ]   && RESET_ARGS+=(--newest)
[ -n "$TO_OFFSET" ]    && RESET_ARGS+=(--offset "$TO_OFFSET")
[ "$DRY_RUN" = 1 ]     || RESET_ARGS+=(--execute)

exec "$RUNTIME" run --rm -i \
  -v "$KAFKACTL_CONFIG_DIR/config.yml:/.config/kafkactl/config.yml:ro,z" \
  "${MOUNT_ARGS[@]}" \
  "${KAFKACTL_ENV_ARGS[@]}" "${EXTRA_RUNTIME_ARGS[@]}" \
  "$KAFKACTL_IMAGE" --context "$CONTEXT" \
  "${RESET_ARGS[@]}"
