# Audit: mongo-explorer

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-3/with_skill/work/skills/mongo-explorer/`
- Date: 2026-04-30
- Findings: 17  (idempotency: 1, reproducibility: 3, context-management: 0, strict-definitions: 2, security: 11)

## Findings

### Idempotency

- `SKILL.md:1` — `SKILL.md` does not declare whether re-running this skill is safe. Line 20 mentions "overwriting any prior version" of the generated directory, but the preamble does not state an overall idempotency stance (does the workflow refuse, overwrite, append, prompt on re-run?). State explicitly that a second invocation overwrites the prior `mongo-<dbname>/` directory and `.env` so the model doesn't have to infer it.

### Reproducibility

- `SKILL.md:33` — the introspection command uses `--eval "..."` as a literal placeholder; no actual queries are specified, so two runs against the same database can produce different schema captures depending on what the model invents. Specify the exact `db.runCommand`/`db.getCollectionNames`/`db.<coll>.findOne` invocations the skill must run.
- `SKILL.md:30` — "Run introspection queries via the `mongo` CLI" depends on `mongo` being installed and on its specific version's behavior, but no precondition is declared. List `mongo` (or `mongosh`) as a required tool in an Inputs/Preconditions section, and pin the expected major version.
- `SKILL.md:44` — "A standard schema-context skill" gives no objective criterion for what "standard" means; two runs can produce different generated SKILL.md shapes. Either name a template file under `references/` and have step 3 copy it, or enumerate the required sections explicitly.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description has no "when to skip" / negative case — likely to over-trigger on near-misses (e.g. when the user already has a `mongo-<dbname>` skill and just wants to refresh schema, or when they want a one-off introspection without generating a skill).
- `SKILL.md:16` — "High-level workflow" lists 5 steps, but the detailed sections below are 5 steps numbered differently: high-level step 3 (write skill) and step 4 (generate `.env`) are merged into detailed Step 3, while detailed Step 4 (smoke-test) does not appear in the high-level list. Renumber so the overview and detail sections match.

### Security

- `SKILL.md:5` — `argument-hint: "[connection-string]"` accepts a credential-bearing URL because `mongodb://user:password@host/db` admits an embedded password; arguments are visible to the model. Split routing (host, port, user, dbname) from the secret: accept the routing components and read `MONGO_PASSWORD` from the runtime environment.
- `SKILL.md:12` — instructs the model to prompt the user for "the full connection string including the password"; the password enters the conversation context the moment the user answers. Refuse-and-instruct (tell the user to export `MONGO_PASSWORD` and supply a credentialless URL) instead of prompting.
- `SKILL.md:12` — "After parsing the connection string, discard the password from the URL since it is no longer needed once we have authenticated" — the password was already in the model's context when this discard ran; discarding does not remove it from the transcript. The fix is to refuse credentialled URLs at input, not to forget the password after.
- `SKILL.md:14` — instructs the model to prompt for the MongoDB password and export it as `MONGO_PASSWORD`; the prompt routes the secret through the model. Refuse-and-instruct instead of prompting.
- `SKILL.md:26` — Step 1 parses `<user>:<password>@<host>:<port>` out of the connection string; the parsing is the giveaway that the password was in the string the model read. Stop accepting credentialled URLs as input.
- `SKILL.md:26` — "discard the password from the in-memory representation of the URL — keep it only in the `MONGO_PASSWORD` variable" — discard-after-read; the secret was already in context when the discard ran.
- `SKILL.md:29` — "ask the user to paste their password. Do not proceed until you have it." — prompts the model for a secret; refuse-and-instruct instead.
- `SKILL.md:36` — "The script reads `$CONNECTION_STRING` so the password travels with the URL into the child process" — the model held the URL (with password) when it built the env for the child process. Pass routing via separate vars and `MONGO_PASSWORD` directly; don't reconstruct the credentialled URL.
- `SKILL.md:48` — generated `./.claude/skills/mongo-<dbname>/.env` contains a concrete value for `MONGO_PASSWORD` (line 54: `MONGO_PASSWORD=<password>` with the instruction to "substitute the real values from the connection string at write time"); the credential persists on disk after the run. Emit a `.env.example` with commented placeholders and have the user populate the real `.env` out-of-band.
- `SKILL.md:67` — generated `scripts/query.sh` hardcodes `MONGO_PASSWORD="${MONGO_PASSWORD:-hunter2}"` as a fallback; a real password baked into a script as the default value persists on disk and may be checked into git. Drop the fallback and have the script `exit 1` with a clear error if `MONGO_PASSWORD` is unset.
- `SKILL.md:80` — example invocation shows a URL with concrete-looking credentials (`mongodb://app_user:hunter2@db.example.com:27017/myapp`); even as documentation, this trains the user on the credentialled-URL shape. Replace with a credentialless URL (`mongodb://app_user@db.example.com:27017/myapp`) plus a note that `MONGO_PASSWORD` must be exported separately.
- `references/database.md:3` — instructs the model to "prompt the user to enter their database password" if `MONGO_PASSWORD` is unset; the prompt routes the secret through the model. Refuse-and-instruct: exit with an error telling the user to export `MONGO_PASSWORD` before re-invoking.

## Passing checks

- `argument-hint` is declared in frontmatter (`SKILL.md:5`), so the calling syntax is discoverable without reading the body.
- `SKILL.md` is well within the 500-line context budget (81 lines total) with detail pushed out to `references/database.md`.
- The description names what the skill produces (introspect a Mongo DB and generate a project-level skill) and a triggering phrase ("Use when the user asks to set up Mongo schema context for a project").

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-6/eval-3/with_skill/work/audit-mongo-explorer-2026-04-30.md`. The skill-creator workflow will treat each finding as feedback for an iteration. The security findings dominate this audit and should be addressed as a redesign of the input shape (split routing from the secret; read `MONGO_PASSWORD` from the environment out-of-band) rather than patched one finding at a time.
