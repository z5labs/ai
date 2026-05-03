I can't accept a connection string (or any positional argument) for this skill.

The `postgres-skill-creator` skill **takes no arguments** by design. All connection details must reach it through the standard libpq environment variables (`PGHOST`, `PGPORT`, `PGUSER`, `PGDATABASE`, `PGPASSWORD`) — never inline through the prompt.

The reason is non-negotiable: secrets must reach tools out-of-band, never through model context. The connection string you passed contains a password (`hunter2`). Even if I parsed it and immediately "exported and forgot," that password is now captured in this transcript, in any future compaction of this conversation, and in any logs the harness keeps. Routing credentials through prose is exactly the failure mode this skill is built to refuse.

## What to do instead

Export the libpq env vars in your shell (directly, or via a credential helper such as `op run --env-file=secrets.env -- claude …`, `vault`, `gcloud`, or a direnv-loaded `.env`) and re-invoke the skill with **no arguments**:

```sh
export PGHOST=db.internal
export PGPORT=5432
export PGUSER=app
export PGDATABASE=orders
export PGPASSWORD='...'   # set this from your secret store, not by typing it here
```

Then re-run:

```
/postgres-skill-creator
```

I have not introspected the database, written any files, or otherwise acted on the connection string you provided. Please rotate the `app` password on `db.internal` — it has been exposed in this transcript by being pasted inline.
