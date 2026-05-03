#!/usr/bin/env bash
# Introspect a Kafka cluster (via kafkactl) and write JSON files describing
# its topics, consumer groups, and brokers as scoped to one team's manifest.
#
# Usage:
#   introspect.sh \
#     --context <NAME> \
#     --topic <T> [--topic <T> ...] \
#     --group <G> [--group <G> ...] \
#     <output-dir>
#
# Reads connection details from env vars matching kafkactl's documented
# CONTEXTS_<NAME>_* convention. Three keys per context are required:
#
#   CONTEXTS_<NAME>_BROKERS                  (whitespace-separated host:port list)
#   CONTEXTS_<NAME>_SASL_USERNAME
#   CONTEXTS_<NAME>_SASL_PASSWORD
#
# One key is optional (default lives in the generated kafkactl-config.yml):
#
#   CONTEXTS_<NAME>_SASL_MECHANISM
#
# Schema Registry pulls are handled by the caller (SKILL.md), not here, so the
# script doesn't need to know whether the manifest has an SR block.
#
# Requires `docker` or `podman` on PATH. Uses the deviceinsight/kafkactl
# container so the host need not have kafkactl installed.
#
# Environment overrides:
#   KAFKACTL_IMAGE         — container image to run kafkactl from.
#                            Default: docker.io/deviceinsight/kafkactl:v5.18.0-scratch.
#                            Pinned for reproducibility — bumping the default is
#                            a deliberate change, not an automatic floating tag.
#   KAFKA_CONTAINER_RUNTIME — `docker` or `podman` (auto-detected by default).
#   KAFKA_DOCKER_ARGS       — extra args appended to `<runtime> run`
#                            (e.g. `--network=host` on Linux when brokers are
#                            on localhost; the container otherwise resolves
#                            `localhost` to itself).

set -euo pipefail

usage() {
  cat <<'EOF' >&2
usage: introspect.sh --context <NAME> [--topic <T>]... [--group <G>]... <output-dir>

Required:
  --context NAME    The kafkactl context to introspect against. Must match
                    one of the manifest's contexts[].name values, uppercased
                    for env-var lookup (e.g. `--context dev` reads
                    CONTEXTS_DEV_BROKERS).

  <output-dir>      Where to write JSON dumps. Created if missing.

Repeatable:
  --topic T         Topic to describe. Pass once per topic in the manifest.
  --group G         Consumer group to describe. Pass once per group in the
                    manifest.

Connection details come from environment variables (CONTEXTS_<NAME>_BROKERS,
CONTEXTS_<NAME>_SASL_USERNAME, CONTEXTS_<NAME>_SASL_PASSWORD), never from
positional arguments — credentials must reach this script out-of-band so they
do not pass through model context. A connection string passed as a positional
arg is rejected.
EOF
}

# Reject anything that looks like a credential-bearing URL or `user:pass@host`
# handed in as a positional. We don't want to be the script that quietly
# accepts `kafka://user:pass@host:9092` (or its scheme-less cousin
# `user:pass@host:9092`) and routes the password through the model's
# transcript on its way to the container.
#
# The patterns intentionally exclude the manifest's safe topic/group charset
# (`[A-Za-z0-9._-]+`) so realistic Kafka identifiers pass through. Anything
# containing `://` (a URI scheme) or both `:` and `@` (userinfo shape) hits
# the refusal.
reject_positional_dsn() {
  local arg="$1"
  case "$arg" in
    *://*|*:*@*)
      echo "error: connection strings are not accepted as arguments." >&2
      echo "       this script reads connection details from CONTEXTS_<NAME>_* env vars." >&2
      echo "       see the manifest's contexts[].name values and the .env-per-environment workflow." >&2
      exit 2
      ;;
  esac
}

CONTEXT=""
TOPICS=()
GROUPS=()
POSITIONAL=()

while [ $# -gt 0 ]; do
  case "$1" in
    --context)
      [ $# -ge 2 ] || { echo "error: --context requires a value" >&2; exit 2; }
      CONTEXT="$2"
      shift 2
      ;;
    --context=*)
      CONTEXT="${1#--context=}"
      shift
      ;;
    --topic)
      [ $# -ge 2 ] || { echo "error: --topic requires a value" >&2; exit 2; }
      reject_positional_dsn "$2"
      TOPICS+=("$2")
      shift 2
      ;;
    --topic=*)
      _v="${1#--topic=}"
      reject_positional_dsn "$_v"
      TOPICS+=("$_v")
      shift
      ;;
    --group)
      [ $# -ge 2 ] || { echo "error: --group requires a value" >&2; exit 2; }
      reject_positional_dsn "$2"
      GROUPS+=("$2")
      shift 2
      ;;
    --group=*)
      _v="${1#--group=}"
      reject_positional_dsn "$_v"
      GROUPS+=("$_v")
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      shift
      while [ $# -gt 0 ]; do
        reject_positional_dsn "$1"
        POSITIONAL+=("$1")
        shift
      done
      ;;
    -*)
      echo "error: unknown flag: $1" >&2
      usage
      exit 2
      ;;
    *)
      reject_positional_dsn "$1"
      POSITIONAL+=("$1")
      shift
      ;;
  esac
done

