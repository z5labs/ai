#!/usr/bin/env bash
set -euo pipefail

# Embedded allowlists — generation-time copies of manifest.yml's topics and
# consumer_groups. Used by the wrapper scripts to refuse off-allowlist names.
# Do not hand-edit; regenerate via /kafka-skill-creator instead.
ALLOWED_TOPICS=(
  payments.orders.v1
  payments.refunds.v1
  internal.audit.v1
)
ALLOWED_GROUPS=(
  payments-orders-projector
  payments-refunds-replayer
)

load_env_file() {
  local file="$1" line key value
  while IFS= read -r line || [ -n "$line" ]; do
    line="${line%$'\r'}"
    if [[ "$line" =~ ^[[:space:]]*(#.*)?$ ]]; then continue; fi
    if [[ "$line" =~ ^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
      key="${BASH_REMATCH[2]}"
      value="${BASH_REMATCH[3]}"
      if [[ "$value" =~ ^\"(.*)\"$ ]] || [[ "$value" =~ ^\'(.*)\'$ ]]; then
        value="${BASH_REMATCH[1]}"
      fi
      export "$key=$value"
    else
      echo "invalid env assignment in $file: $line" >&2
      exit 1
    fi
  done < "$file"
}

resolve_env_file() {
  if [ -n "${ENV_FILE:-}" ]; then :
  elif [ -n "${KAFKA_ENV_FILE:-}" ]; then ENV_FILE="$KAFKA_ENV_FILE"
  elif [ -f ./.env ]; then ENV_FILE="./.env"
  fi
  if [ -n "${ENV_FILE:-}" ]; then
    [ -f "$ENV_FILE" ] || { echo "env file not found: $ENV_FILE" >&2; exit 1; }
    load_env_file "$ENV_FILE"
  fi
}

require_allowed() {
  local kind="$1" value="$2"; shift 2
  local allowed=("$@") a
  for a in "${allowed[@]}"; do
    if [ "$a" = "$value" ]; then return 0; fi
  done
  {
    echo "error: $kind '$value' is not in this skill's allowlist."
    echo "       allowed $kind names (from manifest.yml at generation time):"
    for a in "${allowed[@]}"; do echo "         - $a"; done
    echo "       to add a name, edit the team's manifest and regenerate."
  } >&2
  exit 2
}

validate_context_env() {
  local context="$1"
  local upper="$(printf '%s' "$context" | tr '[:lower:]' '[:upper:]')"
  upper="${upper//-/_}"
  local required=(
    "CONTEXTS_${upper}_BROKERS"
    "CONTEXTS_${upper}_SASL_USERNAME"
    "CONTEXTS_${upper}_SASL_PASSWORD"
  )
  local sr_auth_var="CONTEXTS_${upper}_SCHEMAREGISTRY_AUTH"
  if [ -n "${!sr_auth_var:-}" ]; then
    required+=("CONTEXTS_${upper}_SCHEMAREGISTRY_URL")
    if [ "${!sr_auth_var}" = "basic" ]; then
      required+=("CONTEXTS_${upper}_SCHEMAREGISTRY_USERNAME")
      required+=("CONTEXTS_${upper}_SCHEMAREGISTRY_PASSWORD")
    fi
  fi
  local missing=() var
  for var in "${required[@]}"; do
    if [ -z "${!var:-}" ]; then missing+=("$var"); fi
  done
  if [ ${#missing[@]} -gt 0 ]; then
    {
      echo "error: missing required environment variables for context '$context':"
      for v in "${missing[@]}"; do echo "  - $v"; done
      echo "       export them or load them via a credential helper, then re-run."
    } >&2
    exit 2
  fi
}

pick_runtime() {
  if [ -n "${KAFKA_CONTAINER_RUNTIME:-}" ]; then RUNTIME="$KAFKA_CONTAINER_RUNTIME"
  elif command -v docker >/dev/null 2>&1; then RUNTIME=docker
  elif command -v podman >/dev/null 2>&1; then RUNTIME=podman
  else echo "neither docker nor podman found on PATH" >&2; exit 1
  fi
}

build_env_args() {
  KAFKACTL_ENV_ARGS=()
  while IFS= read -r v; do
    [ -z "$v" ] && continue
    KAFKACTL_ENV_ARGS+=(-e "$v")
  done < <(compgen -e | grep -E '^(CONTEXTS_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS$)' || true)
  read -r -a EXTRA_RUNTIME_ARGS <<< "${KAFKA_DOCKER_ARGS:-}"
}

prepare_kafkactl_config() {
  local context="$1"
  KAFKACTL_CONFIG_DIR="$(mktemp -d -t kafkactl-cfg-XXXXXX)"
  trap 'rm -rf -- "$KAFKACTL_CONFIG_DIR"' EXIT
  cat > "$KAFKACTL_CONFIG_DIR/config.yml" <<EOF
contexts:
  $context: {}
EOF
}

require_value() {
  local flag="$1" remaining="$2" value="$3"
  [ "$remaining" -ge 2 ] || { echo "error: $flag requires a value" >&2; echo "${USAGE:-}" >&2; exit 2; }
  [ -n "$value" ]        || { echo "error: $flag requires a non-empty value" >&2; echo "${USAGE:-}" >&2; exit 2; }
  case "$value" in
    -*) echo "error: $flag requires a value, got another flag: $value" >&2; echo "${USAGE:-}" >&2; exit 2 ;;
  esac
}

require_eq_value() {
  local arg="$1" value="${1#*=}"
  [ -n "$value" ] || { echo "error: ${arg%%=*} requires a non-empty value" >&2; echo "${USAGE:-}" >&2; exit 2; }
  case "$value" in
    -*) echo "error: ${arg%%=*} requires a value, got another flag: $value" >&2; echo "${USAGE:-}" >&2; exit 2 ;;
  esac
}

KAFKACTL_IMAGE="${KAFKACTL_IMAGE:-docker.io/deviceinsight/kafkactl:v5.18.0-scratch}"

# Per-context static values from the manifest. Cluster-shape fields that
# don't change between environments — secrets still come from .env at
# runtime. Forwarded by build_env_args's filter into the kafkactl container.

# context: dev
export CONTEXTS_DEV_SASL_ENABLED=true
export CONTEXTS_DEV_SASL_MECHANISM=scram-sha512
export CONTEXTS_DEV_TLS_ENABLED=false
export CONTEXTS_DEV_SCHEMAREGISTRY_AUTH=basic
