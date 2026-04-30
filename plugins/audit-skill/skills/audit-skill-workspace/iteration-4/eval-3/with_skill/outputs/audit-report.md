# Audit: mongo-explorer

- Target: `/home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-3/with_skill/work/skills/mongo-explorer/`
- Date: 2026-04-30
- Findings: 11  (idempotency: 1, reproducibility: 0, context-management: 0, strict-definitions: 3, security: 7)

## Findings

### Idempotency

- `SKILL.md:1` — no preamble-level idempotency stance. The workflow mentions "overwriting any prior version" for the generated skill directory (`SKILL.md:20`), but the SKILL as a whole does not state whether re-invoking `mongo-explorer` itself is safe (e.g. what happens to a pre-existing `.env` with different values, or to an existing `mongo-<dbname>/` whose contents were edited by hand). Add a one-line stance near the top — "re-running overwrites the generated skill directory and `.env`" or equivalent.

### Reproducibility

No findings.

### Context management

No findings.

### Strict definitions

- `SKILL.md:3` — description says nothing about when NOT to fire. A skill named "mongo-explorer" with the verb "introspect" will over-trigger on adjacent tasks (querying Mongo data, running ad-hoc Mongo commands, generating non-schema docs). Add a "skip when …" clause naming at least one near-miss.
- `SKILL.md:5` — `argument-hint: "[connection-string]"` is the only declared input, but the workflow also consumes a prompted password (`SKILL.md:14, 28`) and an env var `MONGO_PASSWORD` (`SKILL.md:14, 26`). Inputs/Arguments section is missing — declare each input's name, source, required-ness, and validation rule. (See also security findings #2 and #3 — the right fix collapses this finding into "routing components are arguments, password is an env var precondition".)
- `SKILL.md:38–69` — outputs (`./.claude/skills/mongo-<dbname>/SKILL.md`, `./.claude/skills/mongo-<dbname>/.env`, `./.claude/skills/mongo-<dbname>/scripts/query.sh`) are described but their pre-existing-file behavior is only stated for the directory as a whole (`SKILL.md:20`). Spell out per-file overwrite/append/refuse behavior, especially for `.env` (a stale or hand-edited `.env` getting clobbered is a footgun).

### Security

- `SKILL.md:5` — `argument-hint: "[connection-string]"` admits a URL-form value whose authority component carries the password (`mongodb://user:password@host`); the model sees the entire string the moment the user invokes the skill. Split routing from the secret: accept routing components (host, port, user, dbname) via separate arguments or env vars and read the password from `MONGO_PASSWORD` populated out-of-band by the user.
- `SKILL.md:12` — instructs the model to prompt the user for the full connection string including the password; the password enters the conversation context the moment the user answers. Refuse-and-instruct (tell the user to export `MONGO_PASSWORD` out-of-band) instead of prompting.
- `SKILL.md:14` — instructs the model to prompt for the MongoDB password and export it as `MONGO_PASSWORD`; the secret enters the conversation context before the export happens. Refuse with a message telling the user to export `MONGO_PASSWORD` themselves before re-invoking.
- `SKILL.md:28` — instructs the model to "ask the user to paste their password"; same leak as line 14. Remove the prompt and require the env var as a precondition.
- `SKILL.md:12, 26` — discard-after-read pattern: "After parsing the connection string, discard the password from the URL" / "discard the password from the in-memory representation". By the time the discard runs, the password is already in the model's context, the transcript, and any prompt cache. The fix is to not accept the password as part of the connection string in the first place (see the `argument-hint` finding above), not to forget it after.
- `SKILL.md:48–56` — generated `.env` bakes a concrete `MONGO_PASSWORD=<password>` from a value the model handled (parsed out of the connection string or received via prompt). The credential persists on disk after the run and the model wrote it there. Emit a `.env.example` with placeholder/commented values and tell the user to fill in `MONGO_PASSWORD` out-of-band; do not write a real secret from the model's context to disk.
- `SKILL.md:62–69` — generated `scripts/query.sh` hardcodes `MONGO_PASSWORD="${MONGO_PASSWORD:-hunter2}"` as a fallback and reconstructs `mongodb://$MONGO_USER:$MONGO_PASSWORD@…` on the command line. Two problems: a literal credential (`hunter2`) gets baked into a generated script (check #4), and the password rides on the `mongo` argv which is visible to other processes via `/proc/<pid>/cmdline`. Drop the fallback (refuse if `MONGO_PASSWORD` is unset) and pass the password to `mongo` via env / stdin / `--password-file`, not via the URL argv.
- `SKILL.md:80` — example invocation `mongodb://app_user:hunter2@db.example.com:27017/myapp` shows a concrete-looking embedded password in user-facing documentation. Even as an "example", it teaches the URL-with-password shape as the supported call form. Replace with a placeholder form (e.g. `mongodb://app_user@db.example.com:27017/myapp` plus a separate `MONGO_PASSWORD` env var note).

## Passing checks

- Reproducibility — no hedge words ("as appropriate", "carefully", "thoroughly") and no undeclared environment reads (date, whoami, git state, etc.).
- Context management — `SKILL.md` is 80 lines (well under the 500-line target); no oversized inline templates or tables.
- Strict definitions — description names a clear verb (introspect) and artifact (project-level skill capturing schema), and gives a concrete "when to use" phrasing. The frontmatter sets `disable-model-invocation: true` paired with an `argument-hint`, so calling syntax is discoverable (the *content* of that argument is a security finding, but the structural slot is filled).

## Next step

Hand this report back to `skill-creator` to revise: `/skill-creator /home/carson/github.com/z5labs/ai/plugins/audit-skill/skills/audit-skill-workspace/iteration-4/eval-3/with_skill/outputs/audit-report.md`. The skill-creator workflow will treat each finding as feedback for an iteration. The seven security findings collectively point at a single redesign — accept routing components as arguments, require `MONGO_PASSWORD` as an out-of-band env var precondition, refuse-and-instruct when it's unset, and emit a `.env.example` rather than a populated `.env`.
