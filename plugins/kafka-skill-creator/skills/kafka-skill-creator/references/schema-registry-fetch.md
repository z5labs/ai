# Schema Registry fetch (Step 2)

The verbatim bash for Step 2 of the parent `SKILL.md` — pulling the latest schema for each manifest topic out of Schema Registry. Read this file when executing Step 2; the parent SKILL.md only summarizes the contract.

## Why this lives outside argv

Two layers of privacy:

1. **Host shell:** `${!varname}` indirect expansion looks up `CONTEXTS_<UPPER>_SCHEMAREGISTRY_*` for the chosen context, then `export`s the values under stable names (`SR_URL`, `SR_USER`, `SR_PASS`). `export` is a shell builtin — no fork, no argv exposure.
2. **Container:** `docker run -e VARNAME` (no `=value`) tells docker to read the named var from the host's environment and forward it to the container. Only the variable *name* lands in `docker run`'s argv. Inside the container, `curl -K -` reads its `--user` config from stdin via a `printf | curl` pipe, so the password never lands in `curl`'s argv either.

If you collapse this back to `docker run -e SR_PASS="$value" ... curl -u "$user:$pass"`, the password ends up in argv at *both* layers. That's the trap this snippet is shaped to avoid.

## Inputs the caller must set before running

| Variable | Source |
|---|---|
| `TEAM` | The manifest's `team` field. Validated path-safe in Precondition 4 of the parent SKILL.md. |
| `CONTEXT` | The same context name used in Step 1 (the manifest's first context). |
| `TOPICS=()` | Bash array of the manifest's `topics:` values (one entry per topic). |

The snippet itself sets two more variables; they are NOT caller inputs but the caller can influence them via env:

| Variable | Set by snippet | Override via |
|---|---|---|
| `RUNTIME` | Auto-detected with the same `command -v docker / podman` block `introspect.sh` uses. | `KAFKA_CONTAINER_RUNTIME` (snippet honors this if set in the env). |
| `EXTRA_ARGS` | Word-split from `KAFKA_DOCKER_ARGS` and spliced into the `docker run` invocation, identical to `introspect.sh`. The common case is `--network=host` on Linux when Schema Registry is on `localhost`. | `KAFKA_DOCKER_ARGS` (default: empty). |

## Snippet

```bash
# Pick the container runtime the same way introspect.sh does. `command -v`
# is more portable than `which` and doesn't fork a subshell on a hit.
if [ -n "${KAFKA_CONTAINER_RUNTIME:-}" ]; then RUNTIME="$KAFKA_CONTAINER_RUNTIME"
elif command -v docker >/dev/null 2>&1; then RUNTIME=docker
elif command -v podman >/dev/null 2>&1; then RUNTIME=podman
else echo "neither docker nor podman found on PATH" >&2; exit 1
fi

CURL_IMAGE="${CURL_IMAGE:-docker.io/curlimages/curl:8.11.1}"

# Word-split KAFKA_DOCKER_ARGS into an array, identical to introspect.sh.
# SKILL.md documents this var as affecting "both the generator (... SR fetch
# step) and the generated wrappers" — the most common need is `--network=host`
# on Linux when Schema Registry is on localhost, which only works if it
# reaches THIS docker run too, not just the kafkactl one.
read -r -a EXTRA_ARGS <<< "${KAFKA_DOCKER_ARGS:-}"

# Derive the per-context var names. <UPPER> is the context name uppercased
# with hyphens replaced by underscores — e.g. "dev-us-east" -> "DEV_US_EAST".
# `tr` rather than `${var^^}` keeps this portable to macOS's default Bash 3.2.
upper="$(printf '%s' "$CONTEXT" | tr '[:lower:]' '[:upper:]')"
upper="${upper//-/_}"
sr_url_var="CONTEXTS_${upper}_SCHEMAREGISTRY_URL"
sr_user_var="CONTEXTS_${upper}_SCHEMAREGISTRY_USERNAME"
sr_pass_var="CONTEXTS_${upper}_SCHEMAREGISTRY_PASSWORD"

# Re-export the values under stable names. `export` is a shell builtin and
# does not fork, so the password never enters any process's argv at this step.
export SR_URL="${!sr_url_var}"
export SR_USER="${!sr_user_var}"
export SR_PASS="${!sr_pass_var}"
export TOPIC

# introspect.sh creates topics/ and groups/ but not schemas/, so make it here.
mkdir -p "/tmp/kafka-introspect-${TEAM}/schemas"

for TOPIC in "${TOPICS[@]}"; do
  safe="$(printf '%s' "$TOPIC" | sed 's/[^A-Za-z0-9._-]/_/g')"
  out_path="/tmp/kafka-introspect-${TEAM}/schemas/${safe}.json"

  # Stage the response in a temp file and only mv it to the final path on
  # success. Plain `> $out_path` would truncate the destination *before*
  # the curl call ran, so a 404 (or any other failure) would leave behind
  # an empty file that downstream rendering can't distinguish from a
  # successfully-fetched empty schema. With this shape, "file exists"
  # reliably means "schema was fetched".
  tmp_out="$(mktemp "${TMPDIR:-/tmp}/sr-fetch.XXXXXX")"

  # `-e SR_*` (no `=value`) tells docker to read each var from the host env
  # and forward it to the container. Argv only carries the var NAME.
  # Inside the container, `curl -K -` reads `--user` from stdin so the
  # password never lands in argv there either.
  #
  # `${EXTRA_ARGS[@]}` is spliced in the same shape introspect.sh uses, so
  # any KAFKA_DOCKER_ARGS the operator set for the kafkactl invocations
  # also applies here.
  if "$RUNTIME" run --rm -i \
       "${EXTRA_ARGS[@]}" \
       -e SR_URL -e SR_USER -e SR_PASS -e TOPIC \
       "$CURL_IMAGE" \
       sh -c 'printf "user = %s:%s\n" "$SR_USER" "$SR_PASS" \
              | curl -sf -K - "$SR_URL/subjects/$TOPIC-value/versions/latest"' \
       > "$tmp_out"
  then
    mv -- "$tmp_out" "$out_path"
  else
    rm -f -- "$tmp_out"
    echo "warning: schema fetch failed for $TOPIC; skipping (no file written)" >&2
  fi
done
```

## Output shape

The response is the standard Schema Registry envelope — a `subject`, `version`, `id`, `schemaType` (`AVRO` / `PROTOBUF` / `JSON`), and `schema` string with the actual definition. **Persist it verbatim** as `<topic>.json` (one extension, regardless of `schemaType`) — both at `/tmp/kafka-introspect-${TEAM}/schemas/${safe}.json` here and, in Step 3 of the parent SKILL.md, at `<output>/references/schemas/<topic>.json`. The verbatim envelope keeps the metadata (`id`, `version`, `schemaType`) alongside the schema body so the model can read whichever it needs without a per-format dispatch step in the generator. Don't extract the inner `schema` string into format-specific files (`.avsc` / `.proto`) — that's a future enhancement, and doing it inconsistently is the bigger trap.

## Failure handling

If Schema Registry returns 404 for a subject, skip it and note the absence in `references/topics.md` for that topic — partial coverage is better than aborting the whole generation.
