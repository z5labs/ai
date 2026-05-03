#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$HERE/_common.sh"

USAGE="usage: consume.sh <topic> --context <ctx> [--from-beginning] [--max N] [--partition P] [--env-file PATH]"

TOPIC=""
CONTEXT=""
ENV_FILE=""
FROM_BEGINNING=0
MAX=""
PARTITION=""
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
    --from-beginning)
      FROM_BEGINNING=1; shift ;;
    --max)
      require_value --max "$#" "${2:-}"
      MAX="$2"; shift 2 ;;
    --max=*)
      require_eq_value "$1"
      MAX="${1#--max=}"; shift ;;
    --partition)
      require_value --partition "$#" "${2:-}"
      PARTITION="$2"; shift 2 ;;
    --partition=*)
      require_eq_value "$1"
      PARTITION="${1#--partition=}"; shift ;;
    -*) echo "error: unknown flag: $1" >&2; echo "$USAGE" >&2; exit 2 ;;
    *) [ -z "$TOPIC" ] || { echo "error: unexpected positional: $1" >&2; echo "$USAGE" >&2; exit 2; }
       TOPIC="$1"; shift ;;
  esac
done

[ -n "$TOPIC" ]   || { echo "$USAGE" >&2; exit 2; }
[ -n "$CONTEXT" ] || { echo "$USAGE" >&2; exit 2; }

require_allowed "topic" "$TOPIC" "${ALLOWED_TOPICS[@]}"
resolve_env_file
validate_context_env "$CONTEXT"
pick_runtime
build_env_args
prepare_kafkactl_config "$CONTEXT"

EXTRA=()
[ "$FROM_BEGINNING" = "1" ]  && EXTRA+=(--from-beginning)
[ -n "$MAX" ]                 && EXTRA+=(--max-messages "$MAX")
[ -n "$PARTITION" ]           && EXTRA+=(--partitions "$PARTITION")

exec "$RUNTIME" run --rm -i \
  -v "$KAFKACTL_CONFIG_DIR/config.yml:/.config/kafkactl/config.yml:ro,z" \
  "${KAFKACTL_ENV_ARGS[@]}" "${EXTRA_RUNTIME_ARGS[@]}" \
  "$KAFKACTL_IMAGE" --context "$CONTEXT" \
  consume "$TOPIC" --output json --exit "${EXTRA[@]}"
