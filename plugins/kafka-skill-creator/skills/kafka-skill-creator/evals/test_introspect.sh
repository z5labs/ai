#!/usr/bin/env bash
# Lightweight tests for scripts/introspect.sh.
#
# Verifies the credential-routing contract from issue #30 and the manifest's
# CONTEXTS_<NAME>_* convention:
#   - Argument shape: --context required, exactly one positional output-dir,
#     no connection strings ever.
#   - Required env vars (CONTEXTS_<NAME>_BROKERS / _SASL_USERNAME /
#     _SASL_PASSWORD) are validated up front and the refusal is a single
#     complete missing-list pointing at credential helpers.
#   - Refusal exits before any container is invoked, so these tests need
#     neither docker/podman nor a Kafka broker.
#   - Positive-path forwarded-env filter behaves as documented (kafkactl-
#     shaped vars in, internal config out).
#
# Run from anywhere; the script resolves its own path:
#   bash evals/test_introspect.sh

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INTROSPECT="$HERE/../scripts/introspect.sh"

if [ ! -x "$INTROSPECT" ] && [ ! -f "$INTROSPECT" ]; then
  echo "FATAL: introspect.sh not found at $INTROSPECT" >&2
  exit 2
fi

PASS=0
FAIL=0
FAILURES=()

# Run introspect.sh with a controlled environment. env_spec is a string of
# `export KEY=VAL` lines (or empty). Args after env_spec are passed to the
# script. Captures merged stdout+stderr and exit code.
run_introspect() {
  local env_spec="$1"; shift
  env -i PATH="$PATH" HOME="$HOME" bash -c "
    $env_spec
    bash '$INTROSPECT' \"\$@\" 2> /tmp/kafka_introspect_err.\$\$ > /tmp/kafka_introspect_out.\$\$
    code=\$?
    cat /tmp/kafka_introspect_out.\$\$
    echo '---STDERR---'
    cat /tmp/kafka_introspect_err.\$\$
    rm -f /tmp/kafka_introspect_out.\$\$ /tmp/kafka_introspect_err.\$\$
    exit \$code
  " _ "$@"
}

