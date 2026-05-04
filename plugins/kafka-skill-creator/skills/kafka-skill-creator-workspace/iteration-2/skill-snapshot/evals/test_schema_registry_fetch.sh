#!/usr/bin/env bash
# Lightweight test for the bash snippet in references/schema-registry-fetch.md.
#
# The snippet is documentation: at Step 2 of SKILL.md the LLM extracts and
# runs it. Without an executable test, regressions slip through silently.
# That's exactly what happened in the PR #67 review — the snippet stopped
# threading KAFKA_DOCKER_ARGS into the docker run invocation, even though
# SKILL.md still documented it as affecting "both the generator (... SR
# fetch step) and the generated wrappers."
#
# This test extracts the snippet, runs it against a fake runtime that
# captures argv, and asserts the documented contract: KAFKA_DOCKER_ARGS
# reaches the runtime; -e SR_* are still forwarded by name; and the
# password value never lands in the runtime's argv.

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SNIPPET_MD="$HERE/../references/schema-registry-fetch.md"

if [ ! -f "$SNIPPET_MD" ]; then
  echo "FATAL: schema-registry-fetch.md not found at $SNIPPET_MD" >&2
  exit 2
fi

# Extract the first ```bash ... ``` block from the markdown.
SNIPPET="$(awk '/^```bash$/{f=1; next} /^```$/{if(f) exit} f' "$SNIPPET_MD")"
if [ -z "$SNIPPET" ]; then
  printf 'FATAL: no fenced bash code block found in %s\n' "$SNIPPET_MD" >&2
  exit 2
fi

PASS=0
FAIL=0
FAILURES=()

# Unique per-run scratch dir so the snippet's mkdir target doesn't collide
# with anything else under /tmp and we can clean up cleanly.
TEAM_SUFFIX="srtest-$$-$(date +%s%N 2>/dev/null || date +%s)"
SCHEMA_DIR="/tmp/kafka-introspect-${TEAM_SUFFIX}"

FAKE_RUNTIME="$(mktemp)"
INVOCATION_LOG="$(mktemp)"
trap 'rm -rf "$FAKE_RUNTIME" "$INVOCATION_LOG" "$SCHEMA_DIR"' EXIT

cat > "$FAKE_RUNTIME" <<'FAKE'
#!/usr/bin/env bash
# Log argv one line per arg, terminated by ---END---.
{
  for a in "$@"; do printf '%s\n' "$a"; done
  echo "---END---"
} >> "$INVOCATION_LOG"
# Drain stdin so a real docker -i shape doesn't hang or SIGPIPE.
cat > /dev/null 2>&1 || true
# Non-zero exit forces the snippet's "fetch failed" branch, so we don't
# write a stub file at the snippet's hard-coded out_path.
exit 1
FAKE
chmod +x "$FAKE_RUNTIME"

export TEAM="$TEAM_SUFFIX"
export CONTEXT=dev
# Indexed arrays don't propagate through `export` (bash 5+ silently allows
# the form, older bashes warn — and the export does nothing either way).
# eval'ing the snippet runs in this shell, so plain assignment is enough.
TOPICS=(payments.orders.v1)
export KAFKA_CONTAINER_RUNTIME="$FAKE_RUNTIME"
export KAFKA_DOCKER_ARGS="--network=host"
export CONTEXTS_DEV_SCHEMAREGISTRY_URL=http://sr.test:8081
export CONTEXTS_DEV_SCHEMAREGISTRY_USERNAME=app
export CONTEXTS_DEV_SCHEMAREGISTRY_PASSWORD=hunter2
export INVOCATION_LOG

# Run the snippet with stdin closed so docker -i mimicry doesn't read the
# test script's tty.
eval "$SNIPPET" </dev/null > /dev/null 2>&1 || true

# Helper: pin the first invocation block (everything before the first
# ---END--- marker).
first_invocation="$(awk '/---END---/{exit} {print}' "$INVOCATION_LOG")"

assert_in() {
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

assert_not_in() {
  local name="$1" pattern="$2"
  if grep -qF -- "$pattern" <<<"$first_invocation"; then
    FAIL=$((FAIL + 1))
    FAILURES+=("$name: pattern [$pattern] must NOT appear in invocation. Captured:
$first_invocation")
    echo "FAIL: $name"
  else
    PASS=$((PASS + 1))
    echo "PASS: $name"
  fi
}

# The KAFKA_DOCKER_ARGS value must reach the runtime as a standalone arg,
# not be silently dropped. This is the regression the review caught.
assert_in "applies KAFKA_DOCKER_ARGS value to runtime (--network=host)" "--network=host"

# -e SR_* (no =value) is the privacy-preserving form: docker reads each
# var from the host env and only its NAME goes into argv. All four must
# survive any future edit that changes how EXTRA_ARGS is spliced in.
for v in SR_URL SR_USER SR_PASS TOPIC; do
  assert_in "forwards -e $v to runtime" "$v"
done

# The password VALUE must never appear in argv. The snippet's whole shape
# (env-var-by-name, curl -K - via stdin) exists to prevent this; if a
# regression collapsed it back to `-e SR_PASS=$value` or `--user $u:$p`,
# the value would leak here.
assert_not_in "password value not present in runtime argv" "hunter2"

# The container image must be passed positionally so docker actually has
# something to run.
assert_in "passes CURL_IMAGE positionally" "docker.io/curlimages/curl:8.11.1"

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
