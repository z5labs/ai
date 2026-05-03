#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$HERE/_common.sh"

USAGE="usage: lag.sh <group> --context <ctx> [--env-file PATH]"

GROUP=""
CONTEXT=""
ENV_FILE=""
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
    -*) echo "error: unknown flag: $1" >&2; echo "$USAGE" >&2; exit 2 ;;
    *) [ -z "$GROUP" ] || { echo "error: unexpected positional: $1" >&2; echo "$USAGE" >&2; exit 2; }
       GROUP="$1"; shift ;;
  esac
done

[ -n "$GROUP" ]   || { echo "$USAGE" >&2; exit 2; }
[ -n "$CONTEXT" ] || { echo "$USAGE" >&2; exit 2; }

require_allowed "consumer-group" "$GROUP" "${ALLOWED_GROUPS[@]}"
resolve_env_file
validate_context_env "$CONTEXT"
pick_runtime
build_env_args
prepare_kafkactl_config "$CONTEXT"

"$RUNTIME" run --rm -i \
  -v "$KAFKACTL_CONFIG_DIR/config.yml:/.config/kafkactl/config.yml:ro,z" \
  "${KAFKACTL_ENV_ARGS[@]}" "${EXTRA_RUNTIME_ARGS[@]}" \
  "$KAFKACTL_IMAGE" --context "$CONTEXT" \
  describe consumer-group "$GROUP" --output json \
| jq '{group: .Group.Name, state: .State, topics: [.Topics[]? | {name: .Name, totalLag: .totalLag, partitions: [.Partitions[]? | {partition: .Partition, lag: .Lag, consumerOffset: .consumerOffset, newestOffset: .newestOffset}]}]}'
