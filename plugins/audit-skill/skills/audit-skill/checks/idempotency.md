# Idempotency checks

Re-running a skill on the same inputs should converge to the same end state, OR the skill should declare up front that it is intentionally non-idempotent (and why). Both stances are fine; what's not fine is leaving the question unanswered, because the model has to guess at re-run safety every time.

## Skip condition

If `SKILL.md` declares the skill's idempotency stance in its preamble or workflow section — e.g. "running this twice is safe and overwrites the prior output" or "this skill is intentionally non-idempotent because each run appends a new entry" — record that as a passing check and skip the rest of this objective.

The declaration must be specific. "Use carefully" or "be aware re-runs may have effects" doesn't qualify — you can't tell from that whether re-running is safe.

## Checks

Run each grep against the resolved skill directory. For each hit, decide whether the surrounding context already addresses idempotency. Only raise a finding if it doesn't.

### 1. Missing idempotency declaration

If the skip condition didn't fire, raise one finding at `SKILL.md:1`:

> **Idempotency** — `SKILL.md` does not declare whether re-running this skill is safe. Authors should state explicitly whether a second invocation overwrites, appends, or refuses, so the model doesn't have to infer it from the workflow.

### 2. Side-effecting steps without state checks

Grep for verbs/commands that mutate state:

```
grep -nE '\b(rm|mv|cp|mkdir|chmod|chown|git push|git commit|git tag|gh pr create|gh issue create|gh pr review|gh release create|curl -X (POST|PUT|DELETE|PATCH)|wget -O)\b' SKILL.md scripts/ references/ 2>/dev/null
```

For each hit, check whether the surrounding 5–10 lines:
- Test for the resource's existence first (e.g. `[ -f X ]`, `if [ -d X ]`, "if the file already exists, ..."), OR
- Explicitly say "overwrite" / "replace" / "delete and recreate" as a documented intent.

Raise a finding when neither holds. Phrase as: `<file>:<line> — <command> mutates state without a precondition or documented overwrite intent`.

### 3. Generated output paths

If the skill writes generated files (`Write` tool, redirect `>`, `tee`, `cat <<EOF >`), the workflow must say what happens when the destination already exists. Search for write-out patterns and check the surrounding text answers:
- Does the skill overwrite, append, refuse, or prompt?
- Is that choice stated, or implied?

If implied or unstated, raise: `<file>:<line> — output path <X> is written without specifying overwrite/append behavior`.

### 4. Stateful external calls

`gh pr create`, `gh issue create`, `git push`, `git tag`, network POSTs — these are not naturally idempotent. The skill must either:
- Check first (`gh pr list --head $branch` before `gh pr create`),
- Document that duplicates are acceptable / expected,
- Or use an idempotent alternative (`gh pr edit` instead of `create`, etc.).

Raise a finding if a stateful external call appears with no precondition and no doc.

## What is NOT a finding

- Pure read-only commands (`grep`, `find`, `cat`, `git log`, `gh pr view`) — these are inherently idempotent.
- Tempfile creation in `/tmp/` or in a workspace directory clearly scoped to the run.
- Side-effects whose non-idempotency is part of the skill's purpose, when that purpose is stated (e.g. "appends a new audit entry").
