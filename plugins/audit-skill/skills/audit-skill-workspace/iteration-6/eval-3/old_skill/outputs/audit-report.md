# Audit: mongo-explorer

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-3/old_skill/work/skills/mongo-explorer/`
- Date: 2026-04-30
- Findings: 17  (idempotency: 2, reproducibility: 3, context-management: 0, strict-definitions: 4, security: 8)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare a top-level stance on whether re-running the skill is safe; Step 3 mentions overwriting the skill directory but the `.env` write and the rest of the workflow leave re-run behavior unstated. State once at the top whether a second invocation overwrites everything, merges, or refuses.
- `SKILL.md:21` — Step 4 writes `./.claude/skills/mongo-<dbname>/.env` without specifying overwrite/append behavior on re-run; if `.env` already contains a more recent password, this silently clobbers it.

### Reproducibility

- `SKILL.md:33` — Step 2 runs the `mongo` CLI but the skill does not list `mongo` as a required tool / precondition; an environment without it produces a different failure mode than one with it.
- `SKILL.md:33` — Step 2 reads `$CONNECTION_STRING` without declaring how that variable is populated (it is not parsed from `argument-hint`, not mentioned in any inputs section); two runs that populate it differently produce different outputs.
- `SKILL.md:33` — `mongo "$CONNECTION_STRING" --eval "..."` — the `--eval` body is literally `"..."`; the actual introspection queries are unspecified, so two implementers will write two different scripts.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description has no "when to skip" / negative case (e.g. "skip when the user wants ad-hoc queries instead of a persistent schema reference"); likely to over-trigger on adjacent Mongo asks.
- `SKILL.md:3` — description claims the generated skill "bakes its collection schema into a reference" but Step 3 (lines 40–69) lists only `SKILL.md`, `.env`, and `scripts/query.sh` as generated files — no schema reference file is produced. Either narrow the description or add a step that writes the schema artifact.
- `SKILL.md:44` — output `SKILL.md` for the generated skill is described only as "A standard schema-context skill"; path/format/template are not stated, so two runs could produce structurally different outputs.
- `SKILL.md:5` — input "connection-string" is declared via `argument-hint` but required-ness (the workflow says "if the user did not provide one, prompt") and validation (URL shape, required components) are not stated.

### Security

- `SKILL.md:12` — instructs the model to prompt the user for "the full connection string including the password"; the password enters the conversation context the moment the user answers. Refuse-and-instruct (tell the user to export `MONGO_PASSWORD` out-of-band) instead of prompting.
- `SKILL.md:14` — instructs the model to prompt the user for their MongoDB password and export it as `MONGO_PASSWORD`; same leak as line 12. Refuse-and-instruct instead.
- `SKILL.md:28` — `references/database.md`-adjacent instruction "ask the user to paste their password" routes the secret through the model's context. Refuse-and-instruct instead.
- `SKILL.md:5` — `argument-hint: "[connection-string]"` admits a `user:password@host` URL whose password is visible to the model the moment the user invokes the skill. Split routing (host/port/user/dbname as args or env) from the secret (`MONGO_PASSWORD` read out-of-band by a script).
- `SKILL.md:12` — example connection string `mongodb://user:password@host:27017/dbname` shows the URL-form credential shape as the documented input, normalizing the leaking shape. Document a no-password URL plus a separate `MONGO_PASSWORD` env var instead.
- `SKILL.md:12` — instructs the model to "discard the password from the URL since it is no longer needed once we have authenticated"; the password was already in context when the discard ran. The fix is to not accept it as an argument in the first place, not to forget it after.
- `SKILL.md:26` — instructs the model to parse and then "discard the password from the in-memory representation of the URL"; same after-the-fact pattern as line 12.
- `SKILL.md:54` — generated `.env` template contains `MONGO_PASSWORD=<password>` with the instruction "Substitute the real values from the connection string at write time"; the credential is written to disk as a concrete value the model just held. Emit a `.env.example` with placeholder values and tell the user to populate `MONGO_PASSWORD` themselves.
- `SKILL.md:67` — generated `scripts/query.sh` hardcodes `MONGO_PASSWORD="${MONGO_PASSWORD:-hunter2}"` as a fallback; a hardcoded credential ships in every generated skill regardless of the user's real password. Drop the fallback and have the script refuse with a clear error when `MONGO_PASSWORD` is unset.

## Passing checks

- `SKILL.md` is 81 lines — comfortably under the 500-line target with no inline content that needs to move out.
- `references/database.md` is short and focused (4 lines); progressive-disclosure shape is correct even though the file's content duplicates the prompting-for-password leak.
- `argument-hint` is present, satisfying the slash-command-style precondition (separate from the credential concern under security).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-3/old_skill/work/audit-mongo-explorer-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration. The security findings are heavily correlated — fixing the input shape (check #2) likely dissolves checks #1, #3, and #4 in one redesign.
