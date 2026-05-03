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

`--output PATH` chooses where the generated skill is written. Defaults to `./.claude/skills/kafka-<team>/`. The primary use case for the override is a team building their own plugin: pointing `--output` at e.g. `plugins/team-payments/skills/kafka-payments/` lands the generated skill directly inside the plugin tree, no copy step. There is no short-form alias — `--output` is the only spelling — so the invocation shape stays singular for evals and operator muscle memory.

The path's parent directory must already exist (the operator chose where the skill goes; the generator does not create arbitrary parent paths). The leaf directory itself is the generator's responsibility:

- If it doesn't exist: create it.
- If it does exist: delete it and recreate it. Overwrite is intentional — manifests drift, and stale allowlists / references mislead.

Before deleting, validate the path against three layers of guards. Same shape `introspect.sh` uses for its `--output` argument; the rationale and patterns are documented at length there.

1. **Reject literal danger shapes.** Refuse if the leaf segment is empty, `/`, `.`, `..`, `~`, or `*`, or if `--output` itself is empty / leading-whitespace.
2. **Reject paths whose components include `..`** (e.g. `--output /tmp/a/../../etc`). `rm -rf` resolves the path before deletion, so a `..` segment slipping past the leaf check would still let the wipe land on a parent directory. Match `..` at any position: leading (`../foo`), trailing (`foo/..`), middle (`foo/../bar`).
3. **No leaf-prefix requirement here.** Unlike `introspect.sh`'s scratch dir (which always starts with `kafka-introspect-`), `--output` is a deliberate operator choice that may legitimately be `./.claude/skills/kafka-payments/`, `plugins/team-payments/skills/kafka-payments/`, or anywhere else. The two guards above plus path-safety on `team` (Precondition 4) are the safety net; the operator owns the rest.

### Optional environment overrides

These env vars adjust runtime behavior of both the generator (`introspect.sh` + the SR fetch step) and the generated wrappers. None are required; the defaults work for a developer with `docker` on their PATH and outbound access to Docker Hub.

| Variable | Purpose | Default |
|---|---|---|
| `KAFKA_CONTAINER_RUNTIME` | Force `docker` or `podman` rather than auto-detecting. | auto-detect |
| `KAFKACTL_IMAGE` | Container image for kafkactl invocations. Pin a different version, or point at a private registry mirror. | `docker.io/deviceinsight/kafkactl:v5.18.0-scratch` |
| `CURL_IMAGE` | Container image for the Schema Registry fetch step. Same override semantics as `KAFKACTL_IMAGE`. | `docker.io/curlimages/curl:8.11.1` |
| `KAFKA_DOCKER_ARGS` | Extra arguments appended to every `docker run` / `podman run` invocation. Common case on Linux when brokers are on `localhost`: `--network=host`. | empty |

Set these before invoking `/kafka-skill-creator` so introspection runs against the same image / runtime / network options the generated wrappers will use. The generated wrappers honor the same vars at runtime, so a private-registry pin set during generation should be re-exported in any session that uses the generated skill.

## Preconditions

Stop and refuse if any of the following are unmet:

1. **Container runtime** — `docker` or `podman` on `PATH` (auto-detected, or override via `KAFKA_CONTAINER_RUNTIME`).
2. **Manifest validates** against `scripts/manifest.schema.json`. v1 hard-fails on `cluster.auth` values other than `SASL_SCRAM` and points at the matching deferred issue. The schema annotates `default:` values for two optional fields, but JSON Schema's `default` is documentation, not validator behavior — fill them in explicitly after validation so two runs against the same manifest are deterministic regardless of whether the operator wrote the defaults out:
   - `cluster.tls` defaults to `required` when omitted.
   - `cluster.schema_registry.auth` defaults to `basic` when the `cluster.schema_registry` block is present and `auth` is omitted.
