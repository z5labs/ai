---
name: kafka-payments
description: Read-only investigation of the payments team's Kafka topics (payments.orders.v1, payments.refunds.v1, internal.audit.v1) and consumer groups across dev. Use whenever the user asks to peek at messages, check consumer-group lag, describe a topic's config, or reset an offset for a payments-owned topic, even if they don't say "Kafka" explicitly. Non-destructive only — does not produce, alter, or delete.
---

This skill knows the payments team's Kafka topics and consumer groups, and provides fixed-shape wrappers for investigation tasks. Posture is **non-destructive**: reads and offset resets only. No produce, no topic create/alter/delete, no consumer-group delete.

## What this skill is for

- Peek at recent messages on a topic (consume.sh)
- Describe a topic's config / partition layout / schema (describe-topic.sh)
- Describe a consumer group's members, subscriptions, lag (describe-group.sh, lag.sh)
- Reset offsets for a consumer group, with --dry-run (reset-offsets.sh)

## What this skill does NOT do

- Produce messages, including for "test data" or "smoke tests"
- Create / alter / delete topics
- Delete consumer groups
- Change ACLs or cluster config
- Operate on topics or groups not owned by this team (see manifest.yml)

## Owned topics and consumer groups

This skill operates only on the team's owned set, embedded in each script's allowlist:

- Topics:
  - payments.orders.v1
  - payments.refunds.v1
  - internal.audit.v1
- Consumer groups:
  - payments-orders-projector
  - payments-refunds-replayer

Attempting to invoke any of the scripts with a topic or group outside this list is refused at script entry. To extend the list, edit the team's `manifest.yml` and re-run `/kafka-skill-creator --manifest …` — do **not** edit the embedded array in scripts directly, it gets overwritten on regeneration.

## Environments

The team operates against these contexts: dev.

Each script reads an env file at runtime, resolved in this order (first match wins): `--env-file PATH` (per-invocation), `KAFKA_ENV_FILE` (whole-shell), `./.env` (cwd). There is no automatic `.env.<ctx>` lookup — the convention is to *name* per-environment files `.env.dev` / `.env.prod` / etc. and select one with `--env-file` or `KAFKA_ENV_FILE`. Each script also passes `--context <ctx>` to kafkactl, which selects which `CONTEXTS_<UPPER>_*` env vars (loaded from the env file) apply.

Cluster shape (broker list, SASL mechanism, TLS, Schema Registry auth) is defined entirely through kafkactl's `CONTEXTS_<NAME>_*` env-var convention — the static fields come from `_common.sh`'s baked-in `export` block, the secrets come from the env file. There is no separate kafkactl config file; everything kafkactl needs is in the environment when the wrapper exec's the container.

### One-time setup

Run from this skill's directory (whatever path the generator wrote it to):

    cp scripts/.env.example .env.dev
    # edit .env.dev — set CONTEXTS_DEV_BROKERS, CONTEXTS_DEV_SASL_USERNAME, CONTEXTS_DEV_SASL_PASSWORD,
    # and (if the team uses Schema Registry) CONTEXTS_DEV_SCHEMAREGISTRY_*.

Repeat for each context. Add the chosen filename(s) to `.gitignore` so secrets don't get committed:

    /.env
    /.env.*

### Per-invocation usage

Paths below are relative to this skill's directory. If you're invoking from elsewhere, use the full path the generator wrote (e.g. `./.claude/skills/kafka-payments/scripts/...`, or `plugins/team-payments/skills/kafka-payments/scripts/...` if the skill lives in a plugin tree).

    bash scripts/describe-topic.sh <topic> --context dev
    bash scripts/lag.sh <group> --context prod --env-file .env.prod
    bash scripts/reset-offsets.sh <group> --topic <topic> --to-earliest --dry-run --context staging

If `--context` doesn't match a context the loaded env file populated, kafkactl's lookup fails closed (the env vars for that context aren't set, so kafkactl won't authenticate). Same fail-closed property postgres-skill-creator relies on.

## Reference docs

- `references/cluster.md` — broker list, controller, cluster id (snapshot at generation time)
- `references/topics.md` — per-topic config, partitions, schema reference
- `references/groups.md` — per-group subscriptions and members
- `references/schemas/<topic>.json` — Schema Registry latest-version dumps (when configured)

These are written **once at generation time**. Re-run `/kafka-skill-creator` when topics drift or new groups appear.
