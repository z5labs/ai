# kafkactl env-var convention

This skill — and every skill it generates — follows kafkactl's documented rule for translating its YAML config into environment variables. The rule lets the same per-context config field be supplied in three different ways without changing the consumer.

## The rule

For any kafkactl config key under `contexts.<context>.<field>...`, the equivalent env var is:

1. Replace `.` with `_`.
2. Replace `-` with `_`.
3. Convert to ALL CAPS.
4. Prefix with `CONTEXTS_<UPPER-CONTEXT-NAME>_`.

So the YAML field `contexts.dev.sasl.password` becomes `CONTEXTS_DEV_SASL_PASSWORD`.

The default context (named `default`) supports a shorthand: drop the `CONTEXTS_DEFAULT_` prefix. `SASL_PASSWORD` and `CONTEXTS_DEFAULT_SASL_PASSWORD` mean the same thing. This skill's generated wrappers always pass `--context <ctx>` explicitly, so the shorthand isn't required, but the env-var filter forwards both shapes for parity with how operators may have set up their host environment.

## Required keys for v1 (per context)

```
CONTEXTS_<CTX>_BROKERS               # whitespace-separated host:port list
CONTEXTS_<CTX>_SASL_USERNAME
CONTEXTS_<CTX>_SASL_PASSWORD
```

Optional, when the manifest declares Schema Registry:

```
CONTEXTS_<CTX>_SCHEMAREGISTRY_URL
CONTEXTS_<CTX>_SCHEMAREGISTRY_USERNAME    # if cluster.schema_registry.auth: basic
CONTEXTS_<CTX>_SCHEMAREGISTRY_PASSWORD    # if cluster.schema_registry.auth: basic
```

The `<CTX>` segment is the context's `name` field uppercased and with `-` replaced by `_`. So `--context dev-us-east` reads `CONTEXTS_DEV_US_EAST_*`.

## Array values

Whitespace-separated. `BROKERS="b1.dev:9093 b2.dev:9093"` becomes the broker list `[b1.dev:9093, b2.dev:9093]` after kafkactl parses it.

## What the forwarding filter passes

`introspect.sh` and the generated wrappers forward env vars whose names match `^(CONTEXTS_|TLS_|SASL_|SCHEMAREGISTRY_|BROKERS$)`. This covers:

- All `CONTEXTS_*` per-context overrides.
- The default-context shorthand `BROKERS`, `SASL_*`, `TLS_*`, `SCHEMAREGISTRY_*`.

It does **not** forward names starting with `KAFKA_` — those are reserved for this plugin's internal config (`KAFKA_ENV_FILE`, `KAFKA_DOCKER_ARGS`, `KAFKA_CONTAINER_RUNTIME`) and are not kafkactl-shaped.

## Why we use env vars instead of a connection string

Same reason `postgres-skill-creator` settled on libpq env vars in [#42](https://github.com/z5labs/ai/issues/42): a credential that reaches the tool through argv or a connection string passes through whatever process invoked the tool — including the model's transcript. Env vars sidestep that path; they reach the container directly via `docker run -e VARNAME` (forwarding the value, never the literal in argv) and never appear in process listings.

## Pairing with credential helpers

The per-environment values (broker addresses, SASL passwords, Schema Registry credentials) live in `.env.<context>` files at runtime. To avoid storing secrets on disk in plain text, run the consumer (Claude Code, your shell, the generator) under a credential helper that resolves them on demand:

```
op run --env-file=kafka-prod.env -- claude
```

…where `kafka-prod.env` declares `CONTEXTS_PROD_SASL_PASSWORD=op://Vault/Kafka/prod-password`. `op run` substitutes the live secret into the subprocess environment and tears it down on exit. The same shape works with `vault read … | direnv …`, `gcloud secrets versions access`, and any other helper that exposes secrets via env vars.
