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

Each script reads a `.env.<ctx>` file at runtime (resolution order: `--env-file PATH` → `KAFKA_ENV_FILE` → `./.env`) and passes `--context <ctx>` to kafkactl. Cluster shape (broker list, SASL mechanism, TLS, Schema Registry auth) is defined entirely through kafkactl's `CONTEXTS_<NAME>_*` env-var convention — the static fields come from `_common.sh`'s baked-in `export` block, the secrets come from the `.env.<ctx>` file. There is no separate kafkactl config file; everything kafkactl needs is in the environment when the wrapper exec's the container.

### One-time setup

    cp .claude/skills/kafka-<team>/scripts/.env.example .env.dev
    # edit .env.dev — set CONTEXTS_DEV_BROKERS, CONTEXTS_DEV_SASL_USERNAME, CONTEXTS_DEV_SASL_PASSWORD,
    # and (if the team uses Schema Registry) CONTEXTS_DEV_SCHEMAREGISTRY_*.

Repeat for each context. Add the chosen filename(s) to `.gitignore` so secrets don't get committed:

    /.env
    /.env.*

### Per-invocation usage

    bash .claude/skills/kafka-<team>/scripts/describe-topic.sh <topic> --context dev
    bash .claude/skills/kafka-<team>/scripts/lag.sh <group> --context prod --env-file .env.prod
    bash .claude/skills/kafka-<team>/scripts/reset-offsets.sh <group> --topic <topic> --to-earliest --dry-run --context staging

If `--context` doesn't match a context the loaded `.env` populated, kafkactl's lookup fails closed (the env vars for that context aren't set, so kafkactl won't authenticate). Same fail-closed property postgres-skill-creator relies on.

## Reference docs

- `references/cluster.md` — broker list, controller, cluster id (snapshot at generation time)
- `references/topics.md` — per-topic config, partitions, schema reference
- `references/groups.md` — per-group subscriptions and members
- `references/schemas/<topic>.json` — Schema Registry latest-version dumps (when configured)

These are written **once at generation time**. Re-run `/kafka-skill-creator` when topics drift or new groups appear.
```

The description copy above is the canonical template. Don't paraphrase it for "tone", don't add team-specific color, don't drop the "even if they don't say 'Kafka' explicitly" clause to make it shorter — the wording is what triggers reliably. Substitution only.
