# kafkactl env-var convention

This skill — and every skill it generates — follows kafkactl's documented rule for translating its YAML config into environment variables. The rule lets the same per-context config field be supplied via env var without changing the consumer.

## The rule

For any kafkactl config key under `contexts.<context>.<field>...`, the equivalent env var is:

1. Replace `.` with `_`.
2. Replace `-` with `_`.
3. Convert to ALL CAPS.
4. Prefix with `CONTEXTS_<UPPER-CONTEXT-NAME>_`.

So the YAML field `contexts.dev.sasl.password` becomes `CONTEXTS_DEV_SASL_PASSWORD`.

The default context (named `default`) supports a shorthand: drop the `CONTEXTS_DEFAULT_` prefix. `SASL_PASSWORD` and `CONTEXTS_DEFAULT_SASL_PASSWORD` mean the same thing. This skill's generated wrappers always pass `--context <ctx>` explicitly, so the shorthand isn't required, but the env-var filter forwards both shapes for parity with how operators may have set up their host environment.

## kafkactl gotchas this skill works around

Two things about kafkactl don't behave the way the env-var rule alone suggests, and both are load-bearing — the skill stops working without compensating for them. They were caught by the e2e fixture (#66); record them here so the next time someone goes to "simplify" the workarounds they understand why those workarounds exist.

### 1. Env-var contexts must already exist in `config.yml`

`CONTEXTS_<NAME>_*` env vars **overlay** existing contexts; they do **not** auto-create them. `kafkactl --context dev get topics` with `CONTEXTS_DEV_BROKERS=...` and no `dev` context in `config.yml` errors with `not a valid context: dev`. The env vars are silently ignored and the bare `BROKERS` shorthand routes everything into `default` instead.

The skill compensates by generating a one-line `config.yml` with an empty-bodied entry per context (`contexts.<ctx>: {}`) into a temp dir and mounting it at `/.config/kafkactl/config.yml`. The env vars then populate the real broker, credentials, and SR settings on top. See `prepare_kafkactl_config` in the generated `_common.sh` and the equivalent block in `scripts/introspect.sh`.

### 2. SASL mechanism casing

The manifest schema declares SCRAM mechanisms in Kafka's canonical form, `SCRAM-SHA-256` and `SCRAM-SHA-512`. kafkactl rejects both with `Unknown sasl mechanism`. It accepts only the squashed-lowercase form: `scram-sha256`, `scram-sha512`.

The skill compensates in two places:

- `scripts/introspect.sh` re-exports `CONTEXTS_<UPPER>_SASL_MECHANISM` after translating the value at runtime — the manifest format reaches the script via the env var, the kafkactl format reaches the container.
- The generated `_common.sh` emits the per-context `SASL_MECHANISM` export in kafkactl's casing directly. The translation happens at generation time, baked into the file the team checks in.

Both paths are needed: introspection runs against operator-supplied env vars, the wrappers run from a generated file with values already known.

## Required keys for v1 (per context)

The required set depends on the manifest's `cluster.auth`. `BROKERS` is always required.

**`auth: SASL_SCRAM`:**

```
CONTEXTS_<CTX>_BROKERS               # whitespace-separated host:port list
CONTEXTS_<CTX>_SASL_USERNAME
CONTEXTS_<CTX>_SASL_PASSWORD
```

**`auth: MTLS`:**

```
CONTEXTS_<CTX>_BROKERS               # whitespace-separated host:port list
CONTEXTS_<CTX>_TLS_CERT              # absolute path to client cert PEM
CONTEXTS_<CTX>_TLS_CERTKEY           # absolute path to client key PEM
CONTEXTS_<CTX>_TLS_CA                # absolute path to CA bundle PEM
```

