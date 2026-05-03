#!/usr/bin/env bash
# Seed the Kafka + Karapace fixture with realistic state for the e2e eval.
#
# Creates three topics with mixed partition counts and config overrides,
# registers an Avro schema for each, produces messages so consumer groups
# have something to commit, and leaves one group with non-zero lag so the
# generated `lag.sh` wrapper has something to bite into.
#
# Idempotent: re-running after a successful seed is a no-op. Topic create
# skips if the topic exists; schema POST swallows 409; produce checks the
# topic's current message count and only emits the missing tail (or skips
# entirely if at-or-above target); drain checks the group's committed
# offset and only consumes the delta. After `down -v`, a fresh `up.sh`
# re-seeds from scratch.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Auto-detect runtime; honor the same KAFKA_CONTAINER_RUNTIME override
# the skill itself uses. Match the skill's image pin so seed and
# introspection exercise the same kafkactl version.
RUNTIME="${KAFKA_CONTAINER_RUNTIME:-}"
if [ -z "$RUNTIME" ]; then
  if command -v docker >/dev/null 2>&1; then
    RUNTIME=docker
  elif command -v podman >/dev/null 2>&1; then
    RUNTIME=podman
  else
    echo "error: neither docker nor podman on PATH" >&2
    exit 1
  fi
fi
KAFKACTL_IMAGE="${KAFKACTL_IMAGE:-docker.io/deviceinsight/kafkactl:v5.18.0-scratch}"
CURL_IMAGE="${CURL_IMAGE:-docker.io/curlimages/curl:8.11.1}"

# Per-context env that kafkactl needs. The plain-text values match what
# env.sh exports plus the static cluster-shape bits the skill normally
# fills in from the manifest. SASL_MECHANISM uses kafkactl's casing
# (`scram-sha512` not `SCRAM-SHA-512`) — see references/kafkactl-env-vars.md.
export CONTEXTS_DEV_BROKERS="${CONTEXTS_DEV_BROKERS:-localhost:9092}"
export CONTEXTS_DEV_SASL_ENABLED=true
export CONTEXTS_DEV_SASL_MECHANISM=scram-sha512
export CONTEXTS_DEV_SASL_USERNAME="${CONTEXTS_DEV_SASL_USERNAME:-app}"
export CONTEXTS_DEV_SASL_PASSWORD="${CONTEXTS_DEV_SASL_PASSWORD:-app-secret}"
export CONTEXTS_DEV_TLS_ENABLED=false
SR_URL="${CONTEXTS_DEV_SCHEMAREGISTRY_URL:-http://localhost:8081}"
SR_USER="${CONTEXTS_DEV_SCHEMAREGISTRY_USERNAME:-sruser}"
SR_PASS="${CONTEXTS_DEV_SCHEMAREGISTRY_PASSWORD:-sr-secret}"

# kafkactl rejects `--context dev` without a `dev` entry in config.yml,
# even when CONTEXTS_DEV_* env vars are set. Generate an empty-bodied
# context entry; the env vars overlay broker/credential values at runtime.
KAFKACTL_CONFIG_DIR="$(mktemp -d -t kafkactl-cfg-XXXXXX)"
trap 'rm -rf -- "$KAFKACTL_CONFIG_DIR"' EXIT
cat > "$KAFKACTL_CONFIG_DIR/config.yml" <<EOF
contexts:
  dev: {}
EOF

# Forward the kafkactl-shaped env vars into the container, same filter
# the skill's _common.sh uses.
FORWARD_PATTERN='^(CONTEXTS_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS$)'
KAFKACTL_ENV_ARGS=()
while IFS= read -r var; do
  [ -z "$var" ] && continue
  KAFKACTL_ENV_ARGS+=(-e "$var")
done < <(compgen -e | grep -E "$FORWARD_PATTERN" || true)

