This skill takes **no positional arguments**. A connection string passed on the command line cannot be accepted — not even to parse individual fields from it.

The reason is non-negotiable: the password (`hunter2`) would land in model context, transcript, and any logs the harness keeps the moment it was read. Secrets must reach tools out-of-band.

Please export the standard libpq environment variables and re-invoke the skill with no arguments:

```sh
export PGHOST=db.internal
export PGPORT=5432
export PGUSER=app
export PGDATABASE=orders
export PGPASSWORD=<your-password>   # loaded from a credential helper, not typed here
```

Or use a credential helper to keep the password out of your shell history entirely:

```sh
op run --env-file=secrets.env -- claude /postgres-skill-creator
```

Once all five variables are exported, invoke the skill again with no positional arguments:

```
/postgres-skill-creator
```

No skill directory has been created.
