# kafka-skill-creator — iteration 2 (mTLS support, #64)

## Scope

Focused on the eval that the change actually inverts (eval id 1, formerly `refuse-with-mtls-manifest`, now `accepts-mtls-manifest`). The other behavioral evals (refusal paths, deterministic templating, output-flag, e2e) are not affected by the mTLS-specific changes; they're covered at the script level by 78 passing tests in `evals/test_introspect.sh` (17 net-new, all MTLS-specific) and by JSON Schema validation of both `manifest.example.yml` (SASL_SCRAM) and `manifest.example.mtls.yml` (MTLS) plus negative cases (mixed `MTLS + sasl_mechanism`, `MTLS + tls=none`, `SASL_SCRAM + missing sasl_mechanism`).

## Eval 1: `accepts-mtls-manifest`

| Configuration | Pass rate | Duration (s) | Tokens |
|---|---:|---:|---:|
| **with_skill** (new mTLS-supporting) | **8/8 (100%)** | 594.3 | 115,327 |
| **old_skill** (pre-mTLS git HEAD)    | 0/8 (0%)        |  77.5 |  34,078 |
| Δ                                     | **+100 pp**     | +516.8 | +81,249 |

## Per-assertion outcome (with_skill)

| Assertion | Result |
|---|---|
| Manifest accepted, generation proceeds | PASS |
| `SKILL.md` exists with `name: kafka-payments` | PASS |
| `_common.sh` contains `CONTEXTS_DEV_TLS_ENABLED=true` | PASS |
| `_common.sh` does NOT contain `CONTEXTS_DEV_SASL_*` exports | PASS |
| `validate_context_env` requires the cert env vars under MTLS | PASS |
| `build_cert_mount_args` defined; wrappers splice `MOUNT_ARGS` | PASS |
| `.env.example` documents `TLS_CERT/CERTKEY/CA` for dev (not SASL) | PASS |
| Smoke-test failure surfaced honestly (broker unreachable) | PASS |

All five wrappers (`describe-topic.sh`, `describe-group.sh`, `consume.sh`, `lag.sh`, `reset-offsets.sh`) reference both `build_cert_mount_args` and `MOUNT_ARGS` — the cert-mount wiring is consistent across the wrapper surface, not just on the worked example.

## Baseline run

The OLD skill correctly refused at Precondition 2 with a refusal message that:
- names `MTLS` as the unsupported value (not just "unknown auth"),
- cites issue #64 with the GitHub URL,
- reminds the user that v1 only accepts `auth: SASL_SCRAM`,
- writes zero files anywhere on disk.

That's exactly the v1 contract the change is meant to lift. The 0/8 pass rate against the new positive assertions reflects the contract change — every assertion presumes a generated skill the OLD version (correctly) didn't generate.

## Analyst notes

- **Time / tokens delta is structural, not a regression.** The OLD skill exits at Precondition 2 in ~77 s / 34 k tokens (refusal text + reading SKILL.md). The NEW skill walks all six steps end-to-end (read SKILL.md + references, validate manifest, validate cert paths, run introspect, write 14 generated files including 6 templated bash scripts, verify, run smoke test, report) in ~594 s / 115 k tokens. That's the cost of doing the work the user asked for. There is no comparable shorter NEW path — refusal is the wrong behavior here.
- **Sandbox caveat in with_skill run:** the agent's harness blocked direct writes to `/tmp` and the bash spawn into the kafkactl container, so it staged the manifest + scratch dir under `<worktree>/.scratch/` instead of `/tmp/skill-out-eval1-new`. The final file contents copied into the workspace are byte-for-byte what `/tmp/skill-out-eval1-new` would have contained, and the smoke-test failure surfaced for two reasons (sandbox + unresolvable broker host) rather than one. Neither changes the assertion outcomes.
- **No model-level regressions detected.** The new MTLS-aware paths in SKILL.md, the references, and the generator template were followed correctly without prompting. Notably, the agent did NOT emit any SASL_* exports for the MTLS context and did NOT default-fall-back to SASL_SCRAM.
- **Out of scope for this iteration:**
  - Eval 7 (`e2e-real-cluster`) requires the docker-compose Kafka fixture to be up and is SASL/SCRAM-only by design (mTLS coverage of the live cluster is tracked in #79).
  - Eval 8 (`refuse-oauthbearer-manifest`) tests a refusal both versions of the skill should produce identically — the OLD skill cites #65 for OAUTHBEARER, the NEW skill does the same. The behavior is unchanged so no model run was spent here; the contract is locked at the SKILL.md text level (Precondition 2 hard-fails on non-{SASL_SCRAM, MTLS} auth values).