3. **Context names are unique after normalization.** Two complications stack here. First, JSON Schema's `uniqueItems: true` only rejects fully-identical objects, so two contexts with the same `name` but different `sasl_mechanism` would slip past it. Second, env-var lookup normalizes context names by uppercasing and replacing `-` with `_` — so `dev-1` and `dev_1` are *distinct* raw strings but collide on the same `CONTEXTS_DEV_1_*` env-var prefix. After the schema validates, walk `contexts[]`, normalize each `name` (uppercase + `-` → `_`), and refuse with a named error if any normalized form repeats. Surface both the raw names and the colliding normalized form in the error so the operator can fix the manifest without digging through env-var conventions.
4. **Path-safety on `team`** — must match `^[A-Za-z0-9_-]+$`. The `--output` path itself is taken at face value (it's an explicit operator choice), but `team` flows into generated frontmatter and the default output path segment, so it stays restricted.
5. **Per-context env vars are populated** for every context declared in the manifest. For each `<ctx>`, derive `<UPPER>` by uppercasing the name and replacing `-` with `_`, then check that all three are set and non-empty:
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

6. **No connection-string positional argument**. `/kafka-skill-creator "kafka://user:pass@broker:9092"` is rejected; do not parse the password out of the URL even silently. The skill takes flags only.

## High-level workflow

1. Parse and validate inputs (manifest or flags); refuse on conflict or missing values per Preconditions.
2. Resolve the output directory; validate path safety on the leaf.
3. Validate per-context env vars (Precondition 5).
4. Run introspection against the **first** context declared in the manifest. The manifest is shared across environments; one context's introspection is enough for the references because schemas should match across environments. (If they don't, the team has a bigger problem than this skill.)
5. Generate the skill files at the resolved output path.
6. Verify the generated skill (file existence, no leftover `<...>` placeholders, scripts marked executable).
7. Smoke test: invoke the generated `describe-topic.sh` against the first manifest topic using the same context's env vars. Surface failure; do not claim success.
8. Report what was written and how to use it.

## Step 1: Run introspection

Before invoking `introspect.sh`, export the manifest's static cluster-shape values for the chosen context. The script treats `CONTEXTS_<NAME>_SASL_MECHANISM` as optional (kafkactl has a default) but kafkactl's default is not guaranteed to match this team's cluster — a SCRAM-SHA-256 broker would silently reject SCRAM-SHA-512 auth and the whole introspection would fail with an opaque error. Setting the manifest values explicitly ties the captured references to the same auth shape the generated wrappers will use:

```bash
upper="$(printf '%s' "$CONTEXT" | tr '[:lower:]' '[:upper:]')"
upper="${upper//-/_}"

# Static shape from the manifest. None of these are secrets.
export CONTEXTS_${upper}_SASL_ENABLED=true
export CONTEXTS_${upper}_SASL_MECHANISM="<sasl_mechanism for this context from manifest>"
export CONTEXTS_${upper}_TLS_ENABLED="<true if cluster.tls=required else false>"
# (CONTEXTS_${upper}_SCHEMAREGISTRY_AUTH only set in Step 2 when SR is configured)

bash <skill-dir>/scripts/introspect.sh \
  --context "$CONTEXT" \
  --topic <T> [--topic <T> ...] \
  --group <G> [--group <G> ...] \
  /tmp/kafka-introspect-<team>
```

`<skill-dir>` is wherever this skill is installed (the directory containing the `SKILL.md` you are reading). Use an absolute path.

`introspect.sh` wipes its output directory on every run before writing, so a re-introspection after a manifest change (topic dropped, group renamed) doesn't leave stale JSON files lying around to confuse downstream rendering. **The leaf segment of the output path must start with `kafka-introspect-`** (e.g. `/tmp/kafka-introspect-<team>`). That prefix is the safety pin that lets the script recursively delete its output without risk of nuking a high-impact directory — passing `/tmp` or `/home/user/work` would be catastrophic to wipe, so the script refuses any leaf without the required prefix. The script also refuses `/`, `.`, `..`, `~`, paths containing `..` segments, and empty / leading-whitespace paths.

`introspect.sh` also re-validates the env vars itself, so a missing variable surfaces with the same exit-2 refusal even if Preconditions somehow passed. Output layout:

- `cluster.json` — broker list and cluster metadata (kafkactl `get brokers -o json`).
- `topics/<topic>.json` — per-topic config and partition layout (kafkactl `describe topic -o json`).
- `groups/<group>.json` — per-group members, subscriptions, lag (kafkactl `describe consumer-group -o json`).

If any individual call fails, `introspect.sh` writes an empty file for that target and continues. Partial coverage is better than aborting the whole generation.

## Step 2: Pull schemas (Schema Registry, when configured)

If the manifest has a `cluster.schema_registry` block, fetch the latest schema for each topic's `<topic>-value` subject. Read `references/schema-registry-fetch.md` for the verbatim bash; the snippet relies on bash indirect expansion (`${!varname}`) plus docker's `-e VARNAME` form (no `=value` — docker reads the value from the host environment so it never lands in `docker run`'s argv) plus `curl -K -` (config-via-stdin so the password stays out of curl's argv inside the container).

Inputs the snippet expects: `TEAM`, `CONTEXT`, `TOPICS=()` (bash array), and `RUNTIME` (the snippet auto-detects this with the same block `introspect.sh` uses).

Persist each response **verbatim** as `<output>/references/schemas/<topic>.json` in Step 3 — the Schema Registry envelope already carries `id`, `version`, `schemaType`, and the schema string, so the model can read whichever piece it needs without a per-format dispatch step. Don't extract the inner `schema` string into format-specific files (`.avsc` / `.proto`); that's a future enhancement, and doing it inconsistently is the bigger trap.