kafkactl() {
  "$RUNTIME" run --rm -i --network=host \
    -v "$KAFKACTL_CONFIG_DIR/config.yml:/.config/kafkactl/config.yml:ro,z" \
    "${KAFKACTL_ENV_ARGS[@]}" \
    "$KAFKACTL_IMAGE" \
    --context dev \
    "$@"
}

create_topic() {
  local name="$1" partitions="$2"
  if kafkactl describe topic "$name" >/dev/null 2>&1; then
    echo "  topic $name already exists, skipping create"
    return
  fi
  echo "  creating topic $name (partitions=$partitions)"
  kafkactl create topic "$name" --partitions "$partitions" --replication-factor 1
}

set_topic_config() {
  local name="$1" key="$2" value="$3"
  echo "  setting $key=$value on $name"
  kafkactl alter topic "$name" --config "$key=$value"
}

register_schema() {
  local subject="$1" avsc_path="$2"
  echo "  registering schema for subject $subject"
  # Karapace expects the schema as a JSON-encoded string in the "schema"
  # field, not as a JSON object — easy to get wrong with a heredoc.
  local payload
  payload=$(jq -Rs --argjson dummy 0 '{schema: ., schemaType: "AVRO"}' < "$avsc_path")
  # Swallow 409 (subject already has this schema). Other failures bubble.
  local code
  code=$("$RUNTIME" run --rm -i --network=host \
    "$CURL_IMAGE" \
    -sS -o /dev/null -w '%{http_code}' \
    -u "${SR_USER}:${SR_PASS}" \
    -H 'Content-Type: application/vnd.schemaregistry.v1+json' \
    --data-binary "@-" \
    "${SR_URL}/subjects/${subject}/versions" <<<"$payload")
  case "$code" in
    200|409) ;;
    *)
      echo "error: schema registration for $subject failed with HTTP $code" >&2
      exit 1
      ;;
  esac
}

# Sum of newest offsets across every partition of a topic. Equals the
# total number of messages the topic has ever held (modulo retention).
# Returns 0 if the topic doesn't exist or has no partitions.
topic_message_count() {
  local topic="$1"
  kafkactl describe topic "$topic" --output json 2>/dev/null \
    | jq -r '[.Partitions[]?.newestOffset] | add // 0'
}

# Sum of committed offsets across every partition of a topic, for a
# given consumer group. Equals the high-water mark the group has
# acknowledged. Returns 0 if the group hasn't committed for this topic.
group_committed_total() {
  local group="$1" topic="$2"
  kafkactl describe consumer-group "$group" --output json 2>/dev/null \
    | jq -r --arg topic "$topic" '[.Topics[]? | select(.Name == $topic) | .Partitions[]?.consumerOffset] | add // 0'
}

# Idempotent produce: takes a TARGET ABSOLUTE message count, not a delta.
# If the topic already has that many messages (or more), this is a no-op.
# Otherwise produce only the missing tail. This keeps re-running seed.sh
# (or up.sh after a down-less restart) a true no-op.
produce_messages() {
  local topic="$1" target="$2"
  local existing
  existing=$(topic_message_count "$topic")
  if [ "$existing" -ge "$target" ]; then
    echo "  $topic already has $existing messages (≥ target $target), skipping produce"
    return
  fi
  local first=$((existing + 1))
  echo "  producing $((target - existing)) messages to $topic (offsets $first..$target)"
  # kafkactl auto-detects the topic's registered Avro schema and serializes
  # each line accordingly, so the per-line JSON has to match the schema.
  # Per-topic payload generators keep this fixture in sync with schemas/*.avsc.
  local i ts
  for i in $(seq "$first" "$target"); do
    ts="$(( 1700000000000 + i * 1000 ))"
    case "$topic" in
      payments.orders.v1)
        printf '{"order_id":"order-%s","customer_id":"cust-%s","amount_cents":%s,"currency":"USD","created_at_epoch_ms":%s}\n' "$i" "$((i % 10))" "$((i * 199))" "$ts"
        ;;
      payments.refunds.v1)
        # `reason` is a union ["null","string"]. kafkactl's Avro JSON
        # encoder doesn't accept the canonical type-tag wrapper for
        # non-null union members, so we always emit null. The schema
        # registration still exercises the union path; produce just
        # avoids the union-with-value form.
        printf '{"refund_id":"refund-%s","order_id":"order-%s","amount_cents":%s,"reason":null,"created_at_epoch_ms":%s}\n' "$i" "$i" "$((i * 50))" "$ts"
        ;;
      internal.audit.v1)
        printf '{"event_id":"evt-%s","actor":"system","action":"login","resource":"user-%s","occurred_at_epoch_ms":%s}\n' "$i" "$i" "$ts"
        ;;
      *)
        echo "error: produce_messages: no payload generator for topic $topic" >&2
        return 1
        ;;
    esac
  done | kafkactl produce "$topic"
}

