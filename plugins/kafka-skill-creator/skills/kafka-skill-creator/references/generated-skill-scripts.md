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
      # CONTEXT_AUTH_MODE_* is the per-context auth-mode constant baked in
      # at generation time (see the readonly block further down). It must
      # NOT be redefinable from a .env file — flipping it would silently
      # bypass validate_context_env's mode-correct branch (e.g. an MTLS
      # context could end up validated as SASL with no SASL creds present,
      # or vice-versa). The readonly declaration would also catch this and
      # exit, but a named refusal here gives the operator the *why*.
      if [[ "$key" =~ ^CONTEXT_AUTH_MODE_ ]]; then
        echo "error: refusing to load $key from $file —" >&2
        echo "       auth mode is set at skill-generation time and cannot be" >&2
        echo "       overridden via .env. Re-generate the skill from a manifest" >&2
        echo "       with the desired cluster.auth value if you need to change it." >&2
        exit 1
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
  # The generator baked a `readonly CONTEXT_AUTH_MODE_<UPPER>=SASL_SCRAM|MTLS`
  # constant per declared context above this function (see "Per-context
  # static values from the manifest" below). That readonly is the source of
  # truth for which credential branch to take — NOT the kafkactl-shaped
  # CONTEXTS_<UPPER>_SASL_ENABLED / _TLS_ENABLED exports, which a .env file
  # could plausibly override and silently flip the wrapper into the wrong
  # validation/mount branch. Auth mode is a manifest-level fact frozen at
  # generation time, so it lives outside the kafkactl env surface.
  local mode_var="CONTEXT_AUTH_MODE_${upper}"
  local mode="${!mode_var:-}"
  case "$mode" in
    SASL_SCRAM)
      required+=("CONTEXTS_${upper}_SASL_USERNAME" "CONTEXTS_${upper}_SASL_PASSWORD")
      ;;
    MTLS)
      required+=(
        "CONTEXTS_${upper}_TLS_CERT"
        "CONTEXTS_${upper}_TLS_CERTKEY"
        "CONTEXTS_${upper}_TLS_CA"
      )
      ;;
    *)
      echo "error: no $mode_var baked into _common.sh — context '$context'" >&2
      echo "       isn't in the manifest the skill was generated from. Edit" >&2
      echo "       scripts/manifest.yml and re-run /kafka-skill-creator." >&2
      exit 2
      ;;
  esac
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
  if [ "$mode" = "MTLS" ]; then
    local cert_problems=() cert_var val
    for cert_var in "CONTEXTS_${upper}_TLS_CERT" "CONTEXTS_${upper}_TLS_CERTKEY" "CONTEXTS_${upper}_TLS_CA"; do
      val="${!cert_var}"
      case "$val" in
        /*) ;;
        *) cert_problems+=("$cert_var=$val (must be an absolute path; got relative)"); continue ;;
      esac
      if [ ! -f "$val" ]; then
        cert_problems+=("$cert_var=$val (file not found on host)")
        continue
      fi
      # `-r` is distinct from `-f`: a file the running process cannot read
      # (wrong owner, restrictive mode, or covered by mandatory access
      # control like SELinux at the host level) would still pass `-f` and
      # then fail opaquely inside the container. Reject unreadable cert
      # paths up front so the operator sees the problem here, not as a
      # cryptic kafkactl error from inside the container.
      [ -r "$val" ] || cert_problems+=("$cert_var=$val (file exists but is not readable by this process; check permissions/SELinux)")
    done
    if [ ${#cert_problems[@]} -gt 0 ]; then
      {
        echo "error: cert paths for context '$context' are not usable:"
        for p in "${cert_problems[@]}"; do echo "  - $p"; done
        echo "       each TLS_CERT / TLS_CERTKEY / TLS_CA must be an absolute path to a"
        echo "       file the host can read; the wrapper bind-mounts each :ro,z into the"
        echo "       kafkactl container at the same path the env var declares."
      } >&2
      exit 2
    fi
  fi
}

# For MTLS contexts, build -v <path>:<path>:ro,z args so kafkactl in the
# container can read each cert at the same path its env var declares.
# `:z` is the SELinux shared-relabel marker (ignored on systems without
# SELinux; required on Fedora/RHEL so the container's process context
# can read the bind-mounted file). Only the active context's paths reach
# the runtime — paths exported for OTHER contexts are filtered OUT by
# build_env_args's per-context-scoped forwarding pattern, so kafkactl in
# the container neither sees them as env vars nor has a mount to read
# them through. Empty under SASL_SCRAM. Sets MOUNT_ARGS for the caller.
# Reads the auth mode from the readonly CONTEXT_AUTH_MODE_<UPPER>
# constant (same source-of-truth as validate_context_env, for the same
# .env-can't-flip-it reason).
build_cert_mount_args() {
  local context="$1"
  local upper
  upper="$(printf '%s' "$context" | tr '[:lower:]' '[:upper:]')"
  upper="${upper//-/_}"
  MOUNT_ARGS=()
  local mode_var="CONTEXT_AUTH_MODE_${upper}"
  if [ "${!mode_var:-}" = "MTLS" ]; then
    local cert_var val
    for cert_var in "CONTEXTS_${upper}_TLS_CERT" "CONTEXTS_${upper}_TLS_CERTKEY" "CONTEXTS_${upper}_TLS_CA"; do
      val="${!cert_var}"
      # `:z` is the SELinux shared-relabel marker; ignored on systems
      # without SELinux. Same reason the config-mount in each wrapper
      # uses `:ro,z`: without it, Fedora/RHEL hosts hit "Permission
      # denied" reading the cert from inside the container even though
      # the file exists and is mounted.
      MOUNT_ARGS+=(-v "$val:$val:ro,z")
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
  local context="$1"
  local upper
  upper="$(printf '%s' "$context" | tr '[:lower:]' '[:upper:]')"
  upper="${upper//-/_}"
  # Per-context scoping: only forward CONTEXTS_<ACTIVE_UPPER>_* plus the
  # bare default-context shorthand (TLS_*, SASL_*, SCHEMAREGISTRY_*,
  # BROKERS). CONTEXTS_<OTHER>_* vars are dropped — kafkactl with
  # `--context <active>` only consults the active context's vars anyway,
  # so forwarding e.g. CONTEXTS_PROD_TLS_CERT to a `--context dev`
  # container would leak the prod cert path string into the container
  # for no kafkactl-level benefit.
  KAFKACTL_ENV_ARGS=()
  while IFS= read -r v; do
    [ -z "$v" ] && continue
    KAFKACTL_ENV_ARGS+=(-e "$v")
  done < <(compgen -e | grep -E "^(CONTEXTS_${upper}_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS\$)" || true)
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

# Per-context auth mode, frozen at generation time. `readonly` so a stray
# `.env` override can't flip the wrapper into the wrong validation/mount
# branch — auth mode is a manifest-level fact, not a per-environment knob.
# Lives outside the kafkactl env surface (CONTEXTS_*/TLS_*/SASL_*/...) so
# `build_env_args`'s forwarding filter never propagates it into the
# container; it's purely an internal-to-the-wrapper selector consumed by
# `validate_context_env` and `build_cert_mount_args`. `load_env_file`
# explicitly refuses any `.env` line that tries to set CONTEXT_AUTH_MODE_*.
readonly CONTEXT_AUTH_MODE_<UPPER-1>=<SASL_SCRAM|MTLS — from cluster.auth>
readonly CONTEXT_AUTH_MODE_<UPPER-2>=<SASL_SCRAM|MTLS>
# (one per declared context)

