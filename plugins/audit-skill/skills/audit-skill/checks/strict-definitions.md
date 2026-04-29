# Strict-definitions checks

A skill is a function: it should have typed inputs, typed outputs, and a clear contract on each step. This objective absorbs trigger-quality (description, when-to-use, when-to-skip) because triggering is itself part of the contract — the description tells the model when it's the right tool. Most of these checks read like a code review for an API.

## Checks

### 1. Description quality

Read the `description` field in the YAML frontmatter. Check it answers all three:

1. **What it does** (the verb, the artifact produced).
2. **When to use** (concrete user phrasings or contexts that should trigger).
3. **When to skip** (near-misses where another skill or no skill is the better choice).

Raise one finding per missing element:
- `SKILL.md:<frontmatter line> — description doesn't say what the skill produces (no verb / artifact named)`
- `SKILL.md:<frontmatter line> — description has no "when to use" examples — likely to under-trigger`
- `SKILL.md:<frontmatter line> — description has no "when to skip" / negative case — likely to over-trigger on near-misses`

The "when to skip" check is the most-missed and most-valuable. Skills that don't say when NOT to fire tend to fire on adjacent-but-wrong tasks.

### 2. Description ↔ workflow scope consistency

The description sells the skill — and once the skill triggers, the model executes the workflow. If the description's promises and the workflow's delivery disagree, the model has been mis-sold and will surprise the user.

Read the description and the workflow / inputs / outputs sections side by side. For each noun phrase or capability the description names, locate where the rest of the skill delivers it. Discrepancies are findings, and they're some of the most actionable ones — usually a one-word edit fixes the contract.

Look specifically for:

- **Plurality mismatch**: description says "files" (plural), `argument-hint` and inputs are singular (one path). Or vice versa: description says "a file", workflow loops over a glob.
- **Promised feature absent**: description says "with optional dry-run", "supports retries", "auto-detects format" — but no step or input in the workflow corresponds to that capability.
- **Scope broader than implementation**: description says "any text format", workflow only handles UTF-8; description says "for any Postgres database", workflow assumes a specific schema.
- **Unmentioned scope**: workflow has a step that materially affects state (writes a file, makes a network call) that the description doesn't surface. The description should mention the side-effect class so triggering decisions account for it.

Phrase as: `SKILL.md:<description-line> — description claims "<exact phrase>" but <workflow-section/inputs/outputs> at <other-line> does not deliver it; either narrow the description or extend the workflow`.

This is distinct from the "examples drift" check under reproducibility — that one is about an example contradicting the rule it sits under. This check is about the description (the trigger contract) contradicting the workflow (the runtime contract).

### 3. Inputs declared

Scan `SKILL.md` for an Inputs / Arguments / Parameters section, or for an `argument-hint` field in frontmatter. For each input the skill consumes (CLI arg, env var, prompt, file path, stdin), check:

- Name.
- Source (which of those four).
- Required vs optional, with a default if optional.
- Validation rule (regex, allowed values, "must exist", etc.).

Raise one finding per ungrounded input: `SKILL.md:<line> — input "<name>" is referenced but its source / required-ness / validation are not stated`.

### 4. Outputs declared

The workflow should end by stating, for each artifact the skill produces:
- Path (or path pattern).
- Format (file type, schema, template referenced).
- What happens to a pre-existing file at that path (overwrite / append / error — overlaps with idempotency, but call it out here too).

Raise: `SKILL.md:<line> — output "<X>" has no documented path/format`.

### 5. Step ordering and preconditions

For each numbered step or `###` workflow phase, check it states what must be true before the step runs and what it produces. Common gaps:
- Steps that depend on a file written by an earlier step but don't say so.
- Steps presented as a flat list when they're actually a DAG (B and C both depend on A — say so explicitly).
- Optional steps that don't declare their condition ("if there's no PR open, skip step 4" — fine; "step 4 (optional)" with no predicate — finding).

Raise: `SKILL.md:<line> — step <X> doesn't state its preconditions / depends on <Y> implicitly`.

### 6. Vague verbs in instructions

Specifically the model-facing imperative verbs. Grep for:

```
grep -nE '\b(handle|process|deal with|address|take care of|manage|consider) [a-z]' SKILL.md
```

Each hit is a candidate finding — the verb gives the model latitude without saying what to actually do. Keep findings only when there's no concrete follow-up sentence ("handle errors" with nothing else = finding; "handle errors by stopping and reporting the failed file" = fine).

### 7. Missing argument-hint for slash-command-style skills

If the frontmatter has `disable-model-invocation: true` or the skill is named like a command (`/foo`), there should be an `argument-hint` field telling users what to type. Otherwise the user has to read SKILL.md to learn the calling syntax.

Raise: `SKILL.md:<frontmatter line> — slash-style skill has no argument-hint`.

## What is NOT a finding

- A description that's terser than ideal but covers all three elements (what / when / skip) — readability is a separate concern.
- Inputs whose source is obvious from a single canonical CLI form (e.g. `argument-hint: "[connection-string]"` makes the source unambiguous).
- A workflow that's genuinely linear and where preconditions are implied by step number — only flag when there's branching, optional steps, or skipped steps.
