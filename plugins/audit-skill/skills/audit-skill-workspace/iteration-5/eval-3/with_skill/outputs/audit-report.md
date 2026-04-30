# Audit: mongo-explorer

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-5/eval-3/with_skill/work/skills/mongo-explorer/SKILL.md`
- Date: 2026-04-30
- Findings: 14  (idempotency: 0, reproducibility: 0, context-management: 0, strict-definitions: 3, security: 11)

## Findings

### Idempotency

No findings.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description has no "when to skip" / negative case — likely to over-trigger on near-misses (e.g. when the user already has a Mongo schema skill, or wants ad-hoc introspection without generating a skill).
- `SKILL.md:3` — description claims the skill takes "a connection string" but does not surface the material side effects the workflow performs at `SKILL.md:21` and `SKILL.md:38-69`: it writes a `.env` containing credentials and a script with hardcoded credential fallbacks under `./.claude/skills/mongo-<dbname>/`. Either narrow the description or surface those writes.
- `SKILL.md:10-14` — inputs ("connection string", "MongoDB password", `MONGO_PASSWORD` env var) are referenced but no Inputs section states their source, required-ness, or validation rule (e.g. accepted URL shapes, what counts as "no password present").

### Security

- `SKILL.md:12` — instructs the model to prompt the user for "the full connection string including the password"; the password enters the conversation context the moment the user answers. Refuse-and-instruct (tell the user to export `MONGO_PASSWORD` out-of-band and re-invoke) instead of prompting.
- `SKILL.md:14` — instructs the model to "prompt the user for their MongoDB password and export it as `MONGO_PASSWORD`"; the secret is in the model's context before it ever reaches the env var. Refuse-and-instruct instead of prompting.
- `SKILL.md:28` — instructs the model to "ask the user to paste their password" when the connection string lacks one; same leak as above. Refuse-and-instruct instead.
- `SKILL.md:5` — `argument-hint: "[connection-string]"` admits an embedded credential (the `mongodb://user:password@host` URL form is documented as expected at line 10); arguments are visible to the model. Split routing from the secret — accept host/port/user/dbname via env vars or a non-credentialed config and read `MONGO_PASSWORD` from a separate env var the model never references.
- `SKILL.md:10` — documents the expected argument shape as `mongodb://user:password@host:27017/dbname`; the URL-form authority leaks the password via the argument. Document the routing-only form and require the password via env var.
- `SKILL.md:24-26` — workflow step parses `<user>:<password>@<host>:<port>` out of the connection string; the parsing is the giveaway that the password was in the argument the model handled.
- `SKILL.md:80` — example shows a concrete-looking URL credential `mongodb://app_user:hunter2@db.example.com:27017/myapp`; replace with the routing-only form (`mongodb://app_user@db.example.com:27017/myapp`) and reference `MONGO_PASSWORD` separately.
- `SKILL.md:12` — instructs the model to "discard the password from the URL" after parsing; the password was already in the context when the discard ran. Fix is to not accept the secret as an argument in the first place, not to forget it after.
- `SKILL.md:26` — instructs the model to "discard the password from the in-memory representation of the URL"; same pattern — the discard is too late. Redesign the input shape so the model never sees the password.
- `SKILL.md:21,47-58` — generated `.env` at `./.claude/skills/mongo-<dbname>/.env` is written with `MONGO_PASSWORD=<password>` substituting "the real values from the connection string at write time"; the credential persists on disk after the run with values the model handled. Emit a `.env.example` with commented placeholders and have the user populate the real `.env` out-of-band.
- `SKILL.md:62-69` — generated `scripts/query.sh` hardcodes `MONGO_USER="${MONGO_USER:-app_user}"` and `MONGO_PASSWORD="${MONGO_PASSWORD:-hunter2}"` as fallbacks; even with an example credential, baking a literal password into the generated script ships a working credential on disk. Drop the fallback — refuse-and-instruct when the env var is unset.

## Passing checks

- Idempotency declaration is present and specific: `SKILL.md:20` states the generated skill directory is overwritten on re-run.
- SKILL.md is 81 lines — well under the 500-line target.
- Outputs are declared with paths, formats, and overwrite behavior at `SKILL.md:38-69`.

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-5/eval-3/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration.
