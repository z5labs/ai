The following environment variables are required but unset: `PGHOST`, `PGPORT`, `PGUSER`, `PGDATABASE`, `PGPASSWORD`. Export them (directly, or via a credential helper such as `op run --env-file=secrets.env -- claude …`) before re-invoking this skill.

Connection details — including the host and port you mentioned — must reach this skill through environment variables, not as arguments. This keeps secrets out of model context and matches how `psql` itself expects credentials.

For example:

```bash
export PGHOST=db.internal
export PGPORT=5432
export PGUSER=myuser
export PGDATABASE=mydb
export PGPASSWORD=mypassword
# then re-invoke:
/postgres-skill-creator
```

Or use a credential helper to avoid putting the password in your shell history:

```bash
op run --env-file=secrets.env -- claude /postgres-skill-creator
```
