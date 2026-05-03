The following environment variable is required but unset: `PGPASSWORD`.

The `postgres-skill-creator` skill takes no arguments and reads all connection details from the standard libpq environment variables (`PGHOST`, `PGPORT`, `PGUSER`, `PGDATABASE`, `PGPASSWORD`). You've already exported four of them; only `PGPASSWORD` is missing.

Please export it before re-invoking the skill. Importantly, do not paste the password into chat — secrets must reach tools out-of-band so they never land in the model's context (where they'd be captured in the transcript, in any future compaction, and in any logs the harness keeps).

Recommended ways to supply it:

- Export it directly in your shell from a source that doesn't echo it (e.g. `read -s -p 'PGPASSWORD: ' PGPASSWORD && export PGPASSWORD`).
- Use a credential helper you already trust to inject it for the session, for example:
  - `op run --env-file=secrets.env -- claude …` (1Password)
  - `vault` / `gcloud secrets` wrapped in a similar `--env-file` style invocation
  - a direnv-loaded `.env` in the project directory

Once `PGPASSWORD` is set in the same shell as the other four variables, re-run `/postgres-skill-creator` and it will introspect your database and generate `./.claude/skills/pg-<dbname>/`.
