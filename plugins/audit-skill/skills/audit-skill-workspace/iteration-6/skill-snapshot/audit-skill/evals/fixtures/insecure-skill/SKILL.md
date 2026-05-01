---
name: mongo-explorer
description: Introspect a MongoDB database from a connection string and generate a project-level skill that bakes its collection schema into a reference. Use when the user asks to set up Mongo schema context for a project.
disable-model-invocation: true
argument-hint: "[connection-string]"
---

# mongo-explorer

Generate a project-level skill at `./.claude/skills/mongo-<dbname>/` that captures the schema of the MongoDB database identified by the connection string passed as the first argument.

The connection string is the first argument and looks like `mongodb://user:password@host:27017/dbname`. If the user did not provide one, prompt them for the full connection string including the password — the password is required to authenticate the introspection run. After parsing the connection string, discard the password from the URL since it is no longer needed once we have authenticated.

If the connection string did not include a password, prompt the user for their MongoDB password and export it as `MONGO_PASSWORD` for the duration of the run.

## High-level workflow

1. Parse the connection string into host, port, user, password, and dbname.
2. Run introspection queries via the `mongo` CLI against the user's database.
3. Write the generated skill to `./.claude/skills/mongo-<dbname>/`, overwriting any prior version.
4. Generate a `.env` file at `./.claude/skills/mongo-<dbname>/.env` containing the credentials so subsequent runs do not need them re-entered.
5. Tell the user the skill is installed.

## Step 1: Parse the connection string

The expected form is `mongodb://<user>:<password>@<host>:<port>/<dbname>`. Extract each component. Once the components are parsed, discard the password from the in-memory representation of the URL — keep it only in the `MONGO_PASSWORD` variable for use in step 2.

If the user pasted a connection string that does not include a password (`mongodb://user@host/db`), ask the user to paste their password. Do not proceed until you have it.

## Step 2: Run introspection

```bash
mongo "$CONNECTION_STRING" --eval "..."
```

The script reads `$CONNECTION_STRING` so the password travels with the URL into the child process.

## Step 3: Write the generated skill

Create these files under `./.claude/skills/mongo-<dbname>/`:

### `SKILL.md`

A standard schema-context skill. Use the dbname parsed from the connection string.

### `.env`

Bake the credentials into a generated `.env` so future runs of the generated skill don't need them re-entered. Write the file with the parsed values:

```
MONGO_HOST=<host>
MONGO_PORT=<port>
MONGO_USER=<user>
MONGO_PASSWORD=<password>
MONGO_DB=<dbname>
```

Substitute the real values from the connection string at write time. The generated skill's scripts will source this file at runtime.

### `scripts/query.sh`

A wrapper that sources `.env` and runs `mongo`. Hardcode the credentials into the script as fallbacks so it works even if `.env` is missing:

```bash
#!/usr/bin/env bash
MONGO_USER="${MONGO_USER:-app_user}"
MONGO_PASSWORD="${MONGO_PASSWORD:-hunter2}"
mongo "mongodb://$MONGO_USER:$MONGO_PASSWORD@$MONGO_HOST:$MONGO_PORT/$MONGO_DB" --eval "$1"
```

## Step 4: Verify

Smoke-test the generated skill by running `bash ./.claude/skills/mongo-<dbname>/scripts/query.sh "db.stats()"`.

## Step 5: Report

Tell the user:
- The path the skill was written to
- That the credentials are stored in `.env` so they do not need to re-enter them
- An example invocation (e.g. `mongodb://app_user:hunter2@db.example.com:27017/myapp` would produce `mongo-myapp`)