# Per-context static values from the manifest. These are the cluster-shape
# fields that don't change between environments (sasl mechanism, tls mode,
# schema registry auth) — secrets still come from .env at runtime. The
# exports flow through build_env_args's forwarding filter into the kafkactl
# container, where kafkactl's CONTEXTS_<NAME>_<FIELD> overlay populates
# the context that prepare_kafkactl_config has declared in config.yml.
#
# Generation rule: emit one block per manifest context, shape depending on
# `cluster.auth`:
#
#   SASL_SCRAM:
#     - CONTEXTS_<UPPER>_SASL_ENABLED=true
#     - CONTEXTS_<UPPER>_SASL_MECHANISM=<translated SCRAM casing>
#     - CONTEXTS_<UPPER>_TLS_ENABLED=<true|false from cluster.tls>
#
#     Translate the manifest's `sasl_mechanism` to kafkactl's casing at
#     generation time: `SCRAM-SHA-256` -> `scram-sha256`, `SCRAM-SHA-512`
#     -> `scram-sha512`. kafkactl rejects the canonical Kafka spelling
#     (`SCRAM-SHA-512`) with "Unknown sasl mechanism" — the lowercase,
#     squashed-dash form is the only one it accepts.
#
#   MTLS:
#     - CONTEXTS_<UPPER>_TLS_ENABLED=true   (mTLS implies TLS, always)
#     - NO SASL_* exports
#
#     These are kafkactl-shape: they describe cluster shape so kafkactl
#     can dial the broker correctly. They do NOT drive the wrapper's
#     branching decisions — those come from the readonly
#     `CONTEXT_AUTH_MODE_<UPPER>` constant declared just above this
#     block. validate_context_env reads CONTEXT_AUTH_MODE_<UPPER> to
#     pick which credential vars to require, and build_cert_mount_args
#     reads it to decide whether to emit -v mounts. That separation
#     is what makes the wrapper resistant to a .env file flipping
#     CONTEXTS_<UPPER>_TLS_ENABLED out from under us — auth mode is
#     a manifest fact, kafkactl exports are runtime config.
#
# Emit the SCHEMAREGISTRY_AUTH line in either case only when the manifest
# declares a `schema_registry` block.

