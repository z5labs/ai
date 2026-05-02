#!/usr/bin/env bash
# Lightweight tests for scripts/introspect.sh.
#
# Verifies the credential-routing contract from issue #42:
#   - The script accepts exactly one argument (output dir), never a connection string.
#   - All five libpq env vars (PGHOST/PGPORT/PGUSER/PGDATABASE/PGPASSWORD) are
#     required; the script refuses with a clear, complete missing-list when any
#     are unset.
#   - The refusal exits before any container is invoked, so these tests need
#     neither docker/podman nor a Postgres server.
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

# Run introspect.sh with a controlled environment. Args after the env spec are
# passed to the script. Captures exit code, stdout, and stderr.
run_introspect() {
  local env_spec="$1"; shift
  # Wipe every PG* var to start, then apply env_spec (a string of KEY=VAL pairs
  # separated by spaces, or empty for "no env vars set").
  env -i PATH="$PATH" HOME="$HOME" bash -c "
    $env_spec
    bash '$INTROSPECT' \"\$@\" 2> /tmp/introspect_stderr.\$\$ > /tmp/introspect_stdout.\$\$
    code=\$?
    cat /tmp/introspect_stdout.\$\$
    echo '---STDERR---'
    cat /tmp/introspect_stderr.\$\$
    rm -f /tmp/introspect_stdout.\$\$ /tmp/introspect_stderr.\$\$
    exit \$code
  " _ "$@"
}

# assert_test "name" expected_exit_code "expected_substring_in_output" run_args...
# Pass an empty string for expected_substring to skip the substring check.
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

# A complete env spec with all five vars set to dummy values. The script will
# proceed past validation and try to invoke the container runtime — for the
# arg-shape tests we *want* it to get past validation so we're testing the arg
# parser, not the env validator.
ALL_ENV='export PGHOST=h PGPORT=5432 PGUSER=u PGDATABASE=d PGPASSWORD=p'

# --- Argument-shape tests -----------------------------------------------------

assert_test \
  "rejects zero args with usage" \
  2 \
  "usage:" \
  "$ALL_ENV"

assert_test \
  "rejects two args (no longer accepts <connection-string> <output-dir>)" \
  2 \
  "usage:" \
  "$ALL_ENV" \
  "postgresql://u@h/d" "/tmp/out"

assert_test \
  "usage line shows only <output-dir>, not <connection-string>" \
  2 \
  "<output-dir>" \
  "$ALL_ENV"

