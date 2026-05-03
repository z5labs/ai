# kafka-skill-creator

A Claude Code plugin that generates a per-team Kafka research skill (`kafka-<team>`) from a manifest. The generated skill bakes the team's owned topics, consumer groups, and Schema Registry definitions into reference docs the model can consult, plus fixed-shape script wrappers for self-service investigation work.

## What it produces

After running, you get a skill at `<output>/` (default `./.claude/skills/kafka-<team>/`) containing:

- `SKILL.md` — what Claude reads when the skill triggers. Model-invocable by default (no `disable-model-invocation`) so it fires on natural-language prompts like "what's lag on the orders projector?", not just on `/kafka-<team>` slash invocations.
- `README.md` — human-facing quickstart with copy-paste examples for each wrapper, the `.env`-per-environment workflow, and the regeneration command. This is the file an engineer reads when onboarding to the team's Kafka tooling.
- `scripts/_common.sh` — shared bootstrap: env-file resolution, allowlist enforcement, container runtime selection, env-var forwarding.
- `scripts/consume.sh` — peek/tail messages on an owned topic, with Schema Registry deserialization when configured.
- `scripts/describe-topic.sh` — config and partition layout for an owned topic.
- `scripts/describe-group.sh` — members, subscriptions, and lag for an owned consumer group.
- `scripts/lag.sh` — convenience over `describe-group.sh` that surfaces only the lag-relevant fields.
- `scripts/reset-offsets.sh` — offset reset against an owned consumer group, with `--dry-run` and no bypass for kafkactl's active-member protection.
- `scripts/kafkactl-config.yml` — committed (no-secrets) kafkactl config naming each context.
- `scripts/manifest.yml` — verbatim copy of the input manifest, for transparency and re-generation.
- `scripts/.env.example` — one block per declared context documenting the `CONTEXTS_<NAME>_*` keys to populate.
- `references/{cluster,topics,groups}.md` — markdown summaries of the introspection dumps.
- `references/schemas/<topic>.json` — Schema Registry latest-version dumps (when configured).

The skill is **team-specific** (which topics and groups it can touch is fixed at generation time) and **environment-agnostic** (the same scripts run against dev / staging / prod by selecting a different `.env.<context>` at runtime).

## Posture (non-destructive only)

The generated skill ships **reads + offset resets**. It deliberately omits:

- Producing messages (no test-data writes)
- Topic create / alter / delete
- Consumer group delete
- ACL / cluster config changes

