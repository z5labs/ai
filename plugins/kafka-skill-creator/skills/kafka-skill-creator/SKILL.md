---
name: kafka-skill-creator
description: Generate a project-level skill (`kafka-<team>`) that gives a team self-service, non-destructive access to its owned Kafka topics and consumer groups across dev/staging/prod, by introspecting the cluster + Schema Registry from a manifest and baking the results into reference docs. Use whenever the user asks to scaffold a kafka-<team> skill, regenerate one after a manifest change, or set up Kafka tooling for a specific team. Skip when the user already has an installed kafka-<team> skill and is asking about *using* it (consume / describe / lag / reset-offsets) rather than (re)building it; that's the generated skill's job, not this generator's. Also skip when the user asks for produce / topic-create / ACL changes — this generator is non-destructive by design and would refuse those at runtime anyway.
disable-model-invocation: true
argument-hint: "--manifest <path-to-manifest.yml> [--output <skill-dir>] | --team <team> --topic <T> --group <G> --context <ctx> [--output <skill-dir>]"
---

Generate a project-level skill that captures one team's Kafka topics, consumer groups, and Schema Registry definitions as reference docs the model can consult during ad-hoc investigation work. The generated skill knows only the team's owned topics and groups — it is **not** a general-purpose Kafka CLI wrapper.

## Posture (non-destructive only)

The generated skill ships fixed wrappers for **reads + offset resets**. It deliberately omits:

- Producing messages (no `kafkactl produce`)
- Topic create / alter / delete
- Consumer group delete
- ACL / config changes

