# Transcript — eval-1-accepts-mtls-manifest / old_skill

## Skill version under test

The pre-MTLS snapshot at:

`plugins/kafka-skill-creator/skills/kafka-skill-creator-workspace/iteration-2/skill-snapshot/SKILL.md`

This is the version of `kafka-skill-creator` from git HEAD before the issue-#64 mTLS changes. Its v1 scope explicitly defers mTLS to issue #64.

## User invocation (simulated)

```
/kafka-skill-creator --manifest /tmp/team-eval1-old.yml --output /tmp/skill-out-eval1-old
```

Manifest contents:

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

Env vars staged in the eval shell (`CONTEXTS_DEV_BROKERS`, `CONTEXTS_DEV_TLS_CERT`, `CONTEXTS_DEV_TLS_CERTKEY`, `CONTEXTS_DEV_TLS_CA`) — irrelevant in practice because the skill refuses before Precondition 5 (per-context env vars) is reached. Direct `/tmp` writes from this sandbox were denied, but that does not affect the refusal path; the manifest text above is exactly what the user supplied.

## Walkthrough of the OLD skill against the manifest

Working through the snapshot's Preconditions in order:

1. **Container runtime** — not yet checked; refusal happens before runtime detection.
2. **Manifest validates against `scripts/manifest.schema.json`** — FAILS. The v1-scope section (SKILL.md lines 23-28) is unambiguous:

   > **Auth**: SASL/SCRAM only ...
   > **Deferred**: ... mTLS ([#64]) ...
   > When a manifest specifies an auth value other than `SASL_SCRAM`, refuse with a one-line pointer to the matching deferred-auth issue. Do not silently accept and degrade.

   And Precondition 2 (line 95) reinforces: "v1 hard-fails on `cluster.auth` values other than `SASL_SCRAM` and points at the matching deferred issue."

   The manifest declares `cluster.auth: MTLS`, which is the value explicitly mapped to deferred issue [#64](https://github.com/z5labs/ai/issues/64).

The skill MUST refuse here. It does NOT silently degrade to SASL_SCRAM. It does NOT generate any files. Preconditions 3-6 are not evaluated.

## Refusal message produced

Saved verbatim to `outputs/refusal.txt`. The message:

- Names `MTLS` as the unsupported value.
- Cites issue #64 (with full GitHub URL).
- Reminds the user that v1 only accepts `cluster.auth: SASL_SCRAM`.
- Explicitly states no files were written under `/tmp/skill-out-eval1-old/`.
- Offers two forward paths (change the manifest, or wait for #64 to land).

## Files generated under `/tmp/skill-out-eval1-old/`

None. The OLD skill correctly halted at Precondition 2 before the "Step 3: Write the generated skill" phase. `outputs/` therefore contains only `refusal.txt`.
