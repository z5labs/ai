#!/usr/bin/env bash
# Introspect a Kafka cluster (via kafkactl) and write JSON files describing
# its topics, consumer groups, and brokers as scoped to one team's manifest.
#
# Usage:
#   introspect.sh \
#     --auth SASL_SCRAM|MTLS \
#     --context <NAME> \
#     --topic <T> [--topic <T> ...] \
#     --group <G> [--group <G> ...] \
#     <output-dir>
#
# Reads connection details from env vars matching kafkactl's documented
# CONTEXTS_<NAME>_* convention. Required keys depend on --auth:
#
#   --auth SASL_SCRAM (default while v1 only had one auth mode — kept
#                      explicit so the script doesn't auto-detect):
#     CONTEXTS_<NAME>_BROKERS                  (whitespace-separated host:port list)
#     CONTEXTS_<NAME>_SASL_USERNAME
#     CONTEXTS_<NAME>_SASL_PASSWORD
#
#   --auth MTLS:
#     CONTEXTS_<NAME>_BROKERS
#     CONTEXTS_<NAME>_TLS_CERT       (absolute path to client cert PEM, must exist)
#     CONTEXTS_<NAME>_TLS_CERTKEY    (absolute path to client key PEM,  must exist)
#     CONTEXTS_<NAME>_TLS_CA         (absolute path to CA bundle PEM,   must exist)
#
#   Each cert path is bind-mounted into the kafkactl container read-only at
#   the same path the env var declares, so kafkactl reads the file at the
#   path it sees in its own env. Only the active --context's cert paths are
#   mounted — paths for other contexts (if they happen to be exported) never
#   reach this run's argv.
#
# One key is optional under SASL_SCRAM (kafkactl's default — pass it
# explicitly to override). Not used under MTLS:
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

# Absolute path to the skill root, resolved from the script's own location.
# Used in error messages that reference sibling docs (e.g. references/...);
# the SKILL.md contract invokes this script via an absolute path from any
# cwd, so a bare `references/...` reference would dangle from the caller's
# working directory.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

usage() {
  cat <<'EOF' >&2
usage: introspect.sh --auth SASL_SCRAM|MTLS --context <NAME> [--topic <T>]... [--group <G>]... <output-dir>

Required:
  --auth MODE       Authentication mode the manifest declared. SASL_SCRAM
                    requires CONTEXTS_<NAME>_SASL_USERNAME / _SASL_PASSWORD.
                    MTLS requires CONTEXTS_<NAME>_TLS_CERT / _TLS_CERTKEY /
                    _TLS_CA, each an absolute path that exists on disk;
                    the cert files are bind-mounted :ro into the kafkactl
                    container at the same path. Other auth values are
                    rejected by the parent SKILL.md before reaching here.

  --context NAME    The kafkactl context to introspect against. Must match
                    one of the manifest's contexts[].name values, uppercased
                    for env-var lookup (e.g. `--context dev` reads
                    CONTEXTS_DEV_BROKERS).

  <output-dir>      Where to write JSON dumps. Wiped and recreated on every
                    run so stale files from a prior manifest don't linger.
                    The leaf segment must start with `kafka-introspect-`
                    (e.g. /tmp/kafka-introspect-<team>) — that prefix is
                    the safety pin that lets the script recursively delete
                    its output without risk of nuking a high-impact dir.
                    The script refuses to wipe: empty path, `/`, `.`, `..`,
                    `~`, anything starting with `~/`, anything starting with
                    whitespace, paths containing `..` segments, and any leaf
                    without the `kafka-introspect-` prefix. A trailing slash
                    is normalized away before the wipe so a symlink-with-
                    trailing-slash can't be followed into its target.

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
AUTH=""
TOPICS=()
# `GROUPS` is a Bash readonly built-in (the supplementary group IDs of
# the current user); assignments to it are silently ignored, so we use
# CONSUMER_GROUPS for the per-script accumulator and let the caller
# spell the flag `--group <name>`.
CONSUMER_GROUPS=()
POSITIONAL=()

while [ $# -gt 0 ]; do
  case "$1" in
    --auth)
      [ $# -ge 2 ] || { echo "error: --auth requires a value" >&2; exit 2; }
      AUTH="$2"
      shift 2
      ;;
    --auth=*)
      AUTH="${1#--auth=}"
      shift
      ;;
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
      CONSUMER_GROUPS+=("$2")
      shift 2
      ;;
    --group=*)
      _v="${1#--group=}"
      reject_positional_dsn "$_v"
      CONSUMER_GROUPS+=("$_v")
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