If Schema Registry returns 404 for a subject, skip it and note the absence in `references/topics.md` for that topic — partial coverage is better than aborting the whole generation.

## Step 3: Write the generated skill

Create these files under `<output>/` (the resolved output directory). Substitute the `<...>` placeholders with real content from the manifest and the introspection dumps.

### `<output>/SKILL.md`

Read `references/generated-skill-md-skeleton.md` for the verbatim template. The substitution rules:

- `<team>` — the manifest's `team` field, verbatim.
- `<top topics>` — the **first up to 5** entries from the manifest's `topics:` list, in manifest order, joined by `", "`. Operators control which topics surface in the triggering description by ordering them in the manifest; this is the deterministic rule. Don't apply subjective rules like "most prominent" or "most-used".
- `<env list>` and `<list of context names>` — the manifest's `contexts[].name` values, in manifest order, joined by `" / "` (e.g. `dev / staging / prod`).
- `<bullet list>` for owned topics and groups — one `- ` line per item.

The description is generated from a fixed template — substitution only, no paraphrasing. Two runs against the same manifest must produce byte-identical descriptions. Don't paraphrase the trigger copy for "tone", don't add team-specific color, don't drop the "even if they don't say 'Kafka' explicitly" clause to make it shorter — the wording is what triggers reliably.

**Do NOT set `disable-model-invocation: true` on the generated skill.** The generated `kafka-<team>` skill is meant to fire automatically when its description matches the user's prose ("what's lag on the orders projector?", "peek at the last few payments.refunds.v1 messages"). Disabling model invocation would force the user to type `/kafka-<team>` explicitly to use it, which defeats the point of having a per-team skill that the model recognizes by topic context. The meta-generator (this `kafka-skill-creator`) is slash-only because *it* requires deliberate invocation; the *generated* skill should not inherit that property. Omit the field — Claude Code's default is "model-invocable" — rather than setting it to `false` (the field's name is a footgun; explicit `false` reads like an extra knob even though it's just the default).

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

# Any other kafkactl-shaped env var (e.g. CONTEXTS_<NAME>_TLS_CERTKEY for
# mTLS-style cert pinning, CONTEXTS_<NAME>_SASL_MECHANISM to override the
# default, or a bare BROKERS shorthand for default-context use) set here is
# also forwarded to kafkactl inside the container, as long as its name
# matches CONTEXTS_*, TLS_*, SASL_*, SCHEMAREGISTRY_*, or BROKERS.
```

Emit a `# ---- context: <name> ----` block for **every** context declared in the manifest.

### Per-context static values exported by `_common.sh`

The manifest declares values that are the same across every environment (`sasl_mechanism`, `cluster.tls`, `cluster.schema_registry.auth`). These would normally live in a `kafkactl-config.yml` config file, but mounting that file into the kafkactl container would require pinning a mount path the container expects, and inconsistencies between the file's contents and the env vars are easy to introduce. kafkactl's documented env-var convention (`CONTEXTS_<NAME>_<FIELD>` autocreates the context with that field set), so the wrappers can route everything through env vars and skip the config file entirely.

`_common.sh` therefore exports per-context static values at generation time, one block per declared context:

```bash
# Per-context static values from the manifest. Secrets come from .env at
# runtime; these are the cluster-shape fields that don't change between
# environments. They flow to the container via the same forwarding filter
# as the .env-supplied secrets (see build_env_args).

# context: dev
export CONTEXTS_DEV_SASL_ENABLED=true
export CONTEXTS_DEV_SASL_MECHANISM=SCRAM-SHA-512
export CONTEXTS_DEV_TLS_ENABLED=true
export CONTEXTS_DEV_SCHEMAREGISTRY_AUTH=basic   # only when manifest declares schema_registry

# context: staging  (same shape)
# context: prod     (same shape)
```

Substitute `SCRAM-SHA-512` with the manifest's `contexts[].sasl_mechanism`, `true`/`false` for `TLS_ENABLED` based on `cluster.tls`, and emit the `SCHEMAREGISTRY_AUTH` line only when the manifest has a `schema_registry` block.

### `<output>/scripts/manifest.yml`

A verbatim copy of the manifest used to generate this skill, for transparency and as input for re-generation. The runtime scripts do not parse it — the allowlist is embedded in `_common.sh`'s arrays.

### `<output>/README.md`

A human-facing README the team can read in their plugin tree (or under `.claude/skills/kafka-<team>/`) without having to open `SKILL.md`. SKILL.md is written for Claude when triggering; the README is for the engineer who's about to use the skill or onboard a new teammate.