# context: <ctx-1>  (SASL_SCRAM)
export CONTEXTS_<UPPER-1>_SASL_ENABLED=true
export CONTEXTS_<UPPER-1>_SASL_MECHANISM=<scram-sha256|scram-sha512 — see translation rule above>
export CONTEXTS_<UPPER-1>_TLS_ENABLED=<true|false>
# export CONTEXTS_<UPPER-1>_SCHEMAREGISTRY_AUTH=basic   # only when SR configured

# context: <ctx-2>  (MTLS — example shape)
# export CONTEXTS_<UPPER-2>_TLS_ENABLED=true
# # No SASL_* exports under MTLS. Cert paths come from .env at runtime.
# # export CONTEXTS_<UPPER-2>_SCHEMAREGISTRY_AUTH=basic   # only when SR configured

# context: <ctx-N>  (same shape as the manifest's auth)
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
build_env_args "$CONTEXT"
build_cert_mount_args "$CONTEXT"
prepare_kafkactl_config "$CONTEXT"

exec "$RUNTIME" run --rm -i \
  -v "$KAFKACTL_CONFIG_DIR/config.yml:/.config/kafkactl/config.yml:ro,z" \
  "${MOUNT_ARGS[@]}" \
  "${KAFKACTL_ENV_ARGS[@]}" "${EXTRA_RUNTIME_ARGS[@]}" \
  "$KAFKACTL_IMAGE" --context "$CONTEXT" \
  describe topic "$TOPIC" --output json
```

The `:z` is the SELinux shared-relabel marker; ignored on systems without SELinux. Without it, Fedora/RHEL hosts hit "Permission denied" reading the config because the host's file context isn't accessible from the container's process context.

`build_cert_mount_args` populates `MOUNT_ARGS` with `-v <path>:<path>:ro,z` entries for each cert path the active context needs (`:z` is the SELinux shared-relabel marker — required on Fedora/RHEL, ignored elsewhere; same flag the config-mount uses for the same reason). Under SASL_SCRAM the array stays empty, so the splice is a no-op; under MTLS it adds three mounts so kafkactl in the container reads each cert at the same absolute path the env var declares. Splice it BEFORE `${KAFKACTL_ENV_ARGS[@]}` rather than after — argument order doesn't matter to docker, but keeping `-v` flags adjacent makes a `docker run` printout easier to scan during incident debugging.

`require_value` and `require_eq_value` come from `_common.sh` (sourced above the flag parser); the wrapper just defines `$USAGE` so those helpers print the right shape on a missing-value error.

The other four scripts swap:

- **`describe-group.sh`** — `require_allowed "consumer-group" "$GROUP" "${ALLOWED_GROUPS[@]}"`; final exec is `... describe consumer-group "$GROUP" --output json`.
- **`lag.sh`** — same as `describe-group.sh`, but pipe the kafkactl JSON output through `jq` to extract just lag-relevant fields. Bundle a `jq` filter inline; do not run kafkactl interactively for this case (the script's stdout is the lag JSON).
- **`consume.sh`** — flag parser additionally accepts `--from-beginning`, `--max N` (translated to `--max-messages N` for kafkactl), `--partition P` (translated to `--partitions P`); `require_allowed "topic"`; final exec adds `consume "$TOPIC" --output json --exit` plus the flags above.
- **`reset-offsets.sh`** — flag parser accepts `--topic <T>`, `--to-earliest`, `--to-latest`, `--to-offset N`, `--dry-run`; `require_allowed "consumer-group"` for `$GROUP` and `require_allowed "topic"` for `$TOPIC`; final exec is `reset consumer-group-offset "$GROUP" --topic "$TOPIC" --output json` plus exactly one `--to-*` selector — `--oldest` for `--to-earliest`, `--newest` for `--to-latest`, `--offset N` for `--to-offset N`. **The group is positional, not `--group <GROUP>`.** Older drafts of this template used the flag form; kafkactl rejects it with `unknown flag: --group` on `reset consumer-group-offset` (which has aliases `cgo`, `offset`). kafkactl's default on this subcommand is dry-run; pass `--execute` to actually apply. The wrapper inverts that: pass `--execute` unless the wrapper's `--dry-run` flag is set. **Forbidden flags** (do not pass through, do not accept on the wrapper's CLI): `--allow-active-members`, `--all-topics`, `--execute-yes`. The wrapper has no `--force` / `--bypass`.