if [ -z "$CONTEXT" ] || [ -z "$AUTH" ] || [ ${#POSITIONAL[@]} -ne 1 ]; then
  usage
  exit 2
fi

# `--auth` selects which env-var validator and (for MTLS) which cert-mount
# branch runs below. Pin the accepted values explicitly — the parent SKILL.md
# already refuses unsupported auth modes (SASL_PLAIN #63, OAUTHBEARER #65)
# before reaching here, so any other value is a script-level misuse.
case "$AUTH" in
  SASL_SCRAM|MTLS) ;;
  *)
    echo "error: --auth value '$AUTH' is not supported here." >&2
    echo "       expected SASL_SCRAM or MTLS." >&2
    exit 2
    ;;
esac

# Context name flows into env-var lookups (CONTEXTS_<NAME>_*) and into kafkactl
# `--context <NAME>`. Restrict to a charset that survives uppercasing without
# corrupting the env-var derivation.
if ! [[ "$CONTEXT" =~ ^[A-Za-z][A-Za-z0-9_-]*$ ]]; then
  echo "error: --context value '$CONTEXT' is not a valid context name." >&2
  echo "       expected ^[A-Za-z][A-Za-z0-9_-]*$ — letters, digits, underscores, hyphens; first char a letter." >&2
  exit 2
fi

OUT="${POSITIONAL[0]}"

# Output-dir wipe guards run BEFORE env-var validation and container-runtime
# selection. Two reasons: (a) these are the most safety-critical checks
# (a misuse turns into rm -rf on the wrong directory), so they should fire
# on the cheapest input we have; (b) the test suite asserts that refusal-
# path tests work without a container runtime on PATH, which only holds if
# RUNTIME selection happens after these guards.
#
# Refuse the operation on path values that would be catastrophic to wipe.
# This is a sanity net for malformed arguments — the contract is that the
# caller passes a scratch location like /tmp/kafka-introspect-<team>, and
# the OS will catch un-writable paths at mkdir time.
case "$OUT" in
  ""|"/"|"."|".."|"~"|"~/"*|[[:space:]]*)
    echo "error: refusing to wipe suspicious output-dir: $OUT" >&2
    echo "       pass a scratch path like /tmp/kafka-introspect-<team>." >&2
    exit 2
    ;;
esac
# `[[:space:]]*` above (rather than `" "*`) is intentional: bash glob
# character classes match all POSIX whitespace, so a leading tab or
# newline (e.g. from a sloppy paste) is rejected the same way a leading
# space is. The naive `" "*` glob only matches a literal leading space
# and would let `$'\t/tmp'` slip through to the rm.
# Reject any path whose components include `..`. `rm -rf` resolves
# `/tmp/out/..` to `/tmp` (or worse), and the string-shape check above
# wouldn't catch it because the literal string isn't `..` itself. The
# four glob alternatives below cover the four positions a `..` component
# can take: leading (`../foo`), trailing (`foo/..`), middle (`foo/../bar`),
# and lone (`..` — already handled in the case above but kept here for
# documentation symmetry).
case "$OUT" in
  "../"*|*"/.."|*"/../"*)
    echo "error: refusing to wipe output-dir containing '..' segment: $OUT" >&2
    echo "       '..' would let rm -rf resolve to a parent directory." >&2
    exit 2
    ;;
