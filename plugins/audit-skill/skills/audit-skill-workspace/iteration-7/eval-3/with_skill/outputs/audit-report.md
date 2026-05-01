# Audit: mongo-explorer

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-3/with_skill/work/skills/mongo-explorer/`
- Date: 2026-04-30
- Findings: 19  (idempotency: 3, reproducibility: 3, context-management: 0, strict-definitions: 4, security: 9)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses so the model doesn't have to infer it from the workflow.
- `SKILL.md:21` — output path `./.claude/skills/mongo-<dbname>/.env` is written without specifying overwrite/append/refuse behavior when a prior `.env` already exists; document the intent.
- `SKILL.md:60` — generated `scripts/query.sh` is written without specifying behavior when a prior version exists; the workflow says the parent directory is overwritten but does not call out the script file specifically.

### Reproducibility

- `SKILL.md:32` — workflow depends on the `mongo` CLI being installed and on the live database state at run time, but neither is listed under Inputs / Preconditions; two runs against the same connection string on different days can diverge silently if the schema changed.
- `SKILL.md:33` — depends on the `$CONNECTION_STRING` shell variable being set in the runtime environment without listing it as a declared input or saying who exports it.
- `SKILL.md:80` — example invocation `mongodb://app_user:hunter2@db.example.com:27017/myapp` does not match the rule on line 14 (password should be supplied via `MONGO_PASSWORD`); the example contradicts the URL-without-password path the rule describes.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description has no "when to skip" / negative case (likely to over-trigger on adjacent tasks like generating a Postgres schema skill or wiring a Mongo client into application code).
- `SKILL.md:3` — description claims "Introspect a MongoDB database from a connection string and generate a project-level skill" but the workflow at lines 14, 21, 47-56, 62-69 also writes credentials to disk and hardcodes fallbacks; the side-effect class (persists credentials to disk) is not surfaced in the description so trigger decisions can't account for it.
- `SKILL.md:12` — input "connection string" is referenced but its required-ness when missing is split between two contradictory paths (line 12 says prompt for the full string including password; line 14 says prompt for `MONGO_PASSWORD` if the URL omits it); pick one source-of-truth and state the validation rule.
- `SKILL.md:24` — step 1 says "discard the password from the in-memory representation of the URL" but step 2 at line 33 reads `$CONNECTION_STRING` (which still contains the password) into a child process; the discard precondition for step 2 is contradicted by the step itself.

### Security

- `SKILL.md:5` — argument-hint `[connection-string]` accepts a `mongodb://user:password@host` URL that admits an embedded password; the model sees the credential as an argument. Split routing from the secret: accept host/port/user/dbname via env vars or non-credentialed config and read the password from a separate env var the model never references.
- `SKILL.md:12` — instructs the model to prompt the user for "the full connection string including the password"; the password enters the conversation context the moment the user answers. Refuse-and-instruct (tell the user to export `MONGO_PASSWORD` out-of-band) instead of prompting.
- `SKILL.md:12` — discard-after-read pattern: "After parsing the connection string, discard the password from the URL since it is no longer needed." The secret was already in the model's context when the discard ran. Fix by not accepting the secret as an argument or prompt in the first place.
- `SKILL.md:14` — instructs the model to prompt the user for their MongoDB password and export it as `MONGO_PASSWORD`; the secret enters the conversation context the moment the user answers. Refuse-and-instruct instead.
- `SKILL.md:26` — second discard-after-read instruction ("discard the password from the in-memory representation of the URL — keep it only in the `MONGO_PASSWORD` variable"); same issue as line 12 — the password has already been seen.
- `SKILL.md:28-29` — third prompt-for-password instruction ("ask the user to paste their password. Do not proceed until you have it."); the model must not prompt for credentials.
- `SKILL.md:47-56` — generated `./.claude/skills/mongo-<dbname>/.env` contains a concrete value for `MONGO_PASSWORD` (the line `MONGO_PASSWORD=<password>` is filled in with "the real values from the connection string at write time"); the credential persists on disk after the run. Emit a `.env.example` with placeholder values and have the user populate the real value out-of-band.
- `SKILL.md:62-69` — generated `scripts/query.sh` hardcodes `MONGO_USER="${MONGO_USER:-app_user}"` and `MONGO_PASSWORD="${MONGO_PASSWORD:-hunter2}"` as fallbacks; the credential is baked into a generated script and persists on disk. Drop the fallback values entirely and refuse-with-error when the env vars are unset.
- `SKILL.md:80` — example invocation `mongodb://app_user:hunter2@db.example.com:27017/myapp` shows a concrete-looking credential in a URL form; the example normalizes URL-embedded credentials as the canonical input shape and trains the model toward that pattern. Show a non-credentialed example (`mongodb://app_user@db.example.com:27017/myapp` plus "export `MONGO_PASSWORD` first") instead.
- `references/database.md:3` — instructs the model to prompt the user to enter the database password when `MONGO_PASSWORD` is unset; the secret enters the conversation context the moment the user answers. Refuse-and-instruct (exit with a clear error telling the user to export `MONGO_PASSWORD` before re-invoking) instead of prompting.

## Passing checks

- SKILL.md is well within the size budget (80 lines; target ≤500).
- The skill identifies a separate env var (`MONGO_PASSWORD`) as the intended credential channel — the right shape exists, but the workflow undermines it by also accepting URL-embedded passwords and prompting (`SKILL.md:14`).
- Step 3 explicitly states overwrite intent for the generated skill directory (`SKILL.md:20`).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-3/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration. The security findings are the dominant theme — every prompt-for-password, discard-after-read, URL-embedded-credential, and on-disk-credential path needs to collapse into a single refuse-and-instruct flow before this skill is safe to merge.