# Idempotent drain: takes a TARGET ABSOLUTE committed offset, not a count.
# If the group has already committed at least that high for this topic,
# this is a no-op. Otherwise consume only the delta needed. Critical:
# kafkactl --max-messages N is "stop after N received in this invocation",
# not "stop at offset N", so we compute the delta explicitly — passing the
# absolute target with no committed-offset accounting would loop forever
# waiting for messages past the topic's tail.
drain_group() {
  local group="$1" topic="$2" target="$3"
  local committed
  committed=$(group_committed_total "$group" "$topic")
  if [ "$committed" -ge "$target" ]; then
    echo "  group $group at offset $committed for $topic (≥ target $target), skipping drain"
    return
  fi
  local need=$((target - committed))
  echo "  draining $need messages for group $group from $topic (committed=$committed, target=$target)"
  # kafkactl rejects --group with --exit (mutually exclusive). --from-beginning
  # only applies when the group has no committed offset — on a re-run it
  # silently resumes from `committed`.
  kafkactl consume "$topic" --group "$group" --from-beginning --max-messages "$need" >/dev/null
}

echo "==> creating topics"
create_topic payments.orders.v1 6
create_topic payments.refunds.v1 3
create_topic internal.audit.v1 1

echo "==> setting DYNAMIC_TOPIC_CONFIG overrides"
# These overrides are what topics.md's source filter must surface.
# internal.audit.v1 intentionally has none — its Notable config table
# should render empty.
set_topic_config payments.orders.v1 retention.ms 604800000
set_topic_config payments.orders.v1 cleanup.policy delete
set_topic_config payments.refunds.v1 cleanup.policy compact

echo "==> registering Avro schemas"
register_schema payments.orders.v1-value "$SCRIPT_DIR/schemas/payments.orders.v1.avsc"
register_schema payments.refunds.v1-value "$SCRIPT_DIR/schemas/payments.refunds.v1.avsc"
register_schema internal.audit.v1-value "$SCRIPT_DIR/schemas/internal.audit.v1.avsc"

echo "==> producing initial messages (target absolute counts; no-op if already met)"
# Targets are absolute totals, not deltas. produce_messages skips when the
# topic is already at-or-above the target, so a re-run after a successful
# seed touches no offsets.
produce_messages payments.orders.v1 100
produce_messages payments.refunds.v1 100
produce_messages internal.audit.v1 10

echo "==> draining and committing consumer groups (idempotent; skips if already committed)"
# payments-orders-projector: drain to offset 100 → lag 0.
drain_group payments-orders-projector payments.orders.v1 100
# payments-refunds-replayer: commits at 100, then we extend the topic to 150
# below so the group's committed offset lags LEO.
drain_group payments-refunds-replayer payments.refunds.v1 100

echo "==> extending payments.refunds.v1 to 150 so payments-refunds-replayer has lag=50"
# Target 150 absolute — the prior produce_messages call sat at 100. After this
# step the topic has 150 messages but the group committed at 100, so lag=50.
produce_messages payments.refunds.v1 150

echo
echo "seed complete."