Read `references/generated-readme-skeleton.md` for the verbatim template. Substitute the `<...>` placeholders with real values from the manifest at generation time:

- `<team>` and `<env list>` — straight from the manifest
- `<bullet list>` — emit one `- ` line per item under `topics:` and `consumer_groups:`
- `<first-topic>` and `<first-group>` — the first entries of those lists, used in the sample-use code blocks so the examples are runnable copy-paste

The README must include a working sample for every wrapper (`consume.sh`, `describe-topic.sh`, `describe-group.sh`, `lag.sh`, `reset-offsets.sh`), the one-time-setup `.env` step, and the regeneration command — those are the things engineers ask first when they open the skill directory.

### `<output>/references/cluster.md`

Render `cluster.json` using the fixed template below — substitution only, no paraphrasing, no judgment calls about which fields to include. Two runs against the same JSON must produce byte-identical output.

```markdown
# Cluster

- Broker count: <N>
- Controller broker id: <id-or-"(unknown)" if the JSON has no controller field>

## Brokers

| id | address |
|---|---|
| <id> | <address> |
| ... | ... |
```

Sort the broker rows by `id` ascending (numeric sort, not lexicographic — broker `10` comes after `9`, not after `1`). If a broker row is missing the `address` field in the JSON, write `(unknown)` — never omit the row, never substitute a guess.

### `<output>/references/topics.md`

For each topic, emit a section using the fixed template below. Iterate topics in the manifest's `topics:` order — same deterministic ordering rule used by the SKILL.md description substitution.

```markdown
## <topic>

- Partitions: <count>
- Replication factor: <rf>
- Cleanup policy: <policy>
- Retention: <ms or "compact-only">
- Schema (latest version): see `references/schemas/<topic>.json`

Notable config (topic-level overrides):

| key | value |
|---|---|
| <key> | <value> |
| ... | ... |
```

The "Notable config" table includes exactly the entries from the JSON's `configs[]` array whose `source` field equals `"DYNAMIC_TOPIC_CONFIG"` — that's the kafkactl / Kafka-admin label for topic-level overrides (as opposed to `DEFAULT_CONFIG` for inherited cluster defaults, `STATIC_BROKER_CONFIG`, etc.). Sort the rows lexicographically by `key` (case-sensitive).

Fallback: if `configs[]` entries don't have a `source` field at all (older kafkactl versions, or a future schema change), include every entry under the same lexicographic sort. Incomplete coverage is better than dropping the section.

Pull the data from `topics/<topic>.json` written by `introspect.sh`.

### `<output>/references/groups.md`

For each consumer group, emit a section using the fixed template below. Iterate groups in the manifest's `consumer_groups:` order so re-runs against the same manifest are byte-identical.

```markdown
## <group>

- State: <Stable / Empty / Rebalancing / ...>
- Subscribed topics: <comma-separated list, sorted lexicographically; or "(none)" if empty>
- Member count: <n> (at generation time — this drifts; treat as a reference, not authoritative)
```

Pull the data from `groups/<group>.json`.

### `<output>/references/schemas/<topic>.json`

The latest Schema Registry response for `<topic>-value`, verbatim. Skip if Schema Registry is not configured or the subject is not registered.

## Step 4: Verify

After writing files, check:

- `<output>/SKILL.md` exists and is non-empty
- `<output>/SKILL.md`'s frontmatter does **not** contain `disable-model-invocation: true` (the generated skill must be model-invocable so it fires on natural-language prompts)
- `<output>/README.md` exists and is non-empty
- `<output>/scripts/{_common,describe-topic,describe-group,lag,consume,reset-offsets}.sh` all exist and are executable
- `<output>/scripts/.env.example` has one block per declared context
- `<output>/scripts/manifest.yml` is a verbatim copy of the input manifest (or the in-memory manifest built from manual flags)
- `<output>/references/cluster.md`, `<output>/references/topics.md`, and `<output>/references/groups.md` all exist and are non-empty
- No file under `<output>/` contains an unsubstituted template token. Refuse if `grep -rE '<(team|env list|list of context names|top topics|bullet list|first-topic|first-group|this-directory|sasl_mechanism from manifest|UPPER(-[0-9]+)?|topic-[0-9]+|group-[0-9]+|ctx-[0-9]+)>' <output>/` returns any match — every entry in that pattern is a substitution placeholder from the skeleton templates that the generator must replace with real values. The regex is enumerated rather than open-ended (`<[a-z_-]+>`) on purpose: `<topic>`, `<group>`, `<ctx>`, `<flag-name>`, `<T>`, `<NAME>`, `<CTX>`, `<FIELD>` legitimately appear in generated USAGE strings and kafkactl env-var docs, so a blanket "any angle-bracket identifier" check would false-fire on the documentation surface of the wrappers.

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
