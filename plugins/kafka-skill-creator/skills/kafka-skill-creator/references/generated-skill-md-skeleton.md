# Generated-skill SKILL.md skeleton

The `SKILL.md` that ships inside every generated `kafka-<team>/`. Read this file when executing Step 3 of the parent `SKILL.md` (specifically the `<output>/SKILL.md` substep). The substitution rules live in the parent SKILL.md (description template + bullet list rendering); this file carries the verbatim body.

The frontmatter `description` is generated from a fixed template — substitution only, no paraphrasing. Two runs against the same manifest must produce byte-identical descriptions. **Do not** add `disable-model-invocation: true` to the frontmatter; the generated skill is meant to fire on natural-language prompts, not require slash invocation.

## Substitutions to make

- `<team>` — the manifest's `team` field, verbatim.
- `<top topics>` — the **first up to 5** entries from the manifest's `topics:` list, in manifest order, joined by `", "`.
- `<env list>` — the manifest's `contexts[].name` values, in manifest order, joined by `" / "` (e.g. `dev / staging / prod`).
- `<bullet list>` for owned topics and groups — emit one `- ` line per item under `topics:` and `consumer_groups:`.
- `<list of context names>` in the Environments section — same as `<env list>` above.

## Skeleton

```markdown
---
name: kafka-<team>
description: Read-only investigation of the <team> team's Kafka topics (<top topics>) and consumer groups across <env list>. Use whenever the user asks to peek at messages, check consumer-group lag, describe a topic's config, or reset an offset for a <team>-owned topic, even if they don't say "Kafka" explicitly. Non-destructive only — does not produce, alter, or delete.
---

This skill knows the <team> team's Kafka topics and consumer groups, and provides fixed-shape wrappers for investigation tasks. Posture is **non-destructive**: reads and offset resets only. No produce, no topic create/alter/delete, no consumer-group delete.

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

- Topics: <bullet list>
- Consumer groups: <bullet list>

Attempting to invoke any of the scripts with a topic or group outside this list is refused at script entry. To extend the list, edit the team's `manifest.yml` and re-run `/kafka-skill-creator --manifest …` — do **not** edit the embedded array in scripts directly, it gets overwritten on regeneration.

## Environments

The team operates against these contexts: <list of context names>.

Each script reads an env file at runtime, resolved in this order (first match wins): `--env-file PATH` (per-invocation), `KAFKA_ENV_FILE` (whole-shell), `./.env` (cwd). There is no automatic `.env.<ctx>` lookup — the convention is to *name* per-environment files `.env.dev` / `.env.prod` / etc. and select one with `--env-file` or `KAFKA_ENV_FILE`. Each script also passes `--context <ctx>` to kafkactl, which selects which `CONTEXTS_<UPPER>_*` env vars (loaded from the env file) apply.

Cluster shape (broker list, SASL mechanism, TLS, Schema Registry auth) is defined entirely through kafkactl's `CONTEXTS_<NAME>_*` env-var convention — the static fields come from `_common.sh`'s baked-in `export` block, the secrets (and, for MTLS contexts, the cert/key/CA paths) come from the env file. The wrapper does mint a one-line `config.yml` per invocation and mount it into the kafkactl container — kafkactl rejects `--context <name>` for contexts not pre-declared in a config file, even when the env-var overlays cover every field — but the config carries no addresses, credentials, or other operator-relevant settings; it's a stub for kafkactl's argv parser, not a thing the operator edits. Everything kafkactl actually uses is in the environment when the wrapper exec's the container.

For MTLS contexts, the wrapper additionally bind-mounts each cert path read-only into the container at the same path the env var declares. That's why the cert vars must be absolute paths to existing host files: docker requires absolute mount sources, and kafkactl reads the cert at the path it sees in its own env. Only the active `--context`'s vars (and the bare default-context shorthand) reach kafkactl in the container — the env-forwarding filter is scoped to `CONTEXTS_<ACTIVE_UPPER>_*`, so paths exported for other contexts don't get forwarded as env vars OR mounted. A prod cert path string can't leak into a dev container even informationally.

### One-time setup

Run from this skill's directory (whatever path the generator wrote it to). The keys to populate per context depend on `cluster.auth` in `scripts/manifest.yml`:

    cp scripts/.env.example .env.dev
    # SASL_SCRAM contexts: set CONTEXTS_DEV_BROKERS, CONTEXTS_DEV_SASL_USERNAME,
    # CONTEXTS_DEV_SASL_PASSWORD.
    # MTLS contexts:       set CONTEXTS_DEV_BROKERS, CONTEXTS_DEV_TLS_CERT,
    # CONTEXTS_DEV_TLS_CERTKEY, CONTEXTS_DEV_TLS_CA — each must be an absolute
    # path to a file the host can read; the wrapper bind-mounts each :ro into
    # the kafkactl container at the same path.
    # Schema Registry (any auth): also set CONTEXTS_DEV_SCHEMAREGISTRY_* if the
    # manifest declares schema_registry.

Repeat for each context. Add the chosen filename(s) to `.gitignore` so secrets don't get committed:

    /.env
    /.env.*

### Per-invocation usage

Paths below are relative to this skill's directory. If you're invoking from elsewhere, use the full path the generator wrote (e.g. `./.claude/skills/kafka-<team>/scripts/...`, or `plugins/team-<team>/skills/kafka-<team>/scripts/...` if the skill lives in a plugin tree).

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
```

The description copy above is the canonical template. Don't paraphrase it for "tone", don't add team-specific color, don't drop the "even if they don't say 'Kafka' explicitly" clause to make it shorter — the wording is what triggers reliably. Substitution only.
