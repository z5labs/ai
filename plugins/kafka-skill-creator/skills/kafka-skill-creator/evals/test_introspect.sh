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
#
# Auto-injects `--auth SASL_SCRAM` when the args neither contain `--auth`
# nor are empty, so existing tests written against the SASL-only v1 keep
# passing after the --auth flag was added. Tests that exercise the auth
# flag itself (missing / unknown / MTLS) pass it explicitly and bypass
# the auto-injection.
run_introspect() {
  local env_spec="$1"; shift
  local has_auth=false
  local arg
  for arg in "$@"; do
    case "$arg" in
      --auth|--auth=*) has_auth=true; break ;;
    esac
  done
  if [ $# -gt 0 ] && ! $has_auth; then
    set -- --auth SASL_SCRAM "$@"
  fi
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

# Whitespace-prefixed paths must be rejected for ALL POSIX whitespace,
# not just a literal space. A naive `" "*` glob would let `$'\t/tmp'` or
# `$'\n/tmp'` slip past — the `[[:space:]]*` glob in the script catches
# tab and newline too. Pin both so a regression to the literal-space
# pattern is caught.
assert_test \
  "refuses output-dir starting with tab" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev $'\t/tmp/kafka-introspect-foo'

assert_test \
  "refuses output-dir starting with newline" \
  2 \
  "refusing to wipe" \
  "$ALL_DEV" \
  --context dev $'\n/tmp/kafka-introspect-foo'

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

# Leaf-prefix guard: even a syntactically-clean absolute path is refused
# unless its leaf starts with `kafka-introspect-`. This is what stops
# `--context dev /tmp` (or `/home/user/work`, or any other innocuous-
# looking but high-impact directory) from being recursively wiped.

assert_test \
  "refuses /tmp as output-dir (leaf doesn't start with kafka-introspect-)" \
  2 \
  "kafka-introspect-" \
  "$ALL_DEV" \
  --context dev /tmp

assert_test \
  "refuses /home/user/work as output-dir" \
  2 \
  "kafka-introspect-" \
  "$ALL_DEV" \
  --context dev /home/user/work

assert_test \
  "refuses /tmp/some-other-dir as output-dir" \
  2 \
  "kafka-introspect-" \
  "$ALL_DEV" \
  --context dev /tmp/some-other-dir

# A caller can technically slip a dash-prefixed positional past the
# argparse loop's `-*)` guard via `--` end-of-options. The leaf-prefix
# check is the safety net that catches it (no leaf starting with `-`
# matches `kafka-introspect-*`); pin that guarantee so the inline
# comment in introspect.sh stays honest.
assert_test \
  "refuses dash-prefixed leaf slipped in via -- end-of-options" \
  2 \
  "kafka-introspect-" \
  "$ALL_DEV" \
  --context dev -- -evil-leaf

# Trailing slash on a kafka-introspect- path should still pass the guard
# (basename strips the slash). This is the positive case for the prefix
# check, exercised here via the env-var-validation refusal so the test
# doesn't need a fake runtime.
assert_test \
  "accepts /tmp/kafka-introspect-payments at the prefix-guard layer (refusal happens later for missing creds)" \
  2 \
  "missing required environment variables" \
  "" \
  --context dev /tmp/kafka-introspect-payments

# --- Ordering invariant: output-dir guards fire before runtime selection -----
#
# The output-path guards must reject before the script tries to pick docker /
# podman. Two reasons matter together:
#  1. The header of this file claims refusal-path tests don't need a container
#     runtime — only true if RUNTIME selection runs after these guards.
#  2. The wipe is the most catastrophic operation in this script; failing on
#     the cheapest input check (path shape) is the right defense in depth.
#
# Build a sandbox PATH directory containing only the utilities introspect.sh
# needs to reach the output-dir guard — explicitly excluding docker/podman.
# Stripping whole directories from $PATH would also drop bash itself
# (/usr/bin contains both bash and podman on most distros), so the symlink-
# in-a-clean-dir shape is the only portable way to get a docker-less PATH.
SANDBOX_BIN="$(mktemp -d "${TMPDIR:-/tmp}/kafka-introspect-test-bin-XXXXXX")"
trap 'rm -rf -- "$SANDBOX_BIN"' EXIT
for cmd in bash sh tr printf cat mkdir rm grep sed paste env dirname; do
  src="$(command -v "$cmd" 2>/dev/null || true)"
  [ -n "$src" ] && ln -sf "$src" "$SANDBOX_BIN/$cmd"
done
no_runtime_output="$(env -i PATH="$SANDBOX_BIN" HOME="$HOME" "$SANDBOX_BIN/bash" -c "
  $ALL_DEV
  bash '$INTROSPECT' --auth SASL_SCRAM --context dev /tmp/some-other-dir 2>&1
" || true)"
if grep -qF -- "kafka-introspect-" <<<"$no_runtime_output" \
   && ! grep -qF -- "neither docker nor podman" <<<"$no_runtime_output"; then
  PASS=$((PASS + 1))
  echo "PASS: output-dir guard fires before runtime selection (no docker/podman needed)"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("ordering: output-dir guard must fire without docker/podman on PATH. Output:
$no_runtime_output")
  echo "FAIL: output-dir guard fires before runtime selection (no docker/podman needed)"
fi

# --- Trailing-slash symlink hazard -------------------------------------------
#
# `rm -rf path/` with a trailing slash on a symlink dereferences the link
# and recursively deletes the *target*. The kafka-introspect- leaf guard
# only checks the link's own name, not what it points at — so a symlink
# named kafka-introspect-foo pointing at $HOME would slip past the leaf
# check and the wipe would land in $HOME if the trailing slash isn't
# normalized away. Pin that the script normalizes OUT to its de-trailed
# form before the rm so the link itself is removed, not its target.

SENTINEL_DIR="$(mktemp -d "${TMPDIR:-/tmp}/kafka-introspect-sentinel-XXXXXX")"
SENTINEL_FILE="$SENTINEL_DIR/DO_NOT_TOUCH"
echo "marker" > "$SENTINEL_FILE"

LINK_PARENT="$(mktemp -d "${TMPDIR:-/tmp}/kafka-introspect-linktest-XXXXXX")"
SYMLINK_PATH="$LINK_PARENT/kafka-introspect-symlinktest"
ln -s "$SENTINEL_DIR" "$SYMLINK_PATH"

SYMLINK_FAKE_RUNTIME="$(mktemp)"
cat > "$SYMLINK_FAKE_RUNTIME" <<'FAKE'
#!/usr/bin/env bash
cat > /dev/null 2>&1 || true
exit 0
FAKE
chmod +x "$SYMLINK_FAKE_RUNTIME"

env -i PATH="$PATH" HOME="$HOME" bash -c "
  $ALL_DEV
  export KAFKA_CONTAINER_RUNTIME='$SYMLINK_FAKE_RUNTIME'
  bash '$INTROSPECT' --auth SASL_SCRAM --context dev '$SYMLINK_PATH/' > /dev/null 2>&1 || true
"

if [ -f "$SENTINEL_FILE" ]; then
  PASS=$((PASS + 1))
  echo "PASS: trailing slash on a symlink output-dir doesn't follow into the target"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("symlink target was deleted! Sentinel $SENTINEL_FILE no longer exists. The trailing-slash symlink hazard guard is broken.")
  echo "FAIL: trailing slash on a symlink output-dir doesn't follow into the target"
fi

# Re-create the symlink and re-test with MULTIPLE trailing slashes. Two
# discriminators matter together:
#   (a) the sentinel must survive — same hazard as the single-slash case;
#       any trailing slash forces `rm -rf` to dereference, not just one.
#   (b) the script must EXIT 0, meaning it actually reached the wipe and
#       the wipe operated on the link itself. A single-strip implementation
#       (`${OUT%/}` once) on `kafka-introspect-foo///` leaves the leaf
#       check seeing an empty string after one strip — the script refuses
#       with exit 2, the rm never runs, and the sentinel survives by
#       *accident* of the refusal. The exit-0 assertion is what catches
#       that false-pass: the loop-strip fix must accept `///` as a clean
#       path, not refuse it on the leaf check.
ln -s "$SENTINEL_DIR" "$SYMLINK_PATH"   # the wipe above correctly removed the link itself; recreate for the next case
echo "marker" > "$SENTINEL_FILE"

multislash_exit=0
env -i PATH="$PATH" HOME="$HOME" bash -c "
  $ALL_DEV
  export KAFKA_CONTAINER_RUNTIME='$SYMLINK_FAKE_RUNTIME'
  bash '$INTROSPECT' --auth SASL_SCRAM --context dev '${SYMLINK_PATH}///' > /dev/null 2>&1
" || multislash_exit=$?

if [ -f "$SENTINEL_FILE" ] && [ "$multislash_exit" -eq 0 ]; then
  PASS=$((PASS + 1))
  echo "PASS: multiple trailing slashes on a symlink output-dir are accepted and don't follow into the target"
else
  FAIL=$((FAIL + 1))
  if [ ! -f "$SENTINEL_FILE" ]; then
    FAILURES+=("symlink target was deleted via multi-trailing-slash bypass! Sentinel $SENTINEL_FILE no longer exists. The trailing-slash strip must loop until no trailing / remains.")
  else
    FAILURES+=("multi-trailing-slash path was refused (exit $multislash_exit) instead of being de-trailed and accepted. A single-strip implementation passes the sentinel check by accident — the script must accept '$SYMLINK_PATH///' and exit 0 after the wipe.")
  fi
  echo "FAIL: multiple trailing slashes on a symlink output-dir are accepted and don't follow into the target"
fi

rm -rf -- "$SENTINEL_DIR" "$LINK_PARENT" "$SYMLINK_FAKE_RUNTIME"

# --- Context-name validation --------------------------------------------------

# Context name flows into env-var derivation; restrict it to a charset that
# survives uppercasing without collisions. Use a kafka-introspect- prefixed
# path so we exercise the context regex and not the (earlier) leaf-prefix
# guard.
assert_test \
  "rejects context name starting with digit" \
  2 \
  "not a valid context name" \
  "$ALL_DEV" \
  --context "1bad" /tmp/kafka-introspect-test

assert_test \
  "rejects context name with dot" \
  2 \
  "not a valid context name" \
  "$ALL_DEV" \
  --context "dev.us-east" /tmp/kafka-introspect-test

# --- Env-var validation tests -------------------------------------------------
#
# Output-dir wipe guards now fire BEFORE env-var validation, so these tests
# need a path whose leaf passes the kafka-introspect- prefix check. Otherwise
# the script would refuse on the path and never reach the env-var logic these
# assertions are about.

assert_test \
  "refuses with missing-list when no env vars set" \
  2 \
  "missing required environment variables" \
  "" \
  --context dev /tmp/kafka-introspect-test

assert_test \
  "names CONTEXTS_DEV_SASL_PASSWORD specifically when only it is missing" \
  2 \
  "CONTEXTS_DEV_SASL_PASSWORD" \
  'export CONTEXTS_DEV_BROKERS=b CONTEXTS_DEV_SASL_USERNAME=u' \
  --context dev /tmp/kafka-introspect-test

assert_test \
  "names CONTEXTS_DEV_BROKERS specifically when only it is missing" \
  2 \
  "CONTEXTS_DEV_BROKERS" \
  'export CONTEXTS_DEV_SASL_USERNAME=u CONTEXTS_DEV_SASL_PASSWORD=p' \
  --context dev /tmp/kafka-introspect-test

# Empty-string vars must count as missing, matching the postgres-skill-creator
# rule — an empty SASL_PASSWORD would otherwise authenticate as "no password"
# rather than refusing.
assert_test \
  "treats set-but-empty SASL_PASSWORD as missing" \
  2 \
  "CONTEXTS_DEV_SASL_PASSWORD" \
  'export CONTEXTS_DEV_BROKERS=b CONTEXTS_DEV_SASL_USERNAME=u CONTEXTS_DEV_SASL_PASSWORD=' \
  --context dev /tmp/kafka-introspect-test

# Multi-missing case: all three required keys must appear in a single error,
# not fail-then-fix-then-fail-again.
multi_output="$(run_introspect "" --context dev /tmp/kafka-introspect-test 2>&1 || true)"
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
helper_output="$(run_introspect "" --context dev /tmp/kafka-introspect-test 2>&1 || true)"
if grep -qE "(op run|vault|direnv|credential helper)" <<<"$helper_output"; then
  PASS=$((PASS + 1))
  echo "PASS: refusal mentions credential-helper path"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("refusal must mention credential helpers (op run / vault / direnv). Output:
$helper_output")
  echo "FAIL: refusal mentions credential-helper path"
fi

# The kafkactl-env-vars.md reference must resolve to a real file. SKILL.md
# invokes this script via an absolute path from arbitrary cwds, so a bare
# relative `references/...` would dangle. Extract the path from the error
# message and assert it exists on disk — the strongest possible assertion
# that the operator can actually follow the pointer.
docref_path="$(grep -oE '/[^ ]*references/kafkactl-env-vars\.md' <<<"$helper_output" | head -n1)"
if [ -n "$docref_path" ] && [ -f "$docref_path" ]; then
  PASS=$((PASS + 1))
  echo "PASS: refusal references kafkactl-env-vars.md by an existing absolute path"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("refusal must reference kafkactl-env-vars.md as an absolute path that exists on disk. Extracted: '$docref_path'. Output:
$helper_output")
  echo "FAIL: refusal references kafkactl-env-vars.md by an existing absolute path"
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
  --context "dev-1" /tmp/kafka-introspect-test

# --- Positive-path test: invocation shape -------------------------------------
#
# The refusal-path tests above never reach the docker invocation. This block
# stubs KAFKA_CONTAINER_RUNTIME with a fake runtime that just logs its
# arguments to a file, then asserts on what introspect.sh would have passed
# to `docker run`. Catches regressions in the -e VAR forwarding and the
# overall arg shape without needing Kafka or a container runtime.

FAKE_RUNTIME="$(mktemp)"
INVOCATION_LOG="$(mktemp)"
# Prefix the temp dir with `kafka-introspect-` so it satisfies the script's
# leaf-prefix safety check. mktemp's positional template form works on
# both GNU and BSD mktemp.
TMPOUT="$(mktemp -d "${TMPDIR:-/tmp}/kafka-introspect-XXXXXX")"
trap 'rm -rf "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT" "$SANDBOX_BIN"' EXIT

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
# our internal config that must NOT be forwarded. Also export CONTEXTS_PROD_*
# vars even though we're invoking with --context dev — those must NOT reach
# the container (path-leak prevention; see FORWARD_PATTERN in introspect.sh).
env -i PATH="$PATH" HOME="$HOME" \
  INVOCATION_LOG="$INVOCATION_LOG" \
  bash -c "
    export CONTEXTS_DEV_BROKERS='b1:9092 b2:9092'
    export CONTEXTS_DEV_SASL_USERNAME=app
    export CONTEXTS_DEV_SASL_PASSWORD=secret
    export CONTEXTS_DEV_TLS_CERTKEY=/etc/certs/key.pem
    export CONTEXTS_DEV_SCHEMAREGISTRY_URL=https://sr.internal:8081
    export CONTEXTS_PROD_BROKERS='prod1:9092'
    export CONTEXTS_PROD_SASL_USERNAME=prod-app
    export CONTEXTS_PROD_SASL_PASSWORD=prod-secret
    export CONTEXTS_PROD_TLS_CERT=/etc/certs/prod-cert.pem
    export BROKERS='b1:9092'
    export TLS_CERTKEY=/etc/certs/key.pem
    export SASL_USERNAME=defaultuser
    export SCHEMAREGISTRY_URL=https://sr.internal:8081
    export KAFKA_DOCKER_ARGS='--network=host'
    export KAFKA_CONTAINER_RUNTIME='$FAKE_RUNTIME'
    export KAFKACTL_IMAGE='myregistry.example.com/kafkactl:custom'
    bash '$INTROSPECT' --auth SASL_SCRAM --context dev --topic payments.orders.v1 --group payments-orders-projector '$TMPOUT'
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

# Other contexts' CONTEXTS_<OTHER>_* vars must NOT be forwarded when --context
# is dev. kafkactl ignores them anyway under --context dev, but forwarding the
# path strings into the container would leak prod cert paths into a dev
# invocation — exactly what FORWARD_PATTERN's per-context scoping prevents.
for v in CONTEXTS_PROD_BROKERS CONTEXTS_PROD_SASL_USERNAME CONTEXTS_PROD_SASL_PASSWORD CONTEXTS_PROD_TLS_CERT; do
  check_invocation_absent "does NOT forward non-active-context var $v under --context dev" "$v"
done

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

# --- Positive-path test: kafkactl compatibility shims -------------------------
#
# Two skill changes that only show up when the script actually reaches the
# container invocation:
#   1. SCRAM mechanism casing translation (SCRAM-SHA-{256,512} -> scram-sha{256,512})
#      because kafkactl rejects the canonical Kafka spelling with
#      "Unknown sasl mechanism".
#   2. config.yml is minted in a temp dir and mounted into the kafkactl
#      container, because kafkactl rejects `--context <name>` for contexts
#      not pre-declared in a config file even when CONTEXTS_<NAME>_*
#      env-var overlays cover every field.
#
# Both regressions broke live-cluster introspection silently before
# (kafkactl errored with opaque messages, the script bubbled the failure
# generically). Stub the runtime with a fake that dumps both argv AND the
# kafkactl-shape env values it would see, so we can verify the translated
# mechanism reaches the container's environment.

FAKE_RUNTIME_COMPAT="$(mktemp)"
INVOCATION_LOG_COMPAT="$(mktemp)"
TMPOUT_COMPAT="$(mktemp -d "${TMPDIR:-/tmp}/kafka-introspect-XXXXXX")"
trap 'rm -rf "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT" "$SANDBOX_BIN" "$FAKE_RUNTIME_COMPAT" "$INVOCATION_LOG_COMPAT" "$TMPOUT_COMPAT"' EXIT

cat > "$FAKE_RUNTIME_COMPAT" <<'FAKE'
#!/usr/bin/env bash
{
  for a in "$@"; do printf 'ARG %s\n' "$a"; done
  # docker -e VAR (no =value) reads the value from the host's env at
  # invocation time, so the value isn't in argv. Dump every kafkactl-shape
  # env so tests can assert on what the container would have seen.
  while IFS= read -r v; do
    [ -z "$v" ] && continue
    printf 'ENV %s=%s\n' "$v" "${!v}"
  done < <(compgen -e | grep -E '^(CONTEXTS_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS$)' || true)
  echo "---END---"
} >> "$INVOCATION_LOG_COMPAT"
cat > /dev/null
exit 0
FAKE
chmod +x "$FAKE_RUNTIME_COMPAT"

# Run introspect.sh once per mechanism, capture the first invocation each time.
run_compat_invocation() {
  local mechanism="$1"
  : > "$INVOCATION_LOG_COMPAT"
  env -i PATH="$PATH" HOME="$HOME" \
    INVOCATION_LOG_COMPAT="$INVOCATION_LOG_COMPAT" \
    bash -c "
      export CONTEXTS_DEV_BROKERS='b1:9092'
      export CONTEXTS_DEV_SASL_USERNAME=app
      export CONTEXTS_DEV_SASL_PASSWORD=secret
      export CONTEXTS_DEV_SASL_MECHANISM='$mechanism'
      export KAFKA_CONTAINER_RUNTIME='$FAKE_RUNTIME_COMPAT'
      bash '$INTROSPECT' --auth SASL_SCRAM --context dev --topic t1 --group g1 '$TMPOUT_COMPAT'
    " > /dev/null 2>&1
  awk '/---END---/{exit} {print}' "$INVOCATION_LOG_COMPAT"
}

assert_compat_line() {
  local name="$1" pattern="$2" captured="$3"
  if grep -qFx -- "$pattern" <<<"$captured"; then
    PASS=$((PASS + 1))
    echo "PASS: $name"
  else
    FAIL=$((FAIL + 1))
    FAILURES+=("$name: expected line [$pattern] in invocation. Captured:
$captured")
    echo "FAIL: $name"
  fi
}

assert_compat_grep() {
  local name="$1" pattern="$2" captured="$3"
  if grep -qE -- "$pattern" <<<"$captured"; then
    PASS=$((PASS + 1))
    echo "PASS: $name"
  else
    FAIL=$((FAIL + 1))
    FAILURES+=("$name: expected pattern [$pattern] in invocation. Captured:
$captured")
    echo "FAIL: $name"
  fi
}

# SCRAM-SHA-512 → scram-sha512
captured_512="$(run_compat_invocation SCRAM-SHA-512)"
assert_compat_line \
  "translates SCRAM-SHA-512 to scram-sha512 for kafkactl" \
  "ENV CONTEXTS_DEV_SASL_MECHANISM=scram-sha512" \
  "$captured_512"

# SCRAM-SHA-256 → scram-sha256
captured_256="$(run_compat_invocation SCRAM-SHA-256)"
assert_compat_line \
  "translates SCRAM-SHA-256 to scram-sha256 for kafkactl" \
  "ENV CONTEXTS_DEV_SASL_MECHANISM=scram-sha256" \
  "$captured_256"

# Already-lowercase passes through unchanged. (The shim accepts both
# canonical and translated forms; an operator who's already exporting the
# kafkactl form shouldn't have it re-mangled.)
captured_passthrough="$(run_compat_invocation scram-sha512)"
assert_compat_line \
  "leaves already-lowercase scram-sha512 alone" \
  "ENV CONTEXTS_DEV_SASL_MECHANISM=scram-sha512" \
  "$captured_passthrough"

# Unknown mechanism is left alone — kafkactl will surface its own error.
# The shim is for casing, not for validation.
captured_unknown="$(run_compat_invocation FUTUREMECH-XYZ)"
assert_compat_line \
  "leaves unknown SASL mechanisms unchanged for kafkactl to reject" \
  "ENV CONTEXTS_DEV_SASL_MECHANISM=FUTUREMECH-XYZ" \
  "$captured_unknown"

# config.yml mount: every invocation must include a -v flag mounting a
# config.yml at /.config/kafkactl/config.yml inside the container, with
# the SELinux shared-relabel marker (`:z`) so Fedora/RHEL hosts work.
assert_compat_grep \
  "mounts config.yml at /.config/kafkactl/config.yml inside the container" \
  '^ARG .*:/.config/kafkactl/config\.yml:ro,z$' \
  "$captured_512"

# --- Auth flag tests (added with mTLS support, #64) ---------------------------
#
# The --auth flag was added so introspect.sh knows whether to require
# SASL_USERNAME/PASSWORD or TLS_CERT/CERTKEY/CA per context, and (for MTLS)
# so it can bind-mount the cert paths into the kafkactl container. Pin the
# refusal shape and the positive-path mount/env behavior.

# Refuses without --auth. The auto-injection in run_introspect doesn't fire
# here because we pass the flag explicitly with no value.
assert_test \
  "refuses missing --auth value" \
  2 \
  "--auth requires a value" \
  "$ALL_DEV" \
  --auth

# Refuses unknown --auth value.
assert_test \
  "refuses unknown --auth value" \
  2 \
  "is not supported here" \
  "$ALL_DEV" \
  --auth BOGUS --context dev /tmp/kafka-introspect-test

# Without --auth at all (overrides the auto-inject by passing args that
# include --context but no --auth — wait, run_introspect auto-injects in
# that case. So the only way to test "no --auth" is via the raw bash call.
# Do that here instead.
no_auth_output="$(env -i PATH="$PATH" HOME="$HOME" bash -c "
  $ALL_DEV
  bash '$INTROSPECT' --context dev /tmp/kafka-introspect-test 2>&1
" || true)"
if grep -qF -- "usage:" <<<"$no_auth_output" \
   && ! grep -qF -- "missing required environment variables" <<<"$no_auth_output"; then
  PASS=$((PASS + 1))
  echo "PASS: refuses without --auth (usage shown, env-var check not reached)"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("expected usage refusal when --auth is missing. Output:
$no_auth_output")
  echo "FAIL: refuses without --auth (usage shown, env-var check not reached)"
fi

# --- MTLS env-var validation tests --------------------------------------------

# Under --auth MTLS, the cert vars (not SASL) are required. Refusal lists
# every missing cert var in one error.
mtls_missing_output="$(run_introspect "" \
  --auth MTLS --context dev /tmp/kafka-introspect-test 2>&1 || true)"
for v in CONTEXTS_DEV_BROKERS CONTEXTS_DEV_TLS_CERT CONTEXTS_DEV_TLS_CERTKEY CONTEXTS_DEV_TLS_CA; do
  if grep -q -- "$v" <<<"$mtls_missing_output"; then
    PASS=$((PASS + 1))
    echo "PASS: MTLS missing-list includes $v"
  else
    FAIL=$((FAIL + 1))
    FAILURES+=("MTLS missing-list should include $v. Output:
$mtls_missing_output")
    echo "FAIL: MTLS missing-list includes $v"
  fi
done

# Under --auth MTLS, the SASL vars must NOT be required (so a SASL-only
# env doesn't accidentally pass MTLS validation).
mtls_sasl_only_output="$(run_introspect \
  'export CONTEXTS_DEV_BROKERS=b CONTEXTS_DEV_SASL_USERNAME=u CONTEXTS_DEV_SASL_PASSWORD=p' \
  --auth MTLS --context dev /tmp/kafka-introspect-test 2>&1 || true)"
if grep -q "CONTEXTS_DEV_TLS_CERT" <<<"$mtls_sasl_only_output"; then
  PASS=$((PASS + 1))
  echo "PASS: MTLS doesn't accept SASL-only env (still demands TLS_CERT etc.)"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("MTLS validation should still demand TLS_CERT when only SASL vars are set. Output:
$mtls_sasl_only_output")
  echo "FAIL: MTLS doesn't accept SASL-only env (still demands TLS_CERT etc.)"
fi

# Cert paths must be absolute. A relative path is rejected with a named
# error.
MTLS_TMP="$(mktemp -d "${TMPDIR:-/tmp}/kafka-mtls-test-XXXXXX")"
trap 'rm -rf -- "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT" "$SANDBOX_BIN" "$FAKE_RUNTIME_COMPAT" "$INVOCATION_LOG_COMPAT" "$TMPOUT_COMPAT" "$MTLS_TMP"' EXIT
echo "stub-cert" > "$MTLS_TMP/client.crt"
echo "stub-key" > "$MTLS_TMP/client.key"
echo "stub-ca"  > "$MTLS_TMP/ca.crt"

assert_test \
  "MTLS rejects relative cert path with named error" \
  2 \
  "must be an absolute path" \
  "export CONTEXTS_DEV_BROKERS=b CONTEXTS_DEV_TLS_CERT=relative.crt CONTEXTS_DEV_TLS_CERTKEY=$MTLS_TMP/client.key CONTEXTS_DEV_TLS_CA=$MTLS_TMP/ca.crt" \
  --auth MTLS --context dev /tmp/kafka-introspect-test

assert_test \
  "MTLS rejects nonexistent cert path with named error" \
  2 \
  "file not found" \
  "export CONTEXTS_DEV_BROKERS=b CONTEXTS_DEV_TLS_CERT=/tmp/does-not-exist.crt CONTEXTS_DEV_TLS_CERTKEY=$MTLS_TMP/client.key CONTEXTS_DEV_TLS_CA=$MTLS_TMP/ca.crt" \
  --auth MTLS --context dev /tmp/kafka-introspect-test

# --- MTLS positive-path: cert mounts reach the runtime ------------------------
#
# When all cert vars are valid, introspect.sh should bind-mount each cert
# at <path>:<path>:ro,z and skip the SCRAM mechanism translation. The `:z`
# is required for SELinux hosts (Fedora/RHEL) — same reason the config-mount
# uses `:ro,z`; without it the container hits Permission denied reading
# the cert file even though the bind-mount is in place. Reuse the fake-
# runtime pattern from the SASL compat tests.

INVOCATION_LOG_MTLS="$(mktemp)"
TMPOUT_MTLS="$(mktemp -d "${TMPDIR:-/tmp}/kafka-introspect-XXXXXX")"
trap 'rm -rf -- "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT" "$SANDBOX_BIN" "$FAKE_RUNTIME_COMPAT" "$INVOCATION_LOG_COMPAT" "$TMPOUT_COMPAT" "$MTLS_TMP" "$INVOCATION_LOG_MTLS" "$TMPOUT_MTLS"' EXIT

env -i PATH="$PATH" HOME="$HOME" \
  INVOCATION_LOG_MTLS="$INVOCATION_LOG_MTLS" \
  bash -c "
    export CONTEXTS_DEV_BROKERS='b1:9093'
    export CONTEXTS_DEV_TLS_CERT='$MTLS_TMP/client.crt'
    export CONTEXTS_DEV_TLS_CERTKEY='$MTLS_TMP/client.key'
    export CONTEXTS_DEV_TLS_CA='$MTLS_TMP/ca.crt'
    export KAFKA_CONTAINER_RUNTIME='$FAKE_RUNTIME_COMPAT'
    export INVOCATION_LOG_COMPAT='$INVOCATION_LOG_MTLS'
    bash '$INTROSPECT' --auth MTLS --context dev --topic t1 --group g1 '$TMPOUT_MTLS'
  " > /dev/null 2>&1
mtls_invocation="$(awk '/---END---/{exit} {print}' "$INVOCATION_LOG_MTLS")"

assert_compat_line \
  "MTLS bind-mounts client cert at <path>:<path>:ro,z" \
  "ARG -v" \
  "$mtls_invocation"

for cert_path in "$MTLS_TMP/client.crt" "$MTLS_TMP/client.key" "$MTLS_TMP/ca.crt"; do
  assert_compat_line \
    "MTLS bind-mounts $cert_path :ro,z into the container (SELinux relabel marker required)" \
    "ARG ${cert_path}:${cert_path}:ro,z" \
    "$mtls_invocation"
done

# Under MTLS, no SASL_MECHANISM should be in the captured env (the SCRAM
# translation block in introspect.sh must be skipped). Because the parent
# shell exports nothing SASL-related, the ENV dump from the fake runtime
# should not contain SASL_MECHANISM at all.
if grep -q "ENV CONTEXTS_DEV_SASL_MECHANISM=" <<<"$mtls_invocation"; then
  FAIL=$((FAIL + 1))
  FAILURES+=("MTLS run unexpectedly carries CONTEXTS_DEV_SASL_MECHANISM. Captured:
$mtls_invocation")
  echo "FAIL: MTLS run does not carry SASL_MECHANISM"
else
  PASS=$((PASS + 1))
  echo "PASS: MTLS run does not carry SASL_MECHANISM"
fi

# Path-leak prevention: cert paths exported for OTHER contexts must NOT
# end up as -v mounts when the current --context only names dev. The
# OTHER context's cert vars get forwarded as env vars (via the same
# CONTEXTS_*_TLS_* pattern in the env-var filter), but no `-v` mount
# should reach argv for them — without the mount, kafkactl can't read
# them inside the container even though it sees the env var.
INVOCATION_LOG_LEAK="$(mktemp)"
TMPOUT_LEAK="$(mktemp -d "${TMPDIR:-/tmp}/kafka-introspect-XXXXXX")"
trap 'rm -rf -- "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT" "$SANDBOX_BIN" "$FAKE_RUNTIME_COMPAT" "$INVOCATION_LOG_COMPAT" "$TMPOUT_COMPAT" "$MTLS_TMP" "$INVOCATION_LOG_MTLS" "$TMPOUT_MTLS" "$INVOCATION_LOG_LEAK" "$TMPOUT_LEAK"' EXIT

# Stage two prod cert files at distinct paths.
PROD_TMP="$(mktemp -d "${TMPDIR:-/tmp}/kafka-mtls-prod-XXXXXX")"
trap 'rm -rf -- "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT" "$SANDBOX_BIN" "$FAKE_RUNTIME_COMPAT" "$INVOCATION_LOG_COMPAT" "$TMPOUT_COMPAT" "$MTLS_TMP" "$INVOCATION_LOG_MTLS" "$TMPOUT_MTLS" "$INVOCATION_LOG_LEAK" "$TMPOUT_LEAK" "$PROD_TMP"' EXIT
echo "prod-stub" > "$PROD_TMP/prod-client.crt"
echo "prod-stub" > "$PROD_TMP/prod-client.key"
echo "prod-stub" > "$PROD_TMP/prod-ca.crt"

env -i PATH="$PATH" HOME="$HOME" \
  INVOCATION_LOG_COMPAT="$INVOCATION_LOG_LEAK" \
  bash -c "
    export CONTEXTS_DEV_BROKERS='b1:9093'
    export CONTEXTS_DEV_TLS_CERT='$MTLS_TMP/client.crt'
    export CONTEXTS_DEV_TLS_CERTKEY='$MTLS_TMP/client.key'
    export CONTEXTS_DEV_TLS_CA='$MTLS_TMP/ca.crt'
    export CONTEXTS_PROD_BROKERS='prod:9093'
    export CONTEXTS_PROD_TLS_CERT='$PROD_TMP/prod-client.crt'
    export CONTEXTS_PROD_TLS_CERTKEY='$PROD_TMP/prod-client.key'
    export CONTEXTS_PROD_TLS_CA='$PROD_TMP/prod-ca.crt'
    export KAFKA_CONTAINER_RUNTIME='$FAKE_RUNTIME_COMPAT'
    bash '$INTROSPECT' --auth MTLS --context dev --topic t1 --group g1 '$TMPOUT_LEAK'
  " > /dev/null 2>&1
leak_invocation="$(awk '/---END---/{exit} {print}' "$INVOCATION_LOG_LEAK")"

# Prod cert paths must NOT appear in any -v argv when --context dev was passed.
if grep -qE "ARG -v ${PROD_TMP}/prod-" <<<"$leak_invocation" \
   || grep -qE "ARG ${PROD_TMP}/prod-.*:${PROD_TMP}/prod-.*:ro" <<<"$leak_invocation"; then
  FAIL=$((FAIL + 1))
  FAILURES+=("path-leak: prod cert paths appeared as -v mounts when --context dev was passed. Captured:
$leak_invocation")
  echo "FAIL: --context dev does not bind-mount prod cert paths"
else
  PASS=$((PASS + 1))
  echo "PASS: --context dev does not bind-mount prod cert paths"
fi

# But the dev cert paths SHOULD appear with the SELinux relabel marker (sanity).
if grep -qF "ARG ${MTLS_TMP}/client.crt:${MTLS_TMP}/client.crt:ro,z" <<<"$leak_invocation"; then
  PASS=$((PASS + 1))
  echo "PASS: --context dev still bind-mounts dev cert paths in the leak-prevention test"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("--context dev should still mount dev cert paths. Captured:
$leak_invocation")
  echo "FAIL: --context dev still bind-mounts dev cert paths in the leak-prevention test"
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
