The following environment variables are required but unset: `PGPASSWORD`. Export it (directly, or via a credential helper such as `op run --env-file=secrets.env -- claude …`) before re-invoking this skill.

All connection details must be present before this skill can run — secrets must reach tools out-of-band, never through model context. If the password lands in your context even briefly, it is captured in transcript and any logs the harness keeps.

To proceed, export the missing variable in your shell:

    export PGPASSWORD=your-password-here

Or use a credential helper to inject it without it appearing in your shell history:

    op run --env-file=secrets.env -- claude …
    # or
    vault env -- claude …
    # or via direnv with a .env file loaded automatically

Once `PGPASSWORD` is exported alongside `PGHOST`, `PGPORT`, `PGUSER`, and `PGDATABASE`, re-invoke the skill.
