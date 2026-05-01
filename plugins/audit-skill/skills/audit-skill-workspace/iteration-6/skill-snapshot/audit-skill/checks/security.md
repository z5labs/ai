# Security checks

The principle: **the model should never see secrets**. A credential a model has seen cannot be unseen — it lives in the conversation transcript for the rest of the session and (depending on the harness) in cached prompts beyond. The blast radius is bounded only by what the credential authorizes, which is why this gets its own objective slot rather than living as a sub-check under reproducibility or context management.

Secrets must reach the tools that need them via the runtime environment (env vars, `.env` files loaded by a script, OS keychain) or out-of-band credential helpers (`1Password` CLI, `vault`, `gcloud auth print-access-token`, direnv, etc.). They must not flow through arguments, prompts, or inline content the model reads.

This objective checks for the four ways skills route secrets through the model anyway, plus the one pattern that does it right (so authors have a target shape).

## Checks

Run each grep against the resolved skill directory. The greps are candidate-finders — they hit narrative prose and substring false positives, so triage every match. Raise a finding whenever the skill routes a secret through a model-visible channel (a prompt the model issues, an argument the model reads, a file the model writes with a concrete value), regardless of any later out-of-band move or discard instruction; once the credential has entered the model's context, the leak has happened. Each check below lists its own finding / not-a-finding examples — defer to those for the specific bar.

### 1. Model-prompted secrets

Any instruction telling the model to ask the user for a credential lands the credential in the model's context the moment the user answers. Even if the next step exports it to an env var and the very next sentence says "now forget it", the leak has already happened.

Grep for prompting verbs:

```
grep -rniE '(prompt|ask) [a-z ]*(for|to (provide|enter|supply|give|paste))' SKILL.md references/ 2>/dev/null
```

Then grep for secret-shaped nouns:

```
grep -rniE '(password|passwd|token|secret|credential|api[_ -]?key|access[_ -]?key|private[_ -]?key|bearer|PGPASSWORD|GH_TOKEN|GITHUB_TOKEN|AWS_SECRET|OPENAI_API_KEY|ANTHROPIC_API_KEY)' SKILL.md references/ 2>/dev/null
```

The pattern omits `\b` word boundaries because BSD `grep -E` doesn't support them portably. Both greps are candidate-finders; expect substring matches and ignore them during triage.

A finding is the intersection: an imperative sentence aimed at the model whose object is one of those secret-shaped nouns. Examples that ARE findings:

- "If `PGPASSWORD` is unset, prompt the user for it and export it for the run."
- "Ask the user to paste their GitHub token."
- "If the API key isn't in the environment, the model should request it from the user."