# Make sure the old usage line is *gone* — guards against a partial revert.
output_check="$(run_introspect "$ALL_ENV" 2>&1 || true)"
if grep -qF "<connection-string>" <<<"$output_check"; then
  FAIL=$((FAIL + 1))
  FAILURES+=("usage line must not mention <connection-string>. Output:
$output_check")
  echo "FAIL: usage line must not mention <connection-string>"
else
  PASS=$((PASS + 1))
  echo "PASS: usage line does not mention <connection-string>"
fi

# --- Env-var validation tests -------------------------------------------------

assert_test \
  "refuses with missing-list when no PG* vars set" \
  2 \
  "missing required environment variables:" \
  "" \
  "/tmp/out"

# When everything except PGPASSWORD is set, the missing list must name PGPASSWORD
# specifically (not just "some env var") — the user should know exactly what to
# fix.
assert_test \
  "names PGPASSWORD specifically when only it is missing" \
  2 \
  "PGPASSWORD" \
  'export PGHOST=h PGPORT=5432 PGUSER=u PGDATABASE=d' \
  "/tmp/out"

assert_test \
  "names PGHOST specifically when only it is missing" \
  2 \
  "PGHOST" \
  'export PGPORT=5432 PGUSER=u PGDATABASE=d PGPASSWORD=p' \
  "/tmp/out"

assert_test \
  "names PGDATABASE specifically when only it is missing" \
  2 \
  "PGDATABASE" \
  'export PGHOST=h PGPORT=5432 PGUSER=u PGPASSWORD=p' \
  "/tmp/out"

# Multi-missing case: the script should list *all* missing vars in one shot, not
# fail-then-fix-then-fail-again. Check that two distinct missing vars both appear
# in the same error.
multi_output="$(run_introspect 'export PGUSER=u PGDATABASE=d' /tmp/out 2>&1 || true)"
if grep -q "PGHOST" <<<"$multi_output" && \
   grep -q "PGPORT" <<<"$multi_output" && \
   grep -q "PGPASSWORD" <<<"$multi_output"; then
  PASS=$((PASS + 1))
  echo "PASS: lists all missing vars in a single error"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("multi-missing test: expected PGHOST, PGPORT, and PGPASSWORD all in one error. Output:
$multi_output")
  echo "FAIL: lists all missing vars in a single error"
fi

# Empty (set-but-blank) env vars should be treated as missing — an empty
# PGPASSWORD would otherwise authenticate as no-password rather than refusing.
assert_test \
  "treats set-but-empty PGPASSWORD as missing" \
  2 \
  "PGPASSWORD" \
  'export PGHOST=h PGPORT=5432 PGUSER=u PGDATABASE=d PGPASSWORD=' \
  "/tmp/out"

# The error message must point users at credential helpers — that's the
# recommended path for populating these vars.
helper_output="$(run_introspect "" /tmp/out 2>&1 || true)"
if grep -qE "(op run|vault|direnv|credential helper)" <<<"$helper_output"; then
  PASS=$((PASS + 1))
  echo "PASS: refusal mentions credential-helper path"
else
  FAIL=$((FAIL + 1))
  FAILURES+=("refusal must mention credential helpers (op run / vault / direnv). Output:
$helper_output")
  echo "FAIL: refusal mentions credential-helper path"
fi

# --- Positive-path test: invocation shape -------------------------------------
#
# The refusal-path tests above never reach the docker invocation. This block
# stubs PG_CONTAINER_RUNTIME with a fake "container runtime" that just logs its
# arguments to a file, then asserts on what introspect.sh would have passed to
# `docker run`. Catches regressions in the -e VAR forwarding and the overall
# arg shape without needing a Postgres or container runtime.

FAKE_RUNTIME="$(mktemp)"
INVOCATION_LOG="$(mktemp)"
TMPOUT="$(mktemp -d)"
trap 'rm -rf "$FAKE_RUNTIME" "$INVOCATION_LOG" "$TMPOUT"' EXIT

cat > "$FAKE_RUNTIME" <<'FAKE'
#!/usr/bin/env bash
# Append every argument as one line, terminated by ---END--- so concurrent
# invocations stay separable. Drain stdin so heredoc-style queries don't break
# the script. Exit 0 so introspect.sh thinks each query succeeded.
{
  for a in "$@"; do printf '%s\n' "$a"; done
  echo "---END---"
} >> "$INVOCATION_LOG"
cat > /dev/null
exit 0
FAKE
chmod +x "$FAKE_RUNTIME"

# Run introspect.sh with the fake runtime and a rich set of libpq env vars
# (including non-original ones like PGSSLMODE/PGAPPNAME/PGOPTIONS) plus our
# internal PG_DOCKER_ARGS — which must NOT be forwarded as -e PG_DOCKER_ARGS.
env -i PATH="$PATH" HOME="$HOME" \
  INVOCATION_LOG="$INVOCATION_LOG" \
  bash -c "
    export PGHOST=db.internal PGPORT=5432 PGUSER=app PGDATABASE=orders PGPASSWORD=secret
    export PGSSLMODE=require PGAPPNAME=introspect-test PGOPTIONS='-c statement_timeout=5000'
    export PG_DOCKER_ARGS='--network=host'
    export PG_CONTAINER_RUNTIME='$FAKE_RUNTIME'
    bash '$INTROSPECT' '$TMPOUT'
  " > /dev/null 2>&1

# Pick the first invocation block (all subsequent ones share the same -e shape).
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

# The five originally-required libpq vars must all be forwarded.
for v in PGHOST PGPORT PGUSER PGDATABASE PGPASSWORD; do
  check_invocation "forwards -e $v to runtime" "$v"
done
# Extended libpq vars must also be forwarded — this is what restores the
# connection-surface coverage that was previously available via the DSN.
for v in PGSSLMODE PGAPPNAME PGOPTIONS; do
  check_invocation "forwards extended libpq var -e $v" "$v"
done
# Internal config vars must NOT be forwarded as -e flags. PG_DOCKER_ARGS is
# our own thing (its *value* gets word-split into the runtime command line as
# extra args, but the *name* must never appear as a forwarded env var).
check_invocation_absent "does NOT forward internal PG_DOCKER_ARGS as -e" "PG_DOCKER_ARGS"
check_invocation_absent "does NOT forward internal PG_CONTAINER_RUNTIME as -e" "PG_CONTAINER_RUNTIME"
check_invocation_absent "does NOT forward PSQL_IMAGE as -e" "PSQL_IMAGE"

# PG_DOCKER_ARGS value must reach the runtime. Word-split inserts --network=host
# as a standalone argument.
check_invocation "applies PG_DOCKER_ARGS value to runtime" "--network=host"

# The image must appear as a positional arg, and the script must NOT pass any
# old-style connection string.
check_invocation "passes PSQL_IMAGE as positional arg" "docker.io/alpine/psql:17.7"
check_invocation_absent "does NOT pass any postgresql:// DSN" "postgresql://app:secret@db.internal:5432/orders"

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