Offset resets are included because they are non-destructive at the Kafka layer (they change a consumer's read position, not the data) and are a real developer-testing use case. They are gated by kafkactl's built-in refusal to reset against an active consumer group; the generated `reset-offsets.sh` does not pass `--allow-active-members` or any equivalent bypass.

State this posture in the generated `SKILL.md` so the line doesn't drift toward "producing test messages is also developer testing."

## v1 scope

- **Auth**: SASL/SCRAM only (with optional Schema Registry HTTP basic auth).
- **Deferred**: SASL/PLAIN ([#63]), mTLS ([#64]), OAUTHBEARER ([#65]), end-to-end eval against a real cluster ([#66]).

When a manifest specifies an auth value other than `SASL_SCRAM`, refuse with a one-line pointer to the matching deferred-auth issue. Do not silently accept and degrade.

[#63]: https://github.com/z5labs/ai/issues/63
[#64]: https://github.com/z5labs/ai/issues/64
[#65]: https://github.com/z5labs/ai/issues/65
[#66]: https://github.com/z5labs/ai/issues/66

## Inputs

Two invocation shapes; pick one per run:

### Manifest mode (preferred)

```
/kafka-skill-creator --manifest <path-to-manifest.yml> [--output <skill-dir>]
```

Read `scripts/manifest.example.yml` for the file shape and `scripts/manifest.schema.json` for the validation rules. The manifest names the team, lists owned topics and consumer groups, declares the cluster auth shape, and lists every environment (context) the team operates against.

### Manual-flag mode (convenience for tiny teams)

```
/kafka-skill-creator \
  --team <team> \
  --topic <T> [--topic <T> ...] \
  --group <G> [--group <G> ...] \
  --context <ctx> [--context <ctx> ...] \
  [--sasl-mechanism SCRAM-SHA-512] \
  [--schema-registry] \
  [--output <skill-dir>]
```

Build an in-memory manifest of the same shape as the file form and continue down the same generation path. **If both `--manifest` and any of the manual flags are supplied, refuse** — pick one.

### Output location

`--output PATH` (alias `--out`) chooses where the generated skill is written. Defaults to `./.claude/skills/kafka-<team>/`. The primary use case for the override is a team building their own plugin: pointing `--output` at e.g. `plugins/team-payments/skills/kafka-payments/` lands the generated skill directly inside the plugin tree, no copy step.

The path's parent directory must already exist; create the leaf directory yourself. If the leaf already exists, **delete it before writing** — overwrite is intentional, manifests drift and stale references mislead. Before deleting, validate that the resolved leaf segment is non-empty and not `/`, `.`, `..`, `~`, or `*` (basic sanity guards against catastrophic deletes from a malformed `--output`).

## Preconditions

Stop and refuse if any of the following are unmet:

1. **Container runtime** — `docker` or `podman` on `PATH` (auto-detected, or override via `KAFKA_CONTAINER_RUNTIME`).
2. **Manifest validates** against `scripts/manifest.schema.json`. v1 hard-fails on `cluster.auth` values other than `SASL_SCRAM` and points at the matching deferred issue.
3. **Path-safety on `team`** — must match `^[A-Za-z0-9_-]+$`. The `--output` path itself is taken at face value (it's an explicit operator choice), but `team` flows into generated frontmatter and the default output path segment, so it stays restricted.
4. **Per-context env vars are populated** for every context declared in the manifest. For each `<ctx>`, derive `<UPPER>` by uppercasing the name and replacing `-` with `_`, then check that all three are set and non-empty:
   - `CONTEXTS_<UPPER>_BROKERS`
   - `CONTEXTS_<UPPER>_SASL_USERNAME`
   - `CONTEXTS_<UPPER>_SASL_PASSWORD`

   And, if the manifest's `cluster.schema_registry` block is present, also:
   - `CONTEXTS_<UPPER>_SCHEMAREGISTRY_URL`

   And, if `cluster.schema_registry.auth: basic`, also:
   - `CONTEXTS_<UPPER>_SCHEMAREGISTRY_USERNAME`
   - `CONTEXTS_<UPPER>_SCHEMAREGISTRY_PASSWORD`

   If **any** are unset for **any** declared context, stop. Print the full missing list and tell the user to export them — directly, or via a credential helper they already use (`op run --env-file=…`, `vault`, `direnv`, `gcloud`). Do **not** prompt for them, accept them inline, or invent placeholders.

   The reason is the same one postgres-skill-creator hardened in [#42](https://github.com/z5labs/ai/issues/42): secrets must reach tools out-of-band, never through model context. The libpq-equivalent here is kafkactl's documented `CONTEXTS_<NAME>_*` convention — see `references/kafkactl-env-vars.md`.

5. **No connection-string positional argument**. `/kafka-skill-creator "kafka://user:pass@broker:9092"` is rejected; do not parse the password out of the URL even silently. The skill takes flags only.

## High-level workflow

1. Parse and validate inputs (manifest or flags); refuse on conflict or missing values per Preconditions.
2. Resolve the output directory; validate path safety on the leaf.
3. Validate per-context env vars (Precondition 4).
4. Run introspection against the **first** context declared in the manifest. The manifest is shared across environments; one context's introspection is enough for the references because schemas should match across environments. (If they don't, the team has a bigger problem than this skill.)
5. Generate the skill files at the resolved output path.
6. Verify the generated skill (file existence, no leftover `<...>` placeholders, scripts marked executable).
7. Smoke test: invoke the generated `describe-topic.sh` against the first manifest topic using the same context's env vars. Surface failure; do not claim success.
8. Report what was written and how to use it.

## Step 1: Run introspection

```bash
bash <skill-dir>/scripts/introspect.sh \
  --context <first-context-name> \
  --topic <T> [--topic <T> ...] \
  --group <G> [--group <G> ...] \
  /tmp/kafka-introspect-<team>
```

`<skill-dir>` is wherever this skill is installed (the directory containing the `SKILL.md` you are reading). Use an absolute path.

`introspect.sh` re-validates the env vars itself, so a missing variable surfaces with the same exit-2 refusal even if Preconditions somehow passed. Output layout:

- `cluster.json` — broker list and cluster metadata (kafkactl `get brokers -o json`).
- `topics/<topic>.json` — per-topic config and partition layout (kafkactl `describe topic -o json`).
- `groups/<group>.json` — per-group members, subscriptions, lag (kafkactl `describe consumer-group -o json`).

If any individual call fails, `introspect.sh` writes an empty file for that target and continues. Partial coverage is better than aborting the whole generation.

## Step 2: Pull schemas (Schema Registry, when configured)

If the manifest has a `cluster.schema_registry` block, fetch the latest schema for each topic's `<topic>-value` subject. Use `curl` inside a pinned container so credentials are forwarded as env vars rather than baked into argv. The `CURL_IMAGE` env var overrides the default for private-registry users:

```bash
CURL_IMAGE="${CURL_IMAGE:-docker.io/curlimages/curl:8.11.1}"
"$RUNTIME" run --rm -i \
  -e SR_USER="$CONTEXTS_<UPPER>_SCHEMAREGISTRY_USERNAME" \
  -e SR_PASS="$CONTEXTS_<UPPER>_SCHEMAREGISTRY_PASSWORD" \
  -e SR_URL="$CONTEXTS_<UPPER>_SCHEMAREGISTRY_URL" \
  "$CURL_IMAGE" \
  -sf -K - "${SR_URL}/subjects/${TOPIC}-value/versions/latest" \
  <<< 'user-config-from-stdin: see below'
```

Use `curl -K -` (read config from stdin) to keep `--user $SR_USER:$SR_PASS` out of `docker run`'s host-side argv. The stdin content is `user = ${SR_USER}:${SR_PASS}` — do **not** put the password on the command line.

Save the response to `/tmp/kafka-introspect-<team>/schemas/<topic>.json` (the response includes a `schemaType` field and a `schema` string; the `schema` field's value is the actual Avro/Protobuf/JSON Schema definition, which you'll write into the generated `references/schemas/` in Step 3).

If a subject is not registered (Schema Registry returns 404), skip it and note the absence in `references/topics.md` for that topic.

## Step 3: Write the generated skill

Create these files under `<output>/` (the resolved output directory). Substitute the `<...>` placeholders with real content from the manifest and the introspection dumps.

### `<output>/SKILL.md`

The frontmatter `name` is what gets registered with Claude Code. The `description` is what determines triggering — name the team, mention the topic domain, and list the most prominent topic names so the model recognizes the right context.

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

Each script reads a `.env.<ctx>` file at runtime (resolution order: `--env-file PATH` → `KAFKA_ENV_FILE` → `./.env`) and passes `--context <ctx>` to kafkactl. The dbname-equivalent (the cluster identity) is fixed per context in `kafkactl-config.yml`; the secrets come from the env file.

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

Tailor the `description` to drive triggering for this team — generic phrasings ("a Kafka skill") undertrigger.

### `<output>/scripts/_common.sh`, `<output>/scripts/describe-topic.sh`, and the four sibling wrappers

Read `references/generated-skill-scripts.md` for the verbatim bash bodies. Substitute the `<...>` placeholders with the team's topic and group lists from the manifest.

`_common.sh` owns the shared bootstrap: env-file resolution (`--env-file PATH` → `KAFKA_ENV_FILE` → `./.env`, with explicit-path-must-exist semantics), allowlist enforcement (`require_allowed`), per-context env-var validation, container runtime selection, and the kafkactl-shaped env-var forwarding filter (`^(CONTEXTS_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS$)`).

Each wrapper sources `_common.sh`, parses its own flags, calls `require_allowed` against `ALLOWED_TOPICS` or `ALLOWED_GROUPS`, validates the chosen context's env vars, then `exec`s kafkactl in the container. Per-script flag and subcommand specifications:

- **`consume.sh <topic> --context <ctx> [--from-beginning] [--max N] [--partition P] [--env-file PATH]`** — `kafkactl consume "$TOPIC" --context "$CONTEXT" --output json --exit` plus `--from-beginning`, `--max-messages N`, `--partitions P` when those flags are present. No `--key-encoding` / `--value-encoding` overrides — kafkactl picks Schema Registry deserialization automatically when SR is configured.
- **`describe-topic.sh <topic> --context <ctx> [--env-file PATH]`** — `kafkactl describe topic "$TOPIC" --context "$CONTEXT" --output json`. (Worked example in `references/generated-skill-scripts.md`.)
- **`describe-group.sh <group> --context <ctx> [--env-file PATH]`** — `kafkactl describe consumer-group "$GROUP" --context "$CONTEXT" --output json`.
- **`lag.sh <group> --context <ctx> [--env-file PATH]`** — same as describe-group.sh but pipes through `jq` to extract lag-relevant fields only.
- **`reset-offsets.sh <group> --topic <T> --to-earliest|--to-latest|--to-offset N [--dry-run] --context <ctx> [--env-file PATH]`** — `kafkactl reset offset --group "$GROUP" --topic "$TOPIC" --context "$CONTEXT"` plus exactly one `--to-*` selector and `--dry-run` if requested. **Forbidden flags** (do not accept, do not pass through): `--allow-active-members`, `--all-topics`, `--execute-yes`. There is no `--force` / `--bypass`. Exactly one of `--to-earliest` / `--to-latest` / `--to-offset` must be present; refuse with usage otherwise.

Each script `chmod +x` after writing. Each script's allowlist re-validation is fail-closed: an off-allowlist topic or group name exits 2 with a named error pointing at `manifest.yml` and the regeneration command.

### `<output>/scripts/.env.example`

A commented template per declared context. Pre-fill the comments with the keys (no values — values are environment-specific):

```
# kafka-<team> .env per environment.
#
# Copy this file to .env.<context> (or just .env), uncomment the keys you
# want to populate, fill in real values, and add the chosen filename to
# .gitignore — these contain secrets.

# ---- context: dev ----
# CONTEXTS_DEV_BROKERS="b1.dev:9093 b2.dev:9093"
# CONTEXTS_DEV_SASL_USERNAME=
# CONTEXTS_DEV_SASL_PASSWORD=
# CONTEXTS_DEV_SCHEMAREGISTRY_URL=https://sr.dev:8081
# CONTEXTS_DEV_SCHEMAREGISTRY_USERNAME=
# CONTEXTS_DEV_SCHEMAREGISTRY_PASSWORD=

# ---- context: staging ----
# (same shape, CONTEXTS_STAGING_*)

# ---- context: prod ----
# (same shape, CONTEXTS_PROD_*)

# Any other kafkactl-shaped env var (TLS_CERTKEY, PGOPTIONS-style overrides,
# etc.) set here is also forwarded to kafkactl inside the container, as long
# as its name matches CONTEXTS_*, TLS_*, SASL_*, SCHEMAREGISTRY_*, or BROKERS.
```

Emit a `# ---- context: <name> ----` block for **every** context declared in the manifest.

### `<output>/scripts/kafkactl-config.yml`

A committed (no-secrets) kafkactl config that names each context and pins per-context defaults. Sample:

```yaml
contexts:
  dev:
    brokers: []   # populated from CONTEXTS_DEV_BROKERS at runtime
    sasl:
      enabled: true
      mechanism: <sasl_mechanism from manifest>
    tls:
      enabled: <true if cluster.tls=required else false>
    schemaRegistry:
      url: ""    # populated from CONTEXTS_DEV_SCHEMAREGISTRY_URL at runtime
  # ... one block per context
current-context: <first context name>
```

This file is committed into the repo and contains **no secrets**. The runtime env vars overlay broker addresses, SASL credentials, and Schema Registry credentials.

### `<output>/scripts/manifest.yml`

A verbatim copy of the manifest used to generate this skill, for transparency and as input for re-generation. The runtime scripts do not parse it — the allowlist is embedded in `_common.sh`'s arrays.

### `<output>/references/cluster.md`

Render `cluster.json` as a short markdown summary: cluster id, broker count, controller id, broker list. One short paragraph plus a table.

### `<output>/references/topics.md`

For each topic, a section like:

```markdown
## <topic>

- Partitions: <count>
- Replication factor: <rf>
- Cleanup policy: <policy>
- Retention: <ms or "compact-only">
- Schema (latest version): see `references/schemas/<topic>.json`

Notable config (only fields that differ from cluster default):

| key | value |
|---|---|
| ... | ... |
```

Pull the data from `topics/<topic>.json` written by `introspect.sh`.

### `<output>/references/groups.md`

For each consumer group, a section like:

```markdown
## <group>

- State: <Stable / Empty / Rebalancing / ...>
- Subscribed topics: <list>
- Member count: <n> (at generation time — this drifts; treat as a reference, not authoritative)
```

Pull the data from `groups/<group>.json`.

### `<output>/references/schemas/<topic>.json`

The latest Schema Registry response for `<topic>-value`, verbatim. Skip if Schema Registry is not configured or the subject is not registered.

## Step 4: Verify

After writing files, check:

- `<output>/SKILL.md` exists and is non-empty
- `<output>/scripts/{_common,describe-topic,describe-group,lag,consume,reset-offsets}.sh` all exist and are executable
- `<output>/scripts/.env.example` has one block per declared context
- `<output>/scripts/manifest.yml` is a verbatim copy of the input manifest (or the in-memory manifest built from manual flags)
- `<output>/references/cluster.md` and at least one `<output>/references/topics/*.md` exist
- No file under `<output>/` contains an unsubstituted `<...>` placeholder (grep for `<` followed by an identifier-shape — there will be matches in markdown table headers and SQL-style placeholders that are intentional, so check by regex `<[a-z_-]+>` and audit hits)

## Step 5: Smoke test

Pick the first topic in the manifest and run:

```bash
bash <output>/scripts/describe-topic.sh <first-topic> --context <first-context>
```

The same env vars used during introspection are still in scope, so the script picks them up without needing a `.env` file. If this fails, the generated skill is broken — surface the error to the user instead of claiming success.

## Step 6: Report

Tell the user:

- The output path the skill was written to
- The number of topics, groups, and contexts captured
- A reminder to copy `.env.example` to per-context `.env.<ctx>` files, fill in real values, and add them to `.gitignore`
- A reminder of the **non-destructive posture** — the skill cannot produce, alter, or delete; that's by design
- If `--output` was pointed at a plugin tree, that they may need to register the new skill in the plugin's `plugin.json` (this generator emits skill files only, not plugin manifests)
- That re-running the generator with the same manifest **overwrites** the output in place; treat any local edits to generated files as ephemeral
