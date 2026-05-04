#!/usr/bin/env bash
set -euo pipefail

# Embedded allowlists — generation-time copies of manifest.yml's topics and
# consumer_groups. Used by the wrapper scripts to refuse off-allowlist names.
# Do not hand-edit; regenerate via /kafka-skill-creator instead.
ALLOWED_TOPICS=(
  payments.orders.v1
)
ALLOWED_GROUPS=(
  payments-orders-projector
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
  # `tr` rather than `${var^^}` keeps this portable to macOS's default
  # Bash 3.2 — `^^` is Bash 4+ only.
  local upper="$(printf '%s' "$context" | tr '[:lower:]' '[:upper:]')"
  upper="${upper//-/_}"
  local required=(
    "CONTEXTS_${upper}_BROKERS"
  )
  # The generator baked one of two static-export shapes per context above
  # this function (see "Per-context static values from the manifest" below):
  #   - SASL_SCRAM contexts get CONTEXTS_<UPPER>_SASL_ENABLED=true
  #   - MTLS contexts get only CONTEXTS_<UPPER>_TLS_ENABLED=true (no SASL)
  # The presence of one or the other tells the wrapper which credential
  # vars to require at runtime. Picking off the static export instead of a
  # separate AUTH variable keeps the contract self-contained — the operator
  # never has to remember a "match the auth flag to the manifest" dance.
  local sasl_enabled_var="CONTEXTS_${upper}_SASL_ENABLED"
  local tls_enabled_var="CONTEXTS_${upper}_TLS_ENABLED"
  if [ "${!sasl_enabled_var:-}" = "true" ]; then
    required+=("CONTEXTS_${upper}_SASL_USERNAME" "CONTEXTS_${upper}_SASL_PASSWORD")
  elif [ "${!tls_enabled_var:-}" = "true" ] && [ -z "${!sasl_enabled_var:-}" ]; then
    # MTLS-only context — cert/key/CA paths come from .env at runtime.
    required+=(
      "CONTEXTS_${upper}_TLS_CERT"
      "CONTEXTS_${upper}_TLS_CERTKEY"
      "CONTEXTS_${upper}_TLS_CA"
    )
  fi
  # When the manifest declared a schema_registry block, the generator
  # baked a CONTEXTS_<UPPER>_SCHEMAREGISTRY_AUTH export above. Use its
  # presence as the signal that SR env vars should also be required for
  # this context — that way the wrapper fails closed with a complete
  # missing-list instead of letting kafkactl reach Schema Registry with
  # no URL and surface a less targeted error.
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
  # MTLS-only post-check: each cert path must be absolute and the file
  # must exist on the host. Docker bind-mount syntax requires absolute
  # paths, and a missing cert produces an opaque kafkactl error from
  # inside the container — surface both up-front in one message.
  if [ "${!tls_enabled_var:-}" = "true" ] && [ -z "${!sasl_enabled_var:-}" ]; then
    local cert_problems=() cert_var val
    for cert_var in "CONTEXTS_${upper}_TLS_CERT" "CONTEXTS_${upper}_TLS_CERTKEY" "CONTEXTS_${upper}_TLS_CA"; do
      val="${!cert_var}"
      case "$val" in
        /*) ;;
        *) cert_problems+=("$cert_var=$val (must be an absolute path; got relative)"); continue ;;
      esac
      [ -f "$val" ] || cert_problems+=("$cert_var=$val (file not found on host)")
    done
    if [ ${#cert_problems[@]} -gt 0 ]; then
      {
        echo "error: cert paths for context '$context' are not usable:"
        for p in "${cert_problems[@]}"; do echo "  - $p"; done
        echo "       each TLS_CERT / TLS_CERTKEY / TLS_CA must be an absolute path to a"
        echo "       file the host can read; the wrapper bind-mounts each :ro into the"
        echo "       kafkactl container at the same path the env var declares."
      } >&2
      exit 2
    fi
  fi
}

# For MTLS contexts, build -v <path>:<path>:ro args so kafkactl in the
# container can read each cert at the same path its env var declares.
# Only the active context's paths reach the runtime — paths exported for
# OTHER contexts are forwarded as env vars by build_env_args (the same
# CONTEXTS_*_TLS_* shape) but never get a mount, so kafkactl has no way
# to read them. Empty under SASL_SCRAM. Sets MOUNT_ARGS for the caller.
build_cert_mount_args() {
  local context="$1"
  local upper
  upper="$(printf '%s' "$context" | tr '[:lower:]' '[:upper:]')"
  upper="${upper//-/_}"
  MOUNT_ARGS=()
  local sasl_enabled_var="CONTEXTS_${upper}_SASL_ENABLED"
  local tls_enabled_var="CONTEXTS_${upper}_TLS_ENABLED"
  if [ "${!tls_enabled_var:-}" = "true" ] && [ -z "${!sasl_enabled_var:-}" ]; then
    local cert_var val
    for cert_var in "CONTEXTS_${upper}_TLS_CERT" "CONTEXTS_${upper}_TLS_CERTKEY" "CONTEXTS_${upper}_TLS_CA"; do
      val="${!cert_var}"
      MOUNT_ARGS+=(-v "$val:$val:ro")
    done
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

# kafkactl rejects `--context <name>` if the context isn't declared in
# its config file; CONTEXTS_<NAME>_* env-var overlays only modify
# already-declared contexts, they don't create them. Generate an empty-
# bodied config.yml on the fly per invocation; env vars supply the
# actual broker addresses, credentials, and SR settings at runtime.
# Sets KAFKACTL_CONFIG_DIR for the caller to mount.
prepare_kafkactl_config() {
  local context="$1"
  KAFKACTL_CONFIG_DIR="$(mktemp -d -t kafkactl-cfg-XXXXXX)"
  trap 'rm -rf -- "$KAFKACTL_CONFIG_DIR"' EXIT
  cat > "$KAFKACTL_CONFIG_DIR/config.yml" <<EOF
contexts:
  $context: {}
EOF
}

# Validate a `--flag value` pair. With `set -u`, a bare `--flag` (no value)
# would otherwise abort with an unbound-variable error from `$2`; these
# helpers turn that into a controlled exit-2 with a usage line.
# require_value:    used by callers passing `--flag value`
#   args: <flag-name> "$#" "${2:-}"
# require_eq_value: used by callers passing `--flag=value`
#   args: "$1"
# Wrappers that source this file are expected to define $USAGE before
# calling these helpers, so the error message can show the right shape.
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

# Per-context static values from the manifest. These are the cluster-shape
# fields that don't change between environments (sasl mechanism, tls mode,
# schema registry auth) — secrets still come from .env at runtime. The
# exports flow through build_env_args's forwarding filter into the kafkactl
# container, where kafkactl's CONTEXTS_<NAME>_<FIELD> overlay populates
# the context that prepare_kafkactl_config has declared in config.yml.

# context: dev  (MTLS)
export CONTEXTS_DEV_TLS_ENABLED=true
# No SASL_* exports under MTLS. Cert paths come from .env at runtime
# (CONTEXTS_DEV_TLS_CERT / CERTKEY / CA) and the wrapper bind-mounts each.