Offset resets are non-destructive at the Kafka layer (they change a consumer's read position, not the data) and gated by kafkactl's built-in refusal against active groups. The wrappers do not pass `--allow-active-members` or any equivalent bypass.

## v1 scope

- **Auth**: SASL/SCRAM, optionally with Schema Registry HTTP basic auth.
- **Container**: `deviceinsight/kafkactl:v5.18.0-scratch` runs every kafkactl invocation (pinned for reproducibility); the host needs only `docker` or `podman`.
- **Deferred** (each tracked as a follow-up issue):
  - SASL/PLAIN — [#63](https://github.com/z5labs/ai/issues/63)
  - mTLS — [#64](https://github.com/z5labs/ai/issues/64)
  - OAUTHBEARER (OIDC) — [#65](https://github.com/z5labs/ai/issues/65)
  - End-to-end eval against a containerized Kafka + Schema Registry — [#66](https://github.com/z5labs/ai/issues/66)

A manifest declaring an unsupported auth value is rejected with a one-line pointer to the matching deferred issue.

## Installation

From a Claude Code session:

```
/plugin marketplace add z5labs/ai
/plugin install kafka-skill-creator@z5labs-ai
```

## Generating a skill

Two invocation shapes:

### Manifest mode (preferred)

```
/kafka-skill-creator --manifest path/to/manifest.yml [--output <skill-dir>]
```

The manifest declares the team, owned topics and consumer groups, cluster auth shape, and contexts (one entry per environment). See `skills/kafka-skill-creator/scripts/manifest.example.yml` for a commented template and `skills/kafka-skill-creator/scripts/manifest.schema.json` for the validation rules.

### Manual-flag mode

For tiny teams that don't want a manifest file:

```
/kafka-skill-creator \
  --team payments \
  --topic payments.orders.v1 --topic payments.refunds.v1 \
  --group payments-orders-projector \
  --context dev --context staging --context prod \
  [--schema-registry] \
  [--output <skill-dir>]
```

Both modes build the same in-memory manifest and run the same generation pipeline.

### Output location

`--output PATH` (alias `--out`) chooses where the generated skill is written. Default is `./.claude/skills/kafka-<team>/`. The override exists so a team building their own plugin can land the generated skill directly inside it — for example:

```
/kafka-skill-creator --manifest team.yml --output plugins/team-payments/skills/kafka-payments/
```

The leaf directory is overwritten on every run; treat any local edits to generated files as ephemeral. If the output path lands inside a plugin tree, you may need to register the new skill in the plugin's `plugin.json` — this generator emits skill files only, not plugin manifests.

## Connection details and credential routing

All connection details (broker addresses, SASL credentials, Schema Registry credentials) are read from environment variables matching kafkactl's `CONTEXTS_<NAME>_*` convention:

| Variable | Purpose |
|---|---|
| `CONTEXTS_<CTX>_BROKERS` | Whitespace-separated `host:port` list. |
| `CONTEXTS_<CTX>_SASL_USERNAME` | SASL username for this context. |
| `CONTEXTS_<CTX>_SASL_PASSWORD` | SASL password — never seen by the model, read directly into the kafkactl container. |
| `CONTEXTS_<CTX>_SCHEMAREGISTRY_URL` | (Optional) Schema Registry endpoint. |
| `CONTEXTS_<CTX>_SCHEMAREGISTRY_USERNAME` | (Optional, basic auth) SR username. |
| `CONTEXTS_<CTX>_SCHEMAREGISTRY_PASSWORD` | (Optional, basic auth) SR password. |

Where `<CTX>` is the context's `name` value, uppercased and with `-` replaced by `_`. See `skills/kafka-skill-creator/references/kafkactl-env-vars.md` for the full convention.

If any required variable is unset for any declared context, the generator stops and lists the missing variables. It will not prompt for them, accept them inline, or fall back to a default — the same posture `postgres-skill-creator` settled on in [#42](https://github.com/z5labs/ai/issues/42).

### Pairing with a credential helper

Because the generator and the generated skill both read from the environment, they compose with any pre-authenticated tool you already use to manage secrets — 1Password CLI, HashiCorp Vault, `gcloud`, `direnv`-loaded `.env` files. Recommended pattern:

```
op run --env-file=kafka.env -- claude
```

…where `kafka.env` declares `CONTEXTS_DEV_SASL_PASSWORD=op://Vault/Kafka/dev-password`, etc. `op run` resolves the secrets at startup, exports them into the subprocess environment, and tears them down on exit.

### Other requirements

- `docker` or `podman` on `PATH`.
- `jq` on `PATH` — the generated `lag.sh` filters kafkactl's JSON output through `jq` to surface only the lag-relevant fields. The other wrappers don't need it (they emit kafkactl's JSON unfiltered), so a host without `jq` can still run `consume.sh`, `describe-topic.sh`, `describe-group.sh`, and `reset-offsets.sh` — `lag.sh` is the only one that fails closed when `jq` is absent.
- The container can reach the brokers and Schema Registry. On Linux when these are on `localhost`, set `KAFKA_DOCKER_ARGS=--network=host` so the container shares the host network namespace.

## Regenerating an installed skill

Manifests drift — new topics, new groups, new contexts. To pull in changes, edit the manifest and re-run:

```
/kafka-skill-creator --manifest team.yml [--output <same-path-as-before>]
```

This **overwrites** the existing skill directory in place. That is intentional — stale allowlists mislead the model. Keep team-specific guidance in a sibling file or a top-level `CLAUDE.md`, not in generated files.

You do **not** need to reinstall the plugin to regenerate the skill. The plugin is the generator; the generated `kafka-<team>/` skill is the output.

## `.env`-per-environment workflow

The same logical cluster is often deployed across multiple environments. The generated wrappers resolve a `.env` file in this order, first match wins:

1. `--env-file PATH` flag
2. `KAFKA_ENV_FILE` env var
3. `./.env` in the current working directory

If `--env-file` or `KAFKA_ENV_FILE` is set, the path must exist or the wrapper exits with an error. If neither is set and `./.env` doesn't exist, no env file is loaded — the wrapper relies on whatever's in the inherited environment.

```
# explicit per-invocation
bash .claude/skills/kafka-payments/scripts/lag.sh payments-orders-projector --context prod --env-file .env.prod

# whole-shell
export KAFKA_ENV_FILE=.env.staging
bash .claude/skills/kafka-payments/scripts/describe-topic.sh payments.orders.v1 --context staging
```

Mismatch between the loaded `.env` and the `--context` value fails closed: kafkactl looks up `CONTEXTS_<CTX>_*` for the `<CTX>` you passed, and if that prefix isn't populated by the env file, no credentials are present. User discipline gates which environment a destructive-feeling op like `reset-offsets.sh` runs against — same model `postgres-skill-creator` uses.

## Runtime overrides

| Variable | Purpose |
|---|---|
| `KAFKACTL_IMAGE` | Container image for kafkactl. Default: `docker.io/deviceinsight/kafkactl:v5.18.0-scratch` (pinned). Override to track a different version, pull from a private registry, or work around a kafkactl regression. |
| `CURL_IMAGE` | Container image for Schema Registry pulls. Default: `docker.io/curlimages/curl:8.11.1` (pinned). Override the same way as `KAFKACTL_IMAGE`. Only used when the manifest declares a `cluster.schema_registry` block. |
| `KAFKA_DOCKER_ARGS` | Extra args appended to `<runtime> run`. Common case on Linux: `KAFKA_DOCKER_ARGS=--network=host`. |
| `KAFKA_CONTAINER_RUNTIME` | Force `docker` or `podman` rather than auto-detecting. |
| `KAFKA_ENV_FILE` | Default `.env` path for the wrappers (overridden by `--env-file`). |

If you set `KAFKACTL_IMAGE` during generation, export the same value in any session that uses the generated skill so the runtime image matches.

## Development

Lightweight evals live under `skills/kafka-skill-creator/evals/`:

- `evals/test_introspect.sh` — shell-level tests for `introspect.sh`'s argument shape, env-var validation, and forwarded-env filter. Run with `bash evals/test_introspect.sh` from the skill directory; no Kafka, Schema Registry, or container runtime needed (the script exercises every refusal path before a container would be spawned, then uses a stubbed runtime to capture the positive-path invocation).
- `evals/evals.json` — skill-level behavioral evals exercising the `SKILL.md` instructions themselves (refusal paths, posture, output-path override).

A full end-to-end eval loop with a containerized Kafka + Schema Registry fixture is tracked in [#66](https://github.com/z5labs/ai/issues/66).
