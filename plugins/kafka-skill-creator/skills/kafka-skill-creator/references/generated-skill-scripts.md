# Generated-skill script templates

Verbatim templates for the scripts the generator writes into `<output>/scripts/`. Read this file when executing Step 3 of `SKILL.md` (Write the generated skill). The high-level per-script specifications (flag list, kafkactl subcommand, allowlisted flags) live in `SKILL.md`; this file only carries the bash bodies.

Two files are templated here:

- `_common.sh` — shared bootstrap sourced by every wrapper (env-file resolution, allowlist enforcement, runtime selection, env-var forwarding filter).
- `describe-topic.sh` — fully worked example of a wrapper that uses `_common.sh`. The other four wrappers (`describe-group.sh`, `lag.sh`, `consume.sh`, `reset-offsets.sh`) follow the same shape — same flag-parser skeleton, same `require_allowed` call, same `exec` of kafkactl in the container — they differ only in the kafkactl subcommand and which allowlist they check.

## `_common.sh`

Substitute the `<...>` placeholders with values from the manifest. Topic and group lists come from `topics:` and `consumer_groups:` respectively.

```bash
#!/usr/bin/env bash
set -euo pipefail

# Embedded allowlists — generation-time copies of manifest.yml's topics and
# consumer_groups. Used by the wrapper scripts to refuse off-allowlist names.
# Do not hand-edit; regenerate via /kafka-skill-creator instead.
ALLOWED_TOPICS=(
  <topic-1>
  <topic-2>
)
ALLOWED_GROUPS=(
  <group-1>
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
    "CONTEXTS_${upper}_SASL_USERNAME"
    "CONTEXTS_${upper}_SASL_PASSWORD"
  )
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
#
# Generation rule: emit one block per manifest context. Translate the
# manifest's `sasl_mechanism` value into kafkactl's casing convention
# at generation time: `SCRAM-SHA-256` -> `scram-sha256`, `SCRAM-SHA-512`
# -> `scram-sha512`. kafkactl rejects the canonical Kafka spelling
# (`SCRAM-SHA-512`) with "Unknown sasl mechanism" — the lowercase,
# squashed-dash form is the only one it accepts. TLS_ENABLED is `true`
# when cluster.tls is `required`, `false` when it is `none`. Emit the
# SCHEMAREGISTRY_AUTH line only when the manifest declares schema_registry.

# context: <ctx-1>
export CONTEXTS_<UPPER-1>_SASL_ENABLED=true
export CONTEXTS_<UPPER-1>_SASL_MECHANISM=<scram-sha256|scram-sha512 — see translation rule above>
export CONTEXTS_<UPPER-1>_TLS_ENABLED=<true|false>
# export CONTEXTS_<UPPER-1>_SCHEMAREGISTRY_AUTH=basic   # only when SR configured

# context: <ctx-2>  (same shape)
```

## `describe-topic.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$HERE/_common.sh"

USAGE="usage: describe-topic.sh <topic> --context <ctx> [--env-file PATH]"

TOPIC=""
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

exec "$RUNTIME" run --rm -i \
  -v "$KAFKACTL_CONFIG_DIR/config.yml:/.config/kafkactl/config.yml:ro,z" \
  "${KAFKACTL_ENV_ARGS[@]}" "${EXTRA_RUNTIME_ARGS[@]}" \
  "$KAFKACTL_IMAGE" --context "$CONTEXT" \
  describe topic "$TOPIC" --output json
```

The `:z` is the SELinux shared-relabel marker; ignored on systems without SELinux. Without it, Fedora/RHEL hosts hit "Permission denied" reading the config because the host's file context isn't accessible from the container's process context.

`require_value` and `require_eq_value` come from `_common.sh` (sourced above the flag parser); the wrapper just defines `$USAGE` so those helpers print the right shape on a missing-value error.

The other four scripts swap:

- **`describe-group.sh`** — `require_allowed "consumer-group" "$GROUP" "${ALLOWED_GROUPS[@]}"`; final exec is `... describe consumer-group "$GROUP" --output json`.
- **`lag.sh`** — same as `describe-group.sh`, but pipe the kafkactl JSON output through `jq` to extract just lag-relevant fields. Bundle a `jq` filter inline; do not run kafkactl interactively for this case (the script's stdout is the lag JSON).
- **`consume.sh`** — flag parser additionally accepts `--from-beginning`, `--max N` (translated to `--max-messages N` for kafkactl), `--partition P` (translated to `--partitions P`); `require_allowed "topic"`; final exec adds `consume "$TOPIC" --output json --exit` plus the flags above.
- **`reset-offsets.sh`** — flag parser accepts `--topic <T>`, `--to-earliest`, `--to-latest`, `--to-offset N`, `--dry-run`; `require_allowed "consumer-group"` for `$GROUP` and `require_allowed "topic"` for `$TOPIC`; final exec is `reset offset --group "$GROUP" --topic "$TOPIC" --output json` plus exactly one `--to-*` selector and `--dry-run` if requested. **Forbidden flags** (do not pass through, do not accept on the wrapper's CLI): `--allow-active-members`, `--all-topics`, `--execute-yes`. The wrapper has no `--force` / `--bypass`.