For MTLS, the wrapper bind-mounts each cert path **read-only** into the kafkactl container at the same path the env var declares (e.g. `/etc/ssl/kafka/prod.crt:/etc/ssl/kafka/prod.crt:ro,z`). That keeps kafkactl's view of the path identical inside and outside the container — no in-container path translation, no in-container cert staging. The `:z` is the SELinux shared-relabel marker — ignored on hosts without SELinux, but required on Fedora/RHEL so the container process can actually read the bind-mounted cert files (without it the host's file context is inaccessible from the container's process context and reads fail with `Permission denied`). Cert paths must be **absolute** (docker bind-mount syntax requires it) and must **exist on the host** at validation time. Only the active `--context`'s cert paths get mounted; paths exported for other contexts are also dropped by the per-context-scoped env-forwarding filter (see "What the forwarding filter passes" below), so kafkactl in the container neither sees them as env vars nor has a mount to read them through — a prod cert path can't leak into a dev container at either layer.

Optional, when the manifest declares Schema Registry (independent of broker auth):

```
CONTEXTS_<CTX>_SCHEMAREGISTRY_URL
CONTEXTS_<CTX>_SCHEMAREGISTRY_USERNAME    # if cluster.schema_registry.auth: basic
CONTEXTS_<CTX>_SCHEMAREGISTRY_PASSWORD    # if cluster.schema_registry.auth: basic
```

The `<CTX>` segment is the context's `name` field uppercased and with `-` replaced by `_`. So `--context dev-us-east` reads `CONTEXTS_DEV_US_EAST_*`.

## Array values

Whitespace-separated. `BROKERS="b1.dev:9093 b2.dev:9093"` becomes the broker list `[b1.dev:9093, b2.dev:9093]` after kafkactl parses it.

## What the forwarding filter passes

`introspect.sh` and the generated wrappers forward env vars whose names match `^(CONTEXTS_<ACTIVE_UPPER>_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS$)`, where `<ACTIVE_UPPER>` is the `--context` value uppercased and hyphen-to-underscored. This covers:

- The **active context's** per-context overrides only (`CONTEXTS_<ACTIVE_UPPER>_*`).
- The default-context shorthand `BROKERS`, `SASL_*`, `TLS_*`, `SCHEMAREGISTRY_*`.

`CONTEXTS_<OTHER>_*` vars (other contexts' per-context overrides) are intentionally **not** forwarded. kafkactl with `--context <active>` only consults the active context's vars anyway, so other-context vars would be functionally useless inside the container — and forwarding e.g. `CONTEXTS_PROD_TLS_CERT` to a `--context dev` container would leak the prod cert path string into the container even though the file is not bind-mounted. Scoping the filter to the active context closes that path-leak.

It also does **not** forward names starting with `KAFKA_` — those are reserved for this plugin's internal config (`KAFKA_ENV_FILE`, `KAFKA_DOCKER_ARGS`, `KAFKA_CONTAINER_RUNTIME`) and are not kafkactl-shaped — nor any name starting with `CONTEXT_AUTH_MODE_`, which is the wrappers' internal auth-mode selector (see `references/generated-skill-scripts.md`).

## Why we use env vars instead of a connection string

Same reason `postgres-skill-creator` settled on libpq env vars in [#42](https://github.com/z5labs/ai/issues/42): a credential that reaches the tool through argv or a connection string passes through whatever process invoked the tool — including the model's transcript. Env vars sidestep that path; they reach the container directly via `docker run -e VARNAME` (forwarding the value, never the literal in argv) and never appear in process listings.

## Pairing with credential helpers

The per-environment values (broker addresses, SASL passwords, Schema Registry credentials) live in `.env.<context>` files at runtime. To avoid storing secrets on disk in plain text, run the consumer (Claude Code, your shell, the generator) under a credential helper that resolves them on demand:

```
op run --env-file=kafka-prod.env -- claude
```

…where `kafka-prod.env` declares `CONTEXTS_PROD_SASL_PASSWORD=op://Vault/Kafka/prod-password`. `op run` substitutes the live secret into the subprocess environment and tears it down on exit. The same shape works with `vault read … | direnv …`, `gcloud secrets versions access`, and any other helper that exposes secrets via env vars.
