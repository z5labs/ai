# End-to-end fixture for kafka-skill-creator

A throwaway, single-broker Kafka + Schema Registry fixture for exercising
the full `kafka-skill-creator` generation pipeline against a real cluster.
The skill's other evals are behavioral (refusal paths, deterministic
template substitution against synthetic JSON) — this one is the only
guard against regressions in `introspect.sh`, the SR fetch path, or any
of the generated wrappers.

## Requirements

- `docker` + `docker compose`, or `podman` + `podman compose`
- `jq` on the host (used by `up.sh` to poll healthcheck status)
- Linux. macOS is not supported yet — `--network=host` semantics on
  Docker Desktop are flaky enough that this fixture is Linux-only for
  now. Tracked in [#78](https://github.com/z5labs/ai/issues/78).

## Quick start

```bash
cd plugins/kafka-skill-creator/skills/kafka-skill-creator/evals/e2e
./up.sh                    # bring up the cluster + seed
source env.sh              # export CONTEXTS_DEV_*, KAFKA_DOCKER_ARGS

# Drive the skill against the fixture:
/kafka-skill-creator --manifest manifest.yml --output /tmp/skill-out

# When done:
./down.sh
```

`./up.sh` is idempotent across `down`-less restarts (the broker's storage
is formatted only once per volume; `down -v` wipes the volume and the
next `up` re-formats and re-seeds).

## What the fixture stands up

- **Apache Kafka 3.9** in single-node KRaft mode with a SASL_PLAINTEXT
  client listener on `localhost:9092`. SCRAM-SHA-512 with two users —
  `admin` (inter-broker) and `app` (clients) — provisioned at format
  time via `kafka-storage format --add-scram`.
- **Karapace** (the Apache 2.0 Schema Registry) on `localhost:8081`
  with HTTP basic auth (`sruser:sr-secret`). Karapace runs with
  `network_mode: host` so it reaches the broker the same way kafkactl
  on the host does.

The seed creates three topics with mixed partition counts and config
overrides, two consumer groups (one stable with lag 0, one with non-zero
lag), and registers an Avro schema per topic against Karapace.

| Topic                  | Partitions | DYNAMIC_TOPIC_CONFIG overrides       |
|------------------------|-----------:|--------------------------------------|
| `payments.orders.v1`   |          6 | `retention.ms`, `cleanup.policy`     |
| `payments.refunds.v1`  |          3 | `cleanup.policy=compact`             |
| `internal.audit.v1`    |          1 | none — exercises empty-overrides path|

| Consumer group               | State after seed                           |
|------------------------------|--------------------------------------------|
| `payments-orders-projector`  | Stable, fully consumed (lag 0)             |
| `payments-refunds-replayer`  | Stable, lag > 0 (50 msgs after last commit)|

## Why these choices

- **`apache/kafka` not `confluentinc/cp-kafka`**: Apache 2.0 vs CSL.
  Avoids the risk that a future Confluent license change makes the
  fixture non-reproducible.
- **Karapace not Confluent Schema Registry**: same reason — Apache 2.0,
  drop-in compatible with the Confluent SR REST API.
- **`cluster.tls: none` (cleartext SASL_PLAINTEXT)**: server TLS would
  require shipping a CA the kafkactl container can trust, which the
  skill has no plumbing for yet. TLS coverage is tracked in
  [#79](https://github.com/z5labs/ai/issues/79).
- **`network_mode: host` on Karapace**: lets it reach `localhost:9092`
  with the same address the host's kafkactl uses. Cross-platform
  alternatives are explored in [#78](https://github.com/z5labs/ai/issues/78).

## Files

```
docker-compose.yml      # service definitions
env.sh                  # `source` to export CONTEXTS_DEV_* + KAFKA_DOCKER_ARGS
up.sh                   # compose up + wait-for-healthy + seed
down.sh                 # compose down -v
seed.sh                 # creates topics, registers schemas, produces, commits
manifest.yml            # the manifest the e2e eval feeds the skill
schemas/*.avsc          # Avro schemas registered to Karapace
kafka/server.properties # broker config (mounted into the kafka container)
kafka/admin.properties  # SASL config used by the broker's healthcheck
kafka/start.sh          # entrypoint: format-with-SCRAM, then start broker
karapace/authfile.json  # checked-in scrypt-hashed creds for sruser:sr-secret
```

## Running the e2e eval

The eval entry lives at `evals/evals.json` (id `7`, name
`e2e-real-cluster`). The skill-creator iteration harness expects:

1. The fixture is up (`./up.sh` succeeded).
2. `env.sh` has been sourced in the parent shell so `CONTEXTS_DEV_*`
   and `KAFKA_DOCKER_ARGS` are inherited by spawned subagents.
3. The orchestrator invokes the eval via the normal skill-creator loop.

There is no CI for this eval today — it's local-only because the full
iteration loop burns more model tokens than fits a per-PR job. A
script-level CI variant (no model in the loop) is tracked in
[#77](https://github.com/z5labs/ai/issues/77).

## Rotating the Karapace credentials

The checked-in `karapace/authfile.json` was generated with:

```bash
podman run --rm ghcr.io/aiven-open/karapace:6.1.4 \
  karapace_mkpasswd -u sruser -a scrypt sr-secret
```

If you change the SR password, regenerate the file and update three
places that hardcode `sr-secret`:

- `env.sh`'s `CONTEXTS_DEV_SCHEMAREGISTRY_PASSWORD`
- the **karapace** healthcheck `curl -u sruser:sr-secret …` line in
  `docker-compose.yml` (the broker healthcheck uses `admin.properties`
  with the admin SCRAM creds, not the SR creds — different user)
- `seed.sh`'s `SR_PASS` default