Examples that are NOT findings (because the secret never enters the model's context):

- "If `PGPASSWORD` is unset, refuse to proceed and tell the user to export it before re-invoking."
- "The user must run `op signin` before invoking this skill; the script reads from `op run --env-file`."

Phrase as: `SKILL.md:<line> — instructs the model to prompt for <secret-name>; the secret enters the conversation context the moment the user answers. Refuse-and-instruct (tell the user to export it out-of-band) instead of prompting.`

### 2. Credentials in arguments

If the skill accepts a credential via an argument the model reads — a positional CLI arg, a flag value, an `argument-hint` field, or a declared input — the model has seen it. URL-form connection strings are a special case: protocols like `postgres://`, `mysql://`, `mongodb://`, `https://`, `ssh://`, `redis://`, and `amqp://` allow `user:password@` embedded in the authority component, and a single string of that shape leaks the password and the routing in one go.

Check the YAML frontmatter and any inputs section:

```
grep -niE 'argument-hint:.*(password|passwd|token|secret|credential|api[_ -]?key|access[_ -]?key|private[_ -]?key|connection[_ -]?string|conn[_ -]?str|dsn|url)' SKILL.md
```

Then look in the body of `SKILL.md` and any examples for URL-form credentials:

```
grep -rnE '(postgres(ql)?|mysql|mongodb|redis|amqp|https?|ssh|ftp)://[^/[:space:]]*:[^/[:space:]@]+@' SKILL.md scripts/ references/ 2>/dev/null
```

Findings:

- An `argument-hint` whose name suggests it carries a credential — `[password]`, `[token]`, `[api-key]`, `[connection-string]`, `[dsn]`, `[url]` (when the URL form admits embedded credentials).
- An input declaration typed `password`/`token`/`secret`/`key`/`credential`.
- A workflow step that parses `user:password@host` out of any string argument — the parsing is the giveaway that the password was in the string.
- An example or template showing a URL with concrete-looking credentials (`postgresql://app_user:hunter2@db.example.com/myapp`).

Phrase as: `SKILL.md:<line> — argument "<name>" can carry a credential (<reason: argument-hint type / URL form admits embedded password / input typed as secret>); arguments are visible to the model. Move the credential to an env var read out-of-band by a script.`

For the URL-form case specifically, the suggestion is: split routing from the secret — accept routing components (host, port, user, dbname) via env vars or a non-credentialed config and read the password from a separate env var the model never references.

### 3. Discard-after-read patterns

A skill that says "extract the password from the URL and then discard it" or "read the token, use it, then forget it" has already lost. By the time the discard instruction runs, the credential is in the model's context window — instructing the model to forget it does not remove it from the transcript, the cache, or any logs the harness keeps.

This pattern usually appears in retrofit comments around inputs that already violate check #2: the author noticed the leak and tried to seal it after the fact rather than redesigning the input shape.

Grep for the verbs:

```
grep -rniE '(discard|drop|strip|remove|forget|erase|wipe|clear) [^.]{0,80}(password|passwd|token|secret|credential|api[_ -]?key|access[_ -]?key|private[_ -]?key|PGPASSWORD|GH_TOKEN)' SKILL.md references/ scripts/ 2>/dev/null
```

For each hit, decide whether the surrounding sentence is:

- **Refuse-to-read** ("if the URL embeds a password, refuse it and ask the user to use env-var form") — fine.
- **Discard-after-read** ("after parsing the connection string, discard the password") — finding.

Phrase as: `<file>:<line> — instructs the model to discard <secret-name> after reading it; the secret was already in the context when the discard ran. The fix is to not accept the secret as an argument in the first place (see check #2), not to forget it after.`

### 4. Secrets written to disk by the skill

If a workflow step writes a generated file with a concrete credential baked in (a `.env` with real values, a script with a hardcoded password, a config file with an embedded token), the credential survives the session. This is distinct from check #1: the model didn't necessarily prompt for the secret, but it routed one through itself to the filesystem.

Templates with placeholders are fine — what matters is whether the value at write time is concrete or a placeholder.

Look for write-out patterns near credential-shaped tokens. Heredocs and redirected echoes are the common shapes:

```
grep -rnE '(cat <<-?EOF|cat <<-?[A-Z]+|tee|>) [^|]*\.env([^.]|$)' SKILL.md scripts/ references/ 2>/dev/null
grep -rniE '(write|generate|emit|create) [a-z ]*\.env([^.]|$)' SKILL.md references/ 2>/dev/null
```

Then look for files generated with Write/redirect that contain `KEY=value` lines where `value` looks like a real credential (not `<placeholder>`, `${VAR}`, or `""`):

```
grep -rnE '(PGPASSWORD|GH_TOKEN|API_KEY|SECRET|TOKEN|PASSWORD)=[^<$"[:space:]\\]' SKILL.md scripts/ references/ 2>/dev/null
```

For each hit, check whether the value is:

- A placeholder (`<password>`, `${PGPASSWORD}`, `""`, `your-token-here`) — fine.
- A literal that came from an argument or a prompt the model handled — finding.
- A value the script reads from the runtime environment without showing the model — fine.

Phrase as: `<file>:<line> — generated <output-path> contains a concrete value for <secret-name>; the credential persists on disk after the run. Emit a placeholder (or a `.env.example`) and have the user fill in the real value out-of-band.`

`.env.example` files with commented placeholder values are explicitly the recommended shape (see check #5) — those are not findings.

### 5. Out-of-band sourcing — the passing pattern

This check exists so the author has a target shape, not just a list of failures. Authors tend to undershoot here: when told "don't accept the password as an argument", they reach for the next-most-similar shape (prompting) instead of the right one (env var read by a script).

A skill handles secrets correctly when it does **all** of the following:

1. **Routing components and the secret are separate inputs.** Host, port, user, database name, namespace — those are routing, can be argument or env var, and the model can see them. The password / token / key is a separate input — the model may know its env-var name (so the skill can document the precondition) but never sees its value.
2. **The secret reaches the tool via the runtime environment.** Scripts read `PGPASSWORD`, `GH_TOKEN`, `OPENAI_API_KEY`, etc. directly from the process environment. No CLI flag carries the value.
3. **Population of the env var is the user's job, done out-of-band.** The skill documents the precondition ("`PGPASSWORD` must be exported before invoking this skill") but does not participate in setting it. Suggest credential helpers in passing — `op run --env-file=…` (1Password CLI), `vault read`, `gcloud auth print-access-token`, direnv-loaded `.env` files, `pass`, OS keychain. Do not prescribe one.
4. **Missing-credential failure is a refuse-and-instruct.** If a required env var is unset, the skill exits with a clear error listing each missing var and tells the user to populate it before re-invoking. It does not ask, does not retry, does not fall back to prompting.

Record this as a passing check on the report only when the skill demonstrably does all four. Half-credit (e.g. "secret comes from an env var, but the skill prompts for it when unset") is still a finding under check #1 — the prompting fallback voids the passing pattern.

Phrase a passing-checks entry as: `Security — secrets routed via env vars (<list the vars>); skill refuses with a clear message when any are unset; routing components handled separately (<file>:<line>).`

## What is NOT a finding

- A skill that mentions a secret-shaped noun in narrative prose without instructing the model to handle it ("this skill stores no secrets" — the word `secret` appears, but no instruction follows). Grep is a candidate-finder; require an imperative aimed at the model.
- A `.env.example` template with commented-out placeholder values — that's the recommended shape under check #5, not a violation of check #4.
- Documentation that mentions a credential helper by name (e.g. "users can populate `PGPASSWORD` via `op run --env-file=…`") — naming the helper is fine; prescribing one specific helper as the only supported path is heavier-handed but still not a security finding (it's a strict-definitions concern at most).
- Connection details that genuinely contain no credential — `postgres://app_user@host/db` (user, no password) is fine as an argument; only the `user:password@` shape is the problem.
- A skill that reads a secret from `stdin` via a tool the model invokes but never sees the value (`gh auth login --with-token < token.txt` where `token.txt` is supplied by the user out-of-band) — the model sees the path, not the secret.
- A skill that calls a CLI tool which manages its own credentials independently of the model (`gh`, `gcloud`, `aws`, `kubectl`, `op`, `vault`, `terraform`, `docker login` — anything that reads from its own token store, OS keychain, or shell config). Whether the skill documents an auth precondition for these tools is a strict-definitions concern (declare your inputs and preconditions), not a security one — the model never holds the credential, so there is nothing to leak. The security objective only fires when the model itself is in the credential's path: handed it as an argument, prompted for it, told to discard it, or told to write it to disk.