# assert_test "name" expected_exit "expected_substr" env_spec args...
# Pass an empty expected_substr to skip the substring check.
assert_test() {
  local name="$1" expected_code="$2" expected_substr="$3" env_spec="$4"
  shift 4
  local output code
  output="$(run_introspect "$env_spec" "$@")"
  code=$?

  local ok=true
  if [ "$code" -ne "$expected_code" ]; then
    ok=false
    FAILURES+=("$name: expected exit $expected_code, got $code. Output:
$output")
  fi
  if [ -n "$expected_substr" ] && ! grep -qF -- "$expected_substr" <<<"$output"; then
    ok=false
    FAILURES+=("$name: expected substring [$expected_substr] not found. Output:
$output")
  fi

  if $ok; then
    PASS=$((PASS + 1))
    echo "PASS: $name"
  else
    FAIL=$((FAIL + 1))
    echo "FAIL: $name"
  fi
}

# Complete env spec for context 'dev'. Used when we want the script to pass
# the env-validator and reach later code paths (e.g. runtime invocation tests).
ALL_DEV='export CONTEXTS_DEV_BROKERS="b1:9092" CONTEXTS_DEV_SASL_USERNAME=u CONTEXTS_DEV_SASL_PASSWORD=p'

# --- Argument-shape tests -----------------------------------------------------

assert_test \
  "rejects zero args with usage" \
  2 \
  "usage:" \
  ""

assert_test \
  "rejects missing --context" \
  2 \
  "usage:" \
  "$ALL_DEV" \
  "/tmp/out"

assert_test \
  "rejects missing positional output-dir" \
  2 \
  "usage:" \
  "$ALL_DEV" \
  --context dev

assert_test \
  "rejects multiple positionals" \
  2 \
  "usage:" \
  "$ALL_DEV" \
  --context dev /tmp/out /tmp/extra

# Connection strings must never reach kafkactl, regardless of where the user
# tries to slip them in.
assert_test \
  "rejects kafka:// connection-string as positional output-dir" \
  2 \
  "connection strings are not accepted" \
  "$ALL_DEV" \
  --context dev "kafka://u:p@b:9092/t"

assert_test \
  "rejects kafka:// as --topic value" \
  2 \
  "connection strings are not accepted" \
  "$ALL_DEV" \
  --context dev --topic "kafka://b:9092/t" /tmp/out

assert_test \
  "rejects user:pass@host as --group value" \
  2 \
  "connection strings are not accepted" \
  "$ALL_DEV" \
  --context dev --group "u:pass@host:9092" /tmp/out

# Equals-form variants must hit the same DSN-rejection path. Earlier the
# `--topic=*` and `--group=*` branches were stripping the prefix and
# pushing straight into the array — a credential-bearing connection
# string written as --topic=kafka://... could have flowed past argv
# inspection, into kafkactl's command line, and on into logs/transcripts.
# Pin the expectation explicitly.
assert_test \
  "rejects kafka:// in --topic=value form" \
  2 \
  "connection strings are not accepted" \
  "$ALL_DEV" \
  --context dev "--topic=kafka://b:9092/t" /tmp/out

assert_test \
  "rejects kafka:// in --group=value form" \
  2 \
  "connection strings are not accepted" \
  "$ALL_DEV" \
  --context dev "--group=kafka://b:9092/g" /tmp/out

assert_test \
  "rejects user:pass@host in --topic=value form" \
  2 \
  "connection strings are not accepted" \
  "$ALL_DEV" \
  --context dev "--topic=u:pass@host:9092" /tmp/out

assert_test \
  "rejects user:pass@host in --group=value form" \
  2 \
  "connection strings are not accepted" \
  "$ALL_DEV" \
  --context dev "--group=u:pass@host:9092" /tmp/out

# --- Output-path guard tests --------------------------------------------------
#
# introspect.sh wipes its output directory before writing so stale files
# from a prior manifest don't linger. That's a real ergonomic improvement,
# but `rm -rf $OUT` on a malformed argument would be catastrophic. Pin the
# guard against the obvious dangerous shapes.

assert_test \
  "refuses to wipe / as output-dir" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev /

assert_test \
  "refuses to wipe . as output-dir" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev .

assert_test \
  "refuses to wipe .. as output-dir" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev ..

assert_test \
  "refuses to wipe ~ as output-dir" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev "~"

# String-shape checks alone are bypassable: rm -rf /tmp/out/.. resolves
# to /tmp before deletion, even though the literal argument doesn't match
# the simple `..` case. Cover each position a `..` path component can
# take: leading, trailing, middle, and lone.

assert_test \
  "refuses output-dir ending in /.. (resolves to parent)" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev /tmp/out/..

assert_test \
  "refuses output-dir with /../ in the middle" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev /tmp/../etc

assert_test \
  "refuses relative output-dir starting with ../" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev ../tmp

assert_test \
  "refuses output-dir with deeper /../../ traversal" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev /tmp/a/../../etc

# --- Context-name validation --------------------------------------------------

# Context name flows into env-var derivation; restrict it to a charset that
# survives uppercasing without collisions.
assert_test \
  "rejects context name starting with digit" \
  2 \
  "not a valid context name" \
  "$ALL_DEV" \
  --context "1bad" /tmp/out

assert_test \
  "rejects context name with dot" \
  2 \
  "not a valid context name" \
  "$ALL_DEV" \
  --context "dev.us-east" /tmp/out

# --- Env-var validation tests -------------------------------------------------

assert_test \
  "refuses with missing-list when no env vars set" \
  2 \
  "missing required environment variables" \
  "" \
  --context dev /tmp/out

assert_test \
  "names CONTEXTS_DEV_SASL_PASSWORD specifically when only it is missing" \
  2 \
  "CONTEXTS_DEV_SASL_PASSWORD" \
  'export CONTEXTS_DEV_BROKERS=b CONTEXTS_DEV_SASL_USERNAME=u' \
  --context dev /tmp/out

assert_test \
  "names CONTEXTS_DEV_BROKERS specifically when only it is missing" \
  2 \
  "CONTEXTS_DEV_BROKERS" \
  'export CONTEXTS_DEV_SASL_USERNAME=u CONTEXTS_DEV_SASL_PASSWORD=p' \
  --context dev /tmp/out

# Empty-string vars must count as missing, matching the postgres-skill-creator
# rule — an empty SASL_PASSWORD would otherwise authenticate as "no password"
# rather than refusing.
assert_test \
  "treats set-but-empty SASL_PASSWORD as missing" \
  2 \
  "CONTEXTS_DEV_SASL_PASSWORD" \
  'export CONTEXTS_DEV_BROKERS=b CONTEXTS_DEV_SASL_USERNAME=u CONTEXTS_DEV_SASL_PASSWORD=' \
  --context dev /tmp/out

# Multi-missing case: all three required keys must appear in a single error,
# not fail-then-fix-then-fail-again.
multi_output="$(run_introspect "" --context dev /tmp/out 2>&1 || true)"
if grep -q "CONTEXTS_DEV_BROKERS" <<<"$multi_output" && \
   grep -q "CONTEXTS_DEV_SASL_USERNAME" <<<"$multi_output" && \
   grep -q "CONTEXTS_DEV_SASL_PASSWORD" <<<"$multi_output"; then
  PASS=$((PASS + 1))
  echo "PASS: lists all three missing vars in a single error"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("multi-missing test: expected BROKERS, SASL_USERNAME, and SASL_PASSWORD all in one error. Output:
$multi_output")
  echo "FAIL: lists all three missing vars in a single error"
fi

# Refusal must point users at credential helpers — that's the documented path
# for populating these vars without leaking them through model context.
helper_output="$(run_introspect "" --context dev /tmp/out 2>&1 || true)"
if grep -qE "(op run|vault|direnv|credential helper)" <<<"$helper_output"; then
  PASS=$((PASS + 1))
  echo "PASS: refusal mentions credential-helper path"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("refusal must mention credential helpers (op run / vault / direnv). Output:
$helper_output")
  echo "FAIL: refusal mentions credential-helper path"
fi

# --- Context normalization (hyphens → underscores in env-var lookup) ---------

# `--context dev-1` should derive prefix CONTEXTS_DEV_1_ . If the user supplied
# CONTEXTS_DEV_1_BROKERS / _SASL_USERNAME / _SASL_PASSWORD, the script must
# pass validation. If they supplied CONTEXTS_DEV-1_BROKERS literally, the
# script should still pass validation because the lookup name is normalized.
assert_test \
  "normalizes hyphens in context name to underscores for env-var lookup" \
  2 \
  "CONTEXTS_DEV_1_BROKERS" \
  "" \
  --context "dev-1" /tmp/out

# --- Positive-path test: invocation shape -------------------------------------
#
# The refusal-path tests above never reach the docker invocation. This block
# stubs KAFKA_CONTAINER_RUNTIME with a fake runtime that just logs its
# arguments to a file, then asserts on what introspect.sh would have passed
# to `docker run`. Catches regressions in the -e VAR forwarding and the
# overall arg shape without needing Kafka or a container runtime.

FAKE_RUNTIME="$(mktemp)"
INVOCATION_LOG="$(mktemp)"
TMPOUT="$(mktemp -d)"
trap 'rm -rf "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT"' EXIT

cat > "$FAKE_RUNTIME" <<'FAKE'
#!/usr/bin/env bash
# Append every argument as one line, terminated by ---END--- so concurrent
# invocations stay separable. Drain stdin so any heredoc-style input doesn't
# break the script. Exit 0 so introspect.sh thinks each kafkactl call
# succeeded.
{
  for a in "$@"; do printf '%s\n' "$a"; done
  echo "---END---"
} >> "$INVOCATION_LOG"
cat > /dev/null
exit 0
FAKE
chmod +x "$FAKE_RUNTIME"

# Run introspect.sh with the fake runtime, a rich set of kafkactl-shaped env
# vars (TLS_CERTKEY, SCHEMAREGISTRY_URL, plus the bare BROKERS shorthand) and
# our internal config that must NOT be forwarded.
env -i PATH="$PATH" HOME="$HOME" \
  INVOCATION_LOG="$INVOCATION_LOG" \
  bash -c "
    export CONTEXTS_DEV_BROKERS='b1:9092 b2:9092'
    export CONTEXTS_DEV_SASL_USERNAME=app
    export CONTEXTS_DEV_SASL_PASSWORD=secret
    export CONTEXTS_DEV_TLS_CERTKEY=/etc/certs/key.pem
    export CONTEXTS_DEV_SCHEMAREGISTRY_URL=https://sr.internal:8081
    export BROKERS='b1:9092'
    export TLS_CERTKEY=/etc/certs/key.pem
    export SASL_USERNAME=defaultuser
    export SCHEMAREGISTRY_URL=https://sr.internal:8081
    export KAFKA_DOCKER_ARGS='--network=host'
    export KAFKA_CONTAINER_RUNTIME='$FAKE_RUNTIME'
    export KAFKACTL_IMAGE='myregistry.example.com/kafkactl:custom'
    bash '$INTROSPECT' --context dev --topic payments.orders.v1 --group payments-orders-projector '$TMPOUT'
  " > /dev/null 2>&1

# Pick the first invocation block.
first_invocation="$(awk '/---END---/{exit} {print}' "$INVOCATION_LOG")"

check_invocation() {
  local name="$1" pattern="$2"
  if grep -qFx -- "$pattern" <<<"$first_invocation"; then
    PASS=$((PASS + 1))
    echo "PASS: $name"
  else
    FAIL=$((FAIL + 1))
    FAILURES+=("$name: expected line [$pattern] in invocation. Captured:
$first_invocation")
    echo "FAIL: $name"
  fi
}

check_invocation_absent() {
  local name="$1" pattern="$2"
  if grep -qFx -- "$pattern" <<<"$first_invocation"; then
    FAIL=$((FAIL + 1))
    FAILURES+=("$name: expected line [$pattern] to be ABSENT. Captured:
$first_invocation")
    echo "FAIL: $name"
  else
    PASS=$((PASS + 1))
    echo "PASS: $name"
  fi
}

# Required CONTEXTS_DEV_* vars must all be forwarded.
for v in CONTEXTS_DEV_BROKERS CONTEXTS_DEV_SASL_USERNAME CONTEXTS_DEV_SASL_PASSWORD; do
  check_invocation "forwards -e $v to runtime" "$v"
done

# Extended kafkactl-shaped vars must also be forwarded — this is what gives
# operators the full kafkactl surface (TLS, Schema Registry, default-context
# shorthand) without rebuilding the image.
for v in CONTEXTS_DEV_TLS_CERTKEY CONTEXTS_DEV_SCHEMAREGISTRY_URL TLS_CERTKEY SASL_USERNAME SCHEMAREGISTRY_URL BROKERS; do
  check_invocation "forwards extended kafkactl var -e $v" "$v"
done

# Internal config must NOT be forwarded as -e flags — these are for this
# script, not for kafkactl. (PG-equivalent guard: no PG_DOCKER_ARGS leakage.)
check_invocation_absent "does NOT forward internal KAFKA_DOCKER_ARGS as -e" "KAFKA_DOCKER_ARGS"
check_invocation_absent "does NOT forward internal KAFKA_CONTAINER_RUNTIME as -e" "KAFKA_CONTAINER_RUNTIME"
check_invocation_absent "does NOT forward KAFKACTL_IMAGE as -e" "KAFKACTL_IMAGE"

# KAFKA_DOCKER_ARGS value must reach the runtime as a standalone argument.
check_invocation "applies KAFKA_DOCKER_ARGS value to runtime" "--network=host"

# Image override must appear as a positional arg.
check_invocation "passes KAFKACTL_IMAGE as positional arg" "myregistry.example.com/kafkactl:custom"

# kafkactl must be invoked with --context <NAME>, not a connection string.
check_invocation "passes --context flag to kafkactl" "--context"
check_invocation "passes context value to kafkactl" "dev"

# No DSN should ever appear in the invocation, even when env vars are rich.
if grep -qE "kafka://|kafkactl://|sasl://|secret@" <<<"$first_invocation"; then
  FAIL=$((FAIL + 1))
  FAILURES+=("invocation contains a connection-string-shaped argument. Captured:
$first_invocation")
  echo "FAIL: no DSN-shaped arg appears in invocation"
else
  PASS=$((PASS + 1))
  echo "PASS: no DSN-shaped arg appears in invocation"
fi

# --- Summary ------------------------------------------------------------------

echo
echo "==============================="
echo "Passed: $PASS"
echo "Failed: $FAIL"
echo "==============================="

if [ "$FAIL" -gt 0 ]; then
  echo
  echo "Failure details:"
  for f in "${FAILURES[@]}"; do
    echo "---"
    echo "$f"
  done
  exit 1
fi
exit 0