esac
# Final guard: the leaf segment must start with `kafka-introspect-`. This
# is what makes the wipe safe in the common case — a malformed argument
# like `/tmp` or `/home/user/work` would pass the syntactic checks above
# but is catastrophic to recursively delete. The SKILL.md contract
# prescribes `/tmp/kafka-introspect-<team>`; pinning the leaf prefix
# here turns that prescription into enforcement, so a misuse hits the
# refusal before any data on disk is touched.
#
# Use pure bash parameter expansion rather than `basename -- "$OUT"`:
# the `--` end-of-options marker is GNU-only and BSD/macOS basename
# rejects it, which would break the very macOS-default-Bash environment
# the script otherwise targets. The `--` end-of-options marker isn't
# needed here either way: parameter expansion treats `-` like any other
# character, and a caller who slips a dash-prefixed OUT past the
# argparse loop via the `--` end-of-options marker (e.g. `... --context
# dev -- -evil`) still hits the leaf-prefix refusal below, since no
# leaf starting with `-` matches `kafka-introspect-*`.
#
# Loop the trailing-slash strip rather than `${OUT%/}` once: a caller
# who passes `.../kafka-introspect-foo///` would otherwise leave OUT
# ending in `//`, and `rm -rf path//` still dereferences the symlink
# (any trailing slash forces deref, not just one). Iterating until no
# trailing slash remains makes the symlink-safety guarantee hold for
# any number of slashes.
out_no_trail="$OUT"
while [ "${out_no_trail%/}" != "$out_no_trail" ]; do
  out_no_trail="${out_no_trail%/}"
done
out_leaf="${out_no_trail##*/}"
case "$out_leaf" in
  kafka-introspect-?*) ;;
  *)
    echo "error: refusing to wipe output-dir whose leaf segment is not 'kafka-introspect-<...>': $OUT" >&2
    echo "       this script wipes the directory recursively before writing; the kafka-introspect- prefix" >&2
    echo "       is the safety pin that prevents accidental destruction of a high-impact directory." >&2
    exit 2
    ;;
esac
# Adopt the de-trailed path for the wipe and recreate steps. `rm -rf path/`
# (with trailing slash) on a symlink dereferences the symlink and recursively
# deletes its target — even if the symlink itself sits in a kafka-introspect-
# directory, the target could be anywhere on disk. Normalizing OUT to the
# non-trailing form ensures `rm -rf -- "$OUT"` removes the link itself,
# never what it points to.
OUT="$out_no_trail"

# Derive the env-var prefix kafkactl uses for the chosen context. kafkactl
# uppercases the context name and prepends `CONTEXTS_`. Hyphens become
# underscores per kafkactl's own normalization.
#
# Use `tr` rather than `${var^^}` for the uppercase — `^^` is a Bash 4+
# parameter-expansion feature, but macOS still ships Bash 3.2 by default
# and we want a developer on a default Mac shell to be able to invoke
# this script without first installing a newer bash.
ctx_upper="$(printf '%s' "$CONTEXT" | tr '[:lower:]' '[:upper:]')"
ctx_upper="${ctx_upper//-/_}"
prefix="CONTEXTS_${ctx_upper}"

# Validate every credential-bearing env var up front so the caller gets a
# single complete missing-list rather than discovering them one kafkactl
# failure at a time. Same shape postgres-skill-creator landed in issue #42.
# Auth mode picks which keys are required: SASL_SCRAM needs username +
# password, MTLS needs the three cert paths.
case "$AUTH" in
  SASL_SCRAM)
    required=(
      "${prefix}_BROKERS"
      "${prefix}_SASL_USERNAME"
      "${prefix}_SASL_PASSWORD"
    )
    ;;
  MTLS)
    required=(
      "${prefix}_BROKERS"
      "${prefix}_TLS_CERT"
      "${prefix}_TLS_CERTKEY"
      "${prefix}_TLS_CA"
    )
    ;;
esac
missing=()
for var in "${required[@]}"; do
  if [ -z "${!var:-}" ]; then
    missing+=("$var")
  fi