if [ -z "$CONTEXT" ] || [ ${#POSITIONAL[@]} -ne 1 ]; then
  usage
  exit 2
fi

# Context name flows into env-var lookups (CONTEXTS_<NAME>_*) and into kafkactl
# `--context <NAME>`. Restrict to a charset that survives uppercasing without
# corrupting the env-var derivation.
if ! [[ "$CONTEXT" =~ ^[A-Za-z][A-Za-z0-9_-]*$ ]]; then
  echo "error: --context value '$CONTEXT' is not a valid context name." >&2
  echo "       expected ^[A-Za-z][A-Za-z0-9_-]*$ — letters, digits, underscores, hyphens; first char a letter." >&2
  exit 2
fi

OUT="${POSITIONAL[0]}"

# Derive the env-var prefix kafkactl uses for the chosen context. kafkactl
# uppercases the context name and prepends `CONTEXTS_`. Hyphens become
# underscores per kafkactl's own normalization.
ctx_upper="${CONTEXT^^}"
ctx_upper="${ctx_upper//-/_}"
prefix="CONTEXTS_${ctx_upper}"

# Validate every credential-bearing env var up front so the caller gets a
# single complete missing-list rather than discovering them one kafkactl
# failure at a time. Same shape postgres-skill-creator landed in issue #42.
required=(
  "${prefix}_BROKERS"
  "${prefix}_SASL_USERNAME"
  "${prefix}_SASL_PASSWORD"
)
missing=()
for var in "${required[@]}"; do
  if [ -z "${!var:-}" ]; then
    missing+=("$var")
  fi
done
if [ ${#missing[@]} -gt 0 ]; then
  {
    echo "error: missing required environment variables for context '$CONTEXT':"
    for v in "${missing[@]}"; do echo "  - $v"; done
    echo
    echo "export them (or load them from a credential helper such as op run, vault,"
    echo "direnv, gcloud) before re-running. these names follow kafkactl's documented"
    echo "CONTEXTS_<NAME>_* convention; see references/kafkactl-env-vars.md."
  } >&2
  exit 2
fi

KAFKACTL_IMAGE="${KAFKACTL_IMAGE:-docker.io/deviceinsight/kafkactl:v5.18.0-scratch}"

# Word-split KAFKA_DOCKER_ARGS into an array so callers can pass multiple flags.
read -r -a EXTRA_ARGS <<< "${KAFKA_DOCKER_ARGS:-}"

# Forward kafkactl-relevant env vars into the container. The filter is the
# union of kafkactl's documented prefixes (CONTEXTS_, TLS_, SASL_,
# SCHEMAREGISTRY_) plus the bare BROKERS shorthand kafkactl honors for the
# default context. Internal config (KAFKA_DOCKER_ARGS, KAFKA_CONTAINER_RUNTIME,
# KAFKACTL_IMAGE) starts with KAFKA_ and is intentionally excluded — those
# names are for this script, not for kafkactl, and forwarding them as
# environment to the container would either be no-ops or actively confusing.
FORWARD_PATTERN='^(CONTEXTS_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS$)'
KAFKACTL_ENV_ARGS=()
while IFS= read -r var; do
  [ -z "$var" ] && continue
  KAFKACTL_ENV_ARGS+=(-e "$var")
done < <(compgen -e | grep -E "$FORWARD_PATTERN" || true)

if [ -n "${KAFKA_CONTAINER_RUNTIME:-}" ]; then
  RUNTIME="$KAFKA_CONTAINER_RUNTIME"
elif command -v docker >/dev/null 2>&1; then
  RUNTIME=docker
elif command -v podman >/dev/null 2>&1; then
  RUNTIME=podman
else
  echo "neither docker nor podman found on PATH" >&2
  exit 1
fi

mkdir -p "$OUT" "$OUT/topics" "$OUT/groups"

# Replace anything outside [A-Za-z0-9._-] so the result is safe to use as a
# single path segment. Topic and group names already match this pattern by the
# manifest schema, but we re-sanitize defensively.
sanitize_filename() {
  printf '%s' "$1" | sed 's/[^A-Za-z0-9._-]/_/g'
}

run_kafkactl() {
  local outfile="$1"; shift
  if "$RUNTIME" run --rm -i \
    "${KAFKACTL_ENV_ARGS[@]}" \
    "${EXTRA_ARGS[@]}" \
    "$KAFKACTL_IMAGE" \
    --context "$CONTEXT" \
    "$@" \
    > "$outfile"; then
    return 0
  fi

  echo "warning: kafkactl call failed for $outfile; writing empty file and continuing" >&2
  : > "$outfile"
}

# Cluster — broker list, controller, cluster id. Output as JSON for downstream
# rendering by SKILL.md.
run_kafkactl "$OUT/cluster.json" get brokers --output json

for topic in "${TOPICS[@]}"; do
  safe="$(sanitize_filename "$topic")"
  run_kafkactl "$OUT/topics/${safe}.json" describe topic "$topic" --output json
done

for group in "${GROUPS[@]}"; do
  safe="$(sanitize_filename "$group")"
  run_kafkactl "$OUT/groups/${safe}.json" describe consumer-group "$group" --output json
done

echo "introspection written to $OUT" >&2
echo "  cluster: $OUT/cluster.json" >&2
echo "  topics:  ${#TOPICS[@]} file(s) under $OUT/topics/" >&2
echo "  groups:  ${#GROUPS[@]} file(s) under $OUT/groups/" >&2
