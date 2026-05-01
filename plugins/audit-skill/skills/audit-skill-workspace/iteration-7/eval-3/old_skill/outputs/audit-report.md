# Audit: mongo-explorer

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-7/eval-3/old_skill/work/skills/mongo-explorer/`
- Date: 2026-04-30
- Findings: 16  (idempotency: 3, reproducibility: 3, context-management: 0, strict-definitions: 3, security: 7)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. State explicitly whether a second invocation overwrites, appends, or refuses.
- `SKILL.md:21` — generated `.env` at `./.claude/skills/mongo-<dbname>/.env` is written without specifying overwrite/append/refuse behavior on re-runs (also a security concern; see Security).
- `SKILL.md:60` — `scripts/query.sh` is written without specifying behavior when the destination already exists; re-runs silently overwrite the prior wrapper.

### Reproducibility

- `SKILL.md:33` — introspection step calls `mongo "$CONNECTION_STRING" --eval "..."` with the eval body elided as `"..."`; two runs cannot produce equivalent outputs because the query body is unspecified.
- `SKILL.md:33` — depends on the `mongo` CLI being installed and on its version-specific behavior without declaring it as a precondition under Inputs.
- `SKILL.md:18` — depends on the user's database state (collections present, schema) without listing the database itself as a declared input shape; same connection string against a mutated DB silently produces a different generated skill.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description has no "when to skip" / negative case (no near-miss like "skip if the project already has a generated mongo-<dbname> skill" or "skip for non-Mongo databases"); likely to over-trigger on adjacent tasks.
- `SKILL.md:5` — `argument-hint: "[connection-string]"` is the only declared input, but the workflow also consumes a separately-prompted `MONGO_PASSWORD` (line 14) and an implicit `mongo` CLI dependency (line 33); list these as declared inputs/preconditions with source and required-ness.
- `SKILL.md:71` — Step 4 "Verify" states a smoke-test command but no precondition (what must exist before it runs) and no failure handling (what to do if `db.stats()` errors); declare the predicate.

### Security

- `SKILL.md:5` — `argument-hint: "[connection-string]"` accepts a URL-form connection string that admits an embedded `user:password@` (check #2). Arguments are visible to the model. Split routing from the secret: accept host/port/user/dbname as routing inputs and read the password from a separate env var the model never sees.
- `SKILL.md:12` — instructs the model to "prompt them for the full connection string including the password" (check #1); the password enters the conversation context the moment the user answers. Refuse-and-instruct: tell the user to export `MONGO_PASSWORD` out-of-band before re-invoking.
- `SKILL.md:12` — instructs the model to "discard the password from the URL" after parsing (check #3); the secret was already in context when the discard ran. Fix is to not accept the URL-form credential at all, not to forget it after.
- `SKILL.md:14` — instructs the model to "prompt the user for their MongoDB password and export it as `MONGO_PASSWORD`" (check #1); the password lands in the transcript as soon as the user answers. Refuse-and-instruct instead.
- `SKILL.md:21` — Step 4 directs the workflow to generate a `.env` file containing the credentials including `MONGO_PASSWORD` at `./.claude/skills/mongo-<dbname>/.env` (check #4); the credential persists on disk after the run, baked in by the model. Emit a `.env.example` with placeholder values and have the user fill in the real value out-of-band.
- `SKILL.md:67` — generated `scripts/query.sh` hardcodes a literal credential fallback `MONGO_PASSWORD="${MONGO_PASSWORD:-hunter2}"` (check #4); a concrete password ships in a generated script. Drop the fallback and refuse-with-error when the env var is unset.
- `references/database.md:3` — instructs the model to "prompt the user to enter their database password" when `MONGO_PASSWORD` is unset (check #1); same issue as `SKILL.md:14`. Refuse-and-instruct: exit with a clear error naming the missing env var and tell the user to export it before re-invoking.

## Passing checks

- SKILL.md is 80 lines — well under the 500-line context-management target.
- Step 3 explicitly states the skill directory is overwritten on re-run (`SKILL.md:20`), giving a partial idempotency declaration for that one artifact (the broader stance is still missing — see Idempotency).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator <path-to-this-file>`. The skill-creator workflow will treat each finding as feedback for an iteration.