done
if [ ${#missing[@]} -gt 0 ]; then
  {
    echo "error: missing required environment variables for context '$CONTEXT' (auth: $AUTH):"
    for v in "${missing[@]}"; do echo "  - $v"; done
    echo
    echo "export them (or load them from a credential helper such as op run, vault,"
    echo "direnv, gcloud) before re-running. these names follow kafkactl's documented"
    echo "CONTEXTS_<NAME>_* convention; see ${SKILL_DIR}/references/kafkactl-env-vars.md."
  } >&2
  exit 2
fi

# For MTLS, the three cert vars must each be absolute paths to existing
# files — kafkactl will read them from inside the container at the same
# path the env var declares, and docker's bind-mount syntax requires the
# host path be absolute. Surface every offender in one error so an operator
# fixing a missing-cert mistake doesn't have to re-run the script three
# times to discover the next problem. CERT_PATHS stays declared (empty)
# under SASL_SCRAM so the later mount-args loop is a no-op without
# tripping `set -u` on an unset array.
CERT_PATHS=()
if [ "$AUTH" = "MTLS" ]; then
  cert_problems=()
  for var in "${prefix}_TLS_CERT" "${prefix}_TLS_CERTKEY" "${prefix}_TLS_CA"; do
    val="${!var}"
    case "$val" in
      /*) ;;
      *)
        cert_problems+=("$var=$val (must be an absolute path; got relative)")
        continue
        ;;
    esac
    if [ ! -f "$val" ]; then
      cert_problems+=("$var=$val (file not found on host)")
      continue
    fi
    # `-r` is distinct from `-f`: a file the running process cannot read
    # (wrong owner, restrictive mode, or covered by mandatory access
    # control like SELinux at the host level) would still pass `-f` and
    # then fail opaquely inside the container — the wrapper's job is to
    # surface every preflight problem up front, so reject unreadable
    # cert paths with a named error too.
    if [ ! -r "$val" ]; then
      cert_problems+=("$var=$val (file exists but is not readable by this process; check permissions/SELinux)")
      continue
    fi
    CERT_PATHS+=("$val")
  done
  if [ ${#cert_problems[@]} -gt 0 ]; then
    {
      echo "error: cert paths for context '$CONTEXT' are not usable:"
      for p in "${cert_problems[@]}"; do echo "  - $p"; done
      echo
      echo "each TLS_CERT / TLS_CERTKEY / TLS_CA must be an absolute path to a file the"
      echo "host can read; the wrapper bind-mounts each path :ro,z into the kafkactl"
      echo "container at the same path the env var declares."
    } >&2
    exit 2
  fi
fi

KAFKACTL_IMAGE="${KAFKACTL_IMAGE:-docker.io/deviceinsight/kafkactl:v5.18.0-scratch}"

# Translate the SASL mechanism to kafkactl's expected casing. The manifest
# (and Kafka itself) spell SCRAM mechanisms `SCRAM-SHA-{256,512}`, but
# kafkactl accepts only the squashed lowercase form `scram-sha{256,512}`
# and rejects the canonical Kafka spelling with `Unknown sasl mechanism`.
# Re-export under the same env-var name so the container's forwarding
# filter picks up the corrected value without us touching the FORWARD_PATTERN.
# Only relevant under SASL_SCRAM — MTLS has no SASL exchange.
if [ "$AUTH" = "SASL_SCRAM" ]; then
  mech_var="${prefix}_SASL_MECHANISM"
  if [ -n "${!mech_var:-}" ]; then
    case "${!mech_var}" in
      SCRAM-SHA-256|scram-sha-256) export "$mech_var"=scram-sha256 ;;
      SCRAM-SHA-512|scram-sha-512) export "$mech_var"=scram-sha512 ;;
      *) ;;  # leave alone; kafkactl will surface its own error
    esac
  fi
fi

# Generate a kafkactl config.yml that pre-declares the chosen context.
# kafkactl rejects `--context <name>` if the context isn't present in
# its config file; CONTEXTS_<NAME>_* env-var overlays only modify
# already-declared contexts, they don't create them. An empty-bodied
# entry is enough — the env vars supply broker addresses, credentials,
# and SR settings at runtime.
KAFKACTL_CONFIG_DIR="$(mktemp -d -t kafkactl-cfg-XXXXXX)"
trap 'rm -rf -- "$KAFKACTL_CONFIG_DIR"' EXIT
cat > "$KAFKACTL_CONFIG_DIR/config.yml" <<EOF
contexts:
  $CONTEXT: {}
EOF

# Word-split KAFKA_DOCKER_ARGS into an array so callers can pass multiple flags.
read -r -a EXTRA_ARGS <<< "${KAFKA_DOCKER_ARGS:-}"

# Forward kafkactl-relevant env vars into the container. The filter is the
# union of:
#   - CONTEXTS_<ACTIVE_UPPER>_*    (only the active context's per-context vars)
#   - TLS_*, SASL_*, SCHEMAREGISTRY_*, BROKERS  (kafkactl's bare shorthand
#     for the `default` context — not per-context, so safe to forward)
#
# Other contexts' CONTEXTS_<OTHER>_* vars are intentionally NOT forwarded.
# kafkactl with --context <active> only consults <active>'s vars anyway, so
# the OTHER vars are functionally useless inside the container — and (this
# is the security-relevant part) forwarding e.g. CONTEXTS_PROD_TLS_CERT to
# a `--context dev` container would leak the prod cert path string into
# the container's environment even though the file is not bind-mounted.
# Scoping the pattern to the active context closes that path-leak.
#
# Internal config (KAFKA_DOCKER_ARGS, KAFKA_CONTAINER_RUNTIME,
# KAFKACTL_IMAGE) starts with KAFKA_ and is intentionally excluded — those
# names are for this script, not for kafkactl, and forwarding them as
# environment to the container would either be no-ops or actively confusing.
FORWARD_PATTERN="^(CONTEXTS_${ctx_upper}_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS\$)"
KAFKACTL_ENV_ARGS=()
while IFS= read -r var; do
  [ -z "$var" ] && continue
  KAFKACTL_ENV_ARGS+=(-e "$var")
done < <(compgen -e | grep -E "$FORWARD_PATTERN" || true)

# Build cert mount args. For MTLS, the three cert paths from CERT_PATHS
# (already validated absolute + existing) bind-mount :ro into the container
# at the same path so kafkactl reads them at the path the env var declares.
# Only the active context's paths reach the runtime: CONTEXTS_<OTHER>_*
# vars are dropped by the per-context-scoped FORWARD_PATTERN above, so
# they don't reach the container as env vars OR as mounts. Empty under
# SASL_SCRAM.
MOUNT_ARGS=()
for cert_path in "${CERT_PATHS[@]}"; do
  # `:z` is the SELinux shared-relabel marker; ignored on systems without
  # SELinux. Same reason `run_kafkactl`'s config-mount uses `:ro,z`:
  # without it, Fedora/RHEL hosts hit "Permission denied" reading the
  # cert from inside the container even though the file exists and is
  # bind-mounted, because the host's file context isn't accessible from
  # the container's process context.
  MOUNT_ARGS+=(-v "$cert_path:$cert_path:ro,z")
done

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

# Wipe and recreate the output directory so a re-introspection after a
# manifest change (topic dropped, group renamed) doesn't leave stale JSON
# files lying around to confuse downstream rendering. The path was already
# validated above, before any env / runtime work; everything between then
# and now has been read-only setup.
rm -rf -- "$OUT"
mkdir -p "$OUT" "$OUT/topics" "$OUT/groups"

# Replace anything outside [A-Za-z0-9._-] so the result is safe to use as a
# single path segment. Topic and group names already match this pattern by the
# manifest schema, but we re-sanitize defensively.
sanitize_filename() {
  printf '%s' "$1" | sed 's/[^A-Za-z0-9._-]/_/g'
}

run_kafkactl() {
  local outfile="$1"; shift
  # `:z` is the SELinux shared-relabel marker; ignored on systems without
  # SELinux. Without it, Fedora/RHEL hosts hit "Permission denied" reading
  # the config because the host's file context isn't accessible from the
  # container's process context.
  if "$RUNTIME" run --rm -i \
    -v "$KAFKACTL_CONFIG_DIR/config.yml:/.config/kafkactl/config.yml:ro,z" \
    "${MOUNT_ARGS[@]}" \
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

for group in "${CONSUMER_GROUPS[@]}"; do
  safe="$(sanitize_filename "$group")"
  run_kafkactl "$OUT/groups/${safe}.json" describe consumer-group "$group" --output json
done

echo "introspection written to $OUT" >&2
echo "  cluster: $OUT/cluster.json" >&2
echo "  topics:  ${#TOPICS[@]} file(s) under $OUT/topics/" >&2
echo "  groups:  ${#CONSUMER_GROUPS[@]} file(s) under $OUT/groups/" >&2
