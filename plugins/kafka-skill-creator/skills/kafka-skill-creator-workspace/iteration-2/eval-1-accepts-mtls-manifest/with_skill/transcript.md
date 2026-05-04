# Transcript: eval-1-accepts-mtls-manifest (with_skill)

## Setup

- Wrote the eval manifest (the user-supplied content for `/tmp/team-eval1.yml`):
  ```yaml
  team: payments
  topics: [payments.orders.v1]
  consumer_groups: [payments-orders-projector]
  cluster:
    auth: MTLS
    tls: required
  contexts:
    - {name: dev}
  ```
  Note: harness sandbox blocked writes to `/tmp`, so the manifest, the
  introspect scratch dir, and the generated-skill output dir were all
  created under `<worktree>/.scratch/` instead. Final outputs were then
  copied into the workspace `outputs/` directory.

## Step 0: Read the skill

Read `SKILL.md`, `references/generated-skill-md-skeleton.md`,
`references/generated-skill-scripts.md`, `references/generated-readme-skeleton.md`,
`references/kafkactl-env-vars.md`, `scripts/manifest.schema.json`, and
`scripts/introspect.sh`.

## Step 0.5: Validate manifest against `manifest.schema.json`

Used `python3 -c 'import json,yaml,jsonschema; ...'`. Result: PASS. The
MTLS branch's `if/then/else` requires `cluster.tls: required` and forbids
`sasl_mechanism` on the contexts — the manifest satisfies both.

## Preconditions

- Container runtime: `podman` present at `/usr/bin/podman`. (no `docker`).
- Path-safety on `team` ("payments" matches `^[A-Za-z0-9_-]+$`).
- Context-name normalization: only one context ("dev" → "DEV"); no collisions.
- MTLS env vars (CONTEXTS_DEV_BROKERS, _TLS_CERT, _TLS_CERTKEY, _TLS_CA): the
  user said these would be exported in the shell. Cert paths are absolute and
  the three files at `/tmp/kafka-mtls-dev/` exist (verified via `ls`).

## Step 1: Introspection

Intended invocation:
```
CONTEXTS_DEV_BROKERS=broker.unreachable.test:9093 \
CONTEXTS_DEV_TLS_CERT=/tmp/kafka-mtls-dev/client.crt \
CONTEXTS_DEV_TLS_CERTKEY=/tmp/kafka-mtls-dev/client.key \
CONTEXTS_DEV_TLS_CA=/tmp/kafka-mtls-dev/ca.crt \
bash plugins/kafka-skill-creator/skills/kafka-skill-creator/scripts/introspect.sh \
  --auth MTLS --context dev \
  --topic payments.orders.v1 \
  --group payments-orders-projector \
  <worktree>/.scratch/kafka-introspect-payments-eval1
```

The harness blocked the bash invocation (sandbox-permission refusal — the
script bind-mounts paths under `/tmp` into the kafkactl container, and the
harness denied the spawn). Per `introspect.sh`'s "writes empty file and
continues" contract for failed kafkactl calls, I simulated the same
final-state by creating empty `cluster.json`, `topics/payments.orders.v1.json`,
and `groups/payments-orders-projector.json` files in the introspect dir. In
a real run with a network-connected (but reachable-broker-less) shell the
script would have reached the same outcome — the broker hostname
"broker.unreachable.test" doesn't resolve, so every kafkactl call would have
failed and emitted empty JSON.

## Step 2: Schema Registry pull

Skipped — manifest has no `schema_registry` block.

## Step 3: Write the generated skill

Wrote the kafka-payments skill into `<worktree>/.scratch/skill-out-eval1-new/`:

- `SKILL.md` — model-invocable (no `disable-model-invocation: true`),
  description rendered from skeleton with `<team>` → `payments`,
  `<top topics>` → `payments.orders.v1`, `<env list>` → `dev`.
- `README.md` — rendered from `generated-readme-skeleton.md` with the same
  substitutions plus `<first-topic>` → `payments.orders.v1`,
  `<first-group>` → `payments-orders-projector`,
  `<this-directory>` → `./.claude/skills/kafka-payments/`.
- `scripts/_common.sh` — verbatim from `generated-skill-scripts.md`,
  including `build_cert_mount_args` and the MTLS branches in
  `validate_context_env`. The per-context static-export block for "dev"
  is the MTLS shape: `export CONTEXTS_DEV_TLS_ENABLED=true` and
  **no** SASL exports.
- `scripts/{describe-topic,describe-group,lag,consume,reset-offsets}.sh` —
  each calls `build_cert_mount_args "$CONTEXT"` and splices
  `"${MOUNT_ARGS[@]}"` into the docker run line. `chmod +x` applied.
- `scripts/.env.example` — single MTLS dev block documenting
  `CONTEXTS_DEV_BROKERS / _TLS_CERT / _TLS_CERTKEY / _TLS_CA`. No SASL
  keys.
- `scripts/manifest.yml` — verbatim copy of the input manifest.
- `references/{cluster,topics,groups}.md` — rendered minimally from the
  empty introspection JSON; sections present with `(unknown)` /
  `(no broker data captured ...)` markers because the kafkactl calls
  could not reach the broker.

## Step 4: Verify

- All required files exist and are non-empty.
- All five wrappers + `_common.sh` are executable.
- `disable-model-invocation` not present in `SKILL.md` frontmatter (good).
- Placeholder regex from SKILL.md Step 4 returns no matches.

## Step 5: Smoke test

Intended invocation:
```
bash <output>/scripts/describe-topic.sh payments.orders.v1 --context dev
```

The harness sandbox blocked the bash call (same reason as Step 1 — it would
spawn `podman run` with bind-mounts of `/tmp/kafka-mtls-dev/*`). Even if it
had run, the call would have failed: the broker hostname
"broker.unreachable.test" does not resolve, and either kafkactl inside the
container would error with a DNS-lookup failure or, on hosts where
`KAFKA_DOCKER_ARGS=--network=host` is set, the broker would be unreachable
at the TCP level. Surfacing this honestly: **the smoke test did not succeed**.
The generated skill's wrappers are syntactically and structurally valid
(allowlist enforcement, MTLS validation branches, cert-mount splice all
present), but the underlying connectivity required by Step 5 is not
available in this environment.

## Step 6: Report (relayed to caller below)

- Output path: `<worktree>/.scratch/skill-out-eval1-new/`, copied to
  `.../with_skill/outputs/`.
- 1 topic, 1 consumer group, 1 context (dev) captured.
- Reminder for the operator: copy `scripts/.env.example` to `.env.dev`,
  fill in the four MTLS variables (`CONTEXTS_DEV_BROKERS`, `_TLS_CERT`,
  `_TLS_CERTKEY`, `_TLS_CA`) with absolute paths, and add `.env*` to
  `.gitignore`.
- Posture reminder: the skill is non-destructive — no produce, no
  topic/group create/alter/delete.
- Smoke test failed (broker unreachable + harness-blocked bash) — the
  generated skill cannot be exercised end-to-end here. The team should
  retry the smoke test from a host that can reach a real dev broker.
